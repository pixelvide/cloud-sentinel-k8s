package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/common"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/utils"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type User struct {
	Model
	Username    string      `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Password    string      `json:"-" gorm:"type:varchar(255)"`
	Name        string      `json:"name,omitempty" gorm:"type:varchar(100);index"`
	AvatarURL   string      `json:"avatar_url,omitempty" gorm:"type:varchar(500)"`
	Provider    string      `json:"provider,omitempty" gorm:"-"`
	OIDCGroups  SliceString `json:"oidc_groups,omitempty" gorm:"-"`
	LastLoginAt *time.Time  `json:"lastLoginAt,omitempty" gorm:"type:timestamp;index"`
	Enabled     bool        `json:"enabled" gorm:"type:boolean;default:true"`
	Sub         string      `json:"sub,omitempty" gorm:"-"`

	Roles             []common.Role `json:"roles,omitempty" gorm:"-"`
	SidebarPreference string        `json:"sidebar_preference,omitempty" gorm:"type:text"`
	Config            *UserConfig   `json:"config,omitempty" gorm:"foreignKey:UserID"`
}

type PersonalAccessToken struct {
	Model
	UserID      uint       `json:"userId" gorm:"not null;index"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	TokenDigest string     `json:"-" gorm:"type:varchar(255);uniqueIndex;not null"`
	Prefix      string     `json:"prefix" gorm:"type:varchar(10);not null"`
	ExpiresAt   *time.Time `json:"expiresAt" gorm:"type:timestamp"`
	LastUsedAt  *time.Time `json:"lastUsedAt" gorm:"type:timestamp"`
	LastUsedIP  string     `json:"lastUsedIP" gorm:"type:text"` // Comma-separated or just the last used IP(s)

	// Relationship
	User User `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type UserIdentity struct {
	Model
	UserID      uint        `json:"user_id" gorm:"index;not null;uniqueIndex:idx_user_provider"`
	Provider    string      `json:"provider" gorm:"type:varchar(50);not null;uniqueIndex:idx_provider_provider_id;uniqueIndex:idx_user_provider"`
	ProviderID  string      `json:"provider_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_provider_provider_id"`
	OIDCGroups  SliceString `json:"oidc_groups,omitempty" gorm:"type:text"`
	LastLoginAt *time.Time  `json:"last_login_at,omitempty" gorm:"type:timestamp"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (u *User) Key() string {
	if u.Username != "" {
		return u.Username
	}
	if u.Name != "" {
		return u.Name
	}
	if u.Sub != "" {
		return u.Sub
	}
	return fmt.Sprintf("%d", u.ID)
}

func AddUser(user *User) error {
	// Hash the password before storing it
	hash, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hash
	return DB.Create(user).Error
}

func CountUsers() (count int64, err error) {
	return count, DB.Model(&User{}).Count(&count).Error
}

func GetUserByID(id uint64) (*User, error) {
	var user User
	if err := DB.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func FindWithSubOrUpsertUser(user *User) error {
	// Capture input values to create identity later if needed
	inputProvider := user.Provider
	if inputProvider == "" {
		return errors.New("user provider is empty")
	}
	inputSub := user.Sub
	if inputSub == "" {
		return errors.New("user sub is empty")
	}
	inputOIDCGroups := user.OIDCGroups

	var identity UserIdentity
	// Try to find identity first
	err := DB.Preload("User").Where("provider = ? AND provider_id = ?", inputProvider, inputSub).First(&identity).Error
	if err == nil {
		klog.Infof("Found existing identity for provider=%s sub=%s user_id=%d", inputProvider, inputSub, identity.UserID)
		// Identity found, update the user object with the one found in DB
		*user = identity.User

		// Restore transient fields from identity
		user.Sub = identity.ProviderID
		user.OIDCGroups = identity.OIDCGroups

		// Update LastLoginAt
		now := time.Now()
		user.LastLoginAt = &now
		identity.LastLoginAt = &now

		// Update OIDC groups on identity if changed
		if fmt.Sprintf("%v", identity.OIDCGroups) != fmt.Sprintf("%v", inputOIDCGroups) {
			identity.OIDCGroups = inputOIDCGroups
			// Update the return object as well
			user.OIDCGroups = inputOIDCGroups
		}

		// Always save identity to update LastLoginAt and optionally OIDC groups
		if err := DB.Save(&identity).Error; err != nil {
			klog.Errorf("Failed to save identity: %v", err)
			return err
		}

		return DB.Save(user).Error
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Errorf("Error searching for identity: %v", err)
		return err
	}

	klog.Infof("Identity not found for provider=%s sub=%s, searching for existing user", inputProvider, inputSub)

	// Identity not found.
	var existingUser User

	// 2. Check if user exists by username (Account Linking)
	if err := DB.Where("username = ?", user.Username).First(&existingUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			klog.Infof("User not found by username=%s, creating new user", user.Username)
			// User does not exist, create new user
			now := time.Now()
			user.LastLoginAt = &now
			if err := DB.Create(user).Error; err != nil {
				klog.Errorf("Failed to create new user: %v", err)
				return err
			}
			existingUser = *user
		} else {
			klog.Errorf("Error searching for user by username: %v", err)
			return err
		}
	} else {
		klog.Infof("Found existing user by username=%s id=%d, linking account", user.Username, existingUser.ID)
	}

	// If we found an existing user (via Sub or Username), we update details
	if existingUser.ID != 0 {
		now := time.Now()
		existingUser.LastLoginAt = &now
		if err := DB.Save(&existingUser).Error; err != nil {
			klog.Errorf("Failed to update existing user: %v", err)
			return err
		}
		*user = existingUser
	}

	// Restore input values to user object (as *user = existingUser might have wiped them)
	user.Sub = inputSub
	user.OIDCGroups = inputOIDCGroups
	user.Provider = inputProvider // Ensure provider is the current one (e.g. if logging in with new provider for existing user)

	klog.Infof("Creating new identity for user_id=%d provider=%s sub=%s", existingUser.ID, inputProvider, inputSub)

	// Create new identity linked to the user
	now := time.Now()
	newIdentity := UserIdentity{
		UserID:      existingUser.ID,
		Provider:    inputProvider,
		ProviderID:  inputSub,
		OIDCGroups:  inputOIDCGroups,
		LastLoginAt: &now,
	}

	if err := DB.Create(&newIdentity).Error; err != nil {
		klog.Errorf("Failed to create new identity: %v", err)
		return err
	}
	return nil
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers returns users with pagination. If limit is 0, defaults to 20.
func ListUsers(limit int, offset int, search string, sortBy string, sortOrder string, role string) (users []User, total int64, err error) {
	if limit <= 0 {
		limit = 20
	}
	// Users are listed normally, PATs are a separate relationship
	query := DB.Model(&User{}).Preload("Config")
	if role != "" {
		query = query.Joins(
			"JOIN role_assignments ra ON ra.subject = users.username AND ra.subject_type = ?",
			SubjectTypeUser,
		).Joins("JOIN roles r ON r.id = ra.role_id").Where("r.name = ?", role)
	}
	if search != "" {
		likeQuery := "%" + search + "%"
		query = query.Where(
			"users.username LIKE ? OR users.name LIKE ?",
			likeQuery,
			likeQuery,
		)
	}
	countQuery := query.Select("users.id").Distinct("users.id")
	err = DB.Table("(?) as sub", countQuery).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	allowedSorts := map[string]string{
		"id":          "users.id",
		"createdAt":   "users.created_at",
		"lastLoginAt": "users.last_login_at",
	}
	sortColumn, ok := allowedSorts[sortBy]
	if !ok {
		sortColumn = "users.id"
	}
	orderExpr := fmt.Sprintf("%s %s", sortColumn, sortOrder)
	if sortColumn == "users.last_login_at" {
		orderExpr = fmt.Sprintf("users.last_login_at IS NULL, users.last_login_at %s", sortOrder)
	}
	idsQuery := query.
		Select("users.id").
		Distinct("users.id").
		Order(orderExpr).
		Limit(limit).
		Offset(offset)
	err = DB.
		Preload("Config").
		Where("id IN (?)", idsQuery).
		Order(orderExpr).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func LoginUser(u *User) error {
	now := time.Now()
	u.LastLoginAt = &now
	return DB.Save(u).Error
}

// DeleteUserByID removes a user by ID
func DeleteUserByID(id uint) error {
	_ = DB.Where("actor_id = ?", id).Delete(&AuditLog{}).Error
	return DB.Delete(&User{}, id).Error
}

// UpdateUser saves provided user (expects ID set)
func UpdateUser(user *User) error {
	return DB.Save(user).Error
}

// ResetPasswordByID sets a new password (hashed) for user with given id
func ResetPasswordByID(id uint, plainPassword string) error {
	var u User
	if err := DB.First(&u, id).Error; err != nil {
		return err
	}
	hash, err := utils.HashPassword(plainPassword)
	if err != nil {
		return err
	}
	u.Password = hash
	return DB.Save(&u).Error
}

// SetUserEnabled sets enabled flag for a user
func SetUserEnabled(id uint, enabled bool) error {
	return DB.Model(&User{}).Where("id = ?", id).Update("enabled", enabled).Error
}

func CheckPassword(hashedPassword, plainPassword string) bool {
	return utils.CheckPasswordHash(plainPassword, hashedPassword)
}

func AddSuperUser(user *User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if err := AddUser(user); err != nil {
		return err
	}
	if err := AddRoleAssignment("admin", SubjectTypeUser, user.Username); err != nil {
		return err
	}
	return nil
}

func NewPersonalAccessToken(userID uint, name string, expiresAt *time.Time) (string, *PersonalAccessToken, error) {
	token := "cspat-" + utils.RandomString(32)
	digest := utils.SHA256Hash(token)
	pat := &PersonalAccessToken{
		UserID:      userID,
		Name:        name,
		TokenDigest: digest,
		Prefix:      token[:10], // cspat- plus first 4 chars
		ExpiresAt:   expiresAt,
	}
	if err := DB.Create(pat).Error; err != nil {
		return "", nil, err
	}
	return token, pat, nil
}

func ListPersonalAccessTokens(userID uint) (tokens []PersonalAccessToken, err error) {
	query := DB.Order("id desc").Preload("User")
	if userID != 0 {
		query = query.Where("user_id = ?", userID)
	}
	err = query.Find(&tokens).Error
	return tokens, err
}

func DeletePersonalAccessToken(id uint, userID uint) error {
	return DB.Where("id = ? AND user_id = ?", id, userID).Delete(&PersonalAccessToken{}).Error
}
