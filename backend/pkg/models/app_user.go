package models

type AppUser struct {
	Model
	UserID  uint `gorm:"not null;uniqueIndex:idx_user_app" json:"user_id"`
	AppID   uint `gorm:"not null;uniqueIndex:idx_user_app" json:"app_id"`
	Enabled bool `gorm:"default:true" json:"enabled"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	App  App  `gorm:"foreignKey:AppID" json:"app,omitempty"`
}
