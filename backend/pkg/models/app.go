package models

type App struct {
	Model
	Name              string `gorm:"uniqueIndex;not null" json:"name"`
	Enabled           bool   `gorm:"default:true" json:"enabled"`
	DefaultUserAccess bool   `gorm:"default:true" json:"default_user_access"`
}
