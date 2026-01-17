package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud-sentinel-k8s/api"
	"cloud-sentinel-k8s/pkg/models"

	"crypto/tls"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
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

		base := os.Getenv("CLOUD_SENTINEL_K8S_BASE")
		if base != "" && !strings.HasPrefix(base, "/") {
			base = "/" + base
		}
		base = strings.TrimSuffix(base, "/")

		oauth2Config = oauth2.Config{
			ClientID:     os.Getenv("OIDC_CLIENT_ID"),
			ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
			RedirectURL:  frontendURL + base + "/api/v1/auth/callback",
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

	base := os.Getenv("CLOUD_SENTINEL_K8S_BASE")
	if base != "" && !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	base = strings.TrimSuffix(base, "/")

	c.Redirect(http.StatusFound, frontendURL+base+"/login")
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

	// User Persistence with Multi-Auth Support
	var dbUser models.User
	var identity models.UserIdentity

	// 1. Check if this specific identity exists
	err = models.DB.Where("provider = ? AND provider_id = ?", "oidc", claims.Sub).First(&identity).Error
	if err == nil {
		// Identity exists, get the user
		if err := models.DB.First(&dbUser, identity.UserID).Error; err != nil {
			log.Printf("Failed to find user for identity: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find associated user"})
			return
		}
	} else {
		// Identity doesn't exist, check if user with same email exists
		err = models.DB.Where("email = ?", claims.Email).First(&dbUser).Error
		if err != nil {
			// No user with this email, create new user
			dbUser = models.User{
				Email: claims.Email,
				Name:  claims.Name,
			}
			if err := models.DB.Create(&dbUser).Error; err != nil {
				log.Printf("Failed to create user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
				return
			}
		}

		// Create the new identity for the user (new or existing)
		identity = models.UserIdentity{
			UserID:     dbUser.ID,
			Provider:   "oidc",
			ProviderID: claims.Sub,
		}
		if err := models.DB.Create(&identity).Error; err != nil {
			log.Printf("Failed to create user identity: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link identity"})
			return
		}
	}

	// Update user info if needed
	if dbUser.Name != claims.Name {
		dbUser.Name = claims.Name
		models.DB.Save(&dbUser)
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

	base := os.Getenv("CLOUD_SENTINEL_K8S_BASE")
	if base != "" && !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	base = strings.TrimSuffix(base, "/")

	// Record Audit Log for login
	api.RecordAuditLogForUser(c, "USER_LOGIN", dbUser.Email, gin.H{
		"email": dbUser.Email,
		"name":  dbUser.Name,
	})

	// Auto-grant access to apps with DefaultUserAccess=true
	var app models.App
	if err := models.DB.Where("name = ?", "cloud-sentinel-k8s").First(&app).Error; err == nil {
		// App exists, check if it has default user access enabled
		if app.Enabled && app.DefaultUserAccess {
			// Check if user already has an access record
			var appUser models.AppUser
			err := models.DB.Where("user_id = ? AND app_id = ?", dbUser.ID, app.ID).First(&appUser).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// Create new AppUser record with access enabled
					appUser = models.AppUser{
						UserID:  dbUser.ID,
						AppID:   app.ID,
						Enabled: true,
					}
					if createErr := models.DB.Create(&appUser).Error; createErr != nil {
						log.Printf("Failed to create AppUser record for user %d and app %d: %v", dbUser.ID, app.ID, createErr)
					} else {
						log.Printf("Auto-granted access to app '%s' for user %s", app.Name, dbUser.Email)
					}
				} else {
					log.Printf("Error checking AppUser record: %v", err)
				}
			}
			// If record exists, no action needed (respect existing enabled/disabled state)
		}
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("Error fetching cloud-sentinel-k8s app: %v", err)
	}

	// Redirect to Frontend
	c.Redirect(http.StatusFound, frontendURL+base)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("auth_token")
		// Test Bypass
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
		if err := models.DB.First(&user, claims.UserID).Error; err != nil {
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
