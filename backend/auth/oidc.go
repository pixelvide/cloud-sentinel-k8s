package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"cloud-sentinel-k8s/db"
	"cloud-sentinel-k8s/models"

	"crypto/tls"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

var (
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
)

func InitOIDC() {
	// Create custom HTTP client to skip TLS verification (needed for Cloudflare Gateway CA)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	customClient := &http.Client{Transport: tr}
	ctx := oidc.ClientContext(context.Background(), customClient)

	provider, err := oidc.NewProvider(ctx, os.Getenv("OIDC_ISSUER"))
	if err != nil {
		log.Printf("Failed to get provider: %v. Retrying in 5 seconds...", err)
		// Retry logic for startup race condition with network
		time.Sleep(5 * time.Second)
		provider, err = oidc.NewProvider(ctx, os.Getenv("OIDC_ISSUER"))
		if err != nil {
			log.Printf("Failed to get provider: %v", err)
			// return // Don't crash
		}
	}

	if provider != nil {
		oidcConfig := &oidc.Config{
			ClientID: os.Getenv("OIDC_CLIENT_ID"),
		}
		verifier = provider.Verifier(oidcConfig)

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}

		oauth2Config = oauth2.Config{
			ClientID:     os.Getenv("OIDC_CLIENT_ID"),
			ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
			RedirectURL:  frontendURL + "/api/v1/auth/callback",
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}
	}
}

func LogoutHandler(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	c.Redirect(http.StatusFound, frontendURL+"/login")
}

func LoginHandler(c *gin.Context) {
	state := generateState()
	c.SetCookie("oauth_state", state, 3600, "/", "", false, true)
	c.Redirect(http.StatusFound, oauth2Config.AuthCodeURL(state))
}

func CallbackHandler(c *gin.Context) {
	state, err := c.Cookie("oauth_state")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state cookie not found"})
		return
	}
	if c.Query("state") != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state mismatch"})
		return
	}

	// Use custom client for token exchange too
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	customClient := &http.Client{Transport: tr}
	ctx := oidc.ClientContext(c.Request.Context(), customClient)

	oauth2Token, err := oauth2Config.Exchange(ctx, c.Query("code"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token: " + err.Error()})
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no id_token field in oauth2 token"})
		return
	}

	idToken, err := verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify ID Token: " + err.Error()})
		return
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse claims: " + err.Error()})
		return
	}

	// Persist User
	user := models.User{
		Email: claims.Email,
		Name:  claims.Name,
		Sub:   claims.Sub,
	}

	// Upsert: On conflict update name, otherwise DoNothing or similar.
	// Simple find or create
	var dbUser models.User
	result := db.DB.Where("sub = ?", claims.Sub).First(&dbUser)
	if result.Error != nil {
		// Create
		if err := db.DB.Create(&user).Error; err != nil {
			log.Printf("Failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
		dbUser = user
	} else {
		// Update (optional, e.g. name change)
		dbUser.Name = claims.Name
		dbUser.Email = claims.Email
		db.DB.Save(&dbUser)
	}

	// Generate internal JWT
	internalToken, err := GenerateToken(dbUser.ID, dbUser.Email, dbUser.Name)
	if err != nil {
		log.Printf("Failed to generate internal token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Set Auth Cookie with internal JWT
	c.SetCookie("auth_token", internalToken, 3600*24, "/", "", false, true) // 1 day

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	// Redirect to Frontend
	c.Redirect(http.StatusFound, frontendURL)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("auth_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Validate our internal JWT
		claims, err := ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Fetch user from DB using the user ID from claims
		var user models.User
		if err := db.DB.First(&user, claims.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
