package models

import (
	"time"
)

type User struct {
	// Equivalent to: id: { primary: true, generated: true, update: false }
	ID uint `gorm:"primaryKey;autoIncrement;column:id;<-:create" json:"id"`

	// Equivalent to: first_name: { type: "varchar", nullable: false }
	FirstName string `gorm:"column:first_name;type:varchar;not null" json:"first_name"`

	// Equivalent to: last_name: { type: "varchar", nullable: false }
	LastName string `gorm:"column:last_name;type:varchar;not null" json:"last_name"`

	// Equivalent to: password: { type: "varchar", nullable: false }
	// JSON tag is "-" so the password is never returned in API responses
	Password string `gorm:"column:password;type:varchar;not null" json:"-"`

	// Equivalent to: username: { type: "varchar", nullable: false, unique: true }
	Username string `gorm:"column:username;type:varchar;not null;unique" json:"username"`

	// Equivalent to: account_created: { default: NOW(), update: false }
	AccountCreated time.Time `gorm:"column:account_created;type:timestamptz;default:CURRENT_TIMESTAMP;<-:create" json:"account_created"`

	// Equivalent to: account_updated: { default: NOW(), update: false, nullable: true }
	// Note: You specified update: false in Node, so I kept <-:create here.
	AccountUpdated time.Time `gorm:"column:account_updated;type:timestamptz;default:CURRENT_TIMESTAMP;<-:create" json:"account_updated"`
}

// TableName ensures the table is named "users"
func (User) TableName() string {
	return "users"
}