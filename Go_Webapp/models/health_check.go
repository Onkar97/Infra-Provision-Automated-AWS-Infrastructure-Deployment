package models

import (
	"time"
)

type HealthCheck struct {
	// Equivalent to: check_id: { primary: true, type: "int", generated: true }
	CheckID uint `gorm:"primaryKey;autoIncrement;column:check_id" json:"check_id"`

	// Equivalent to: check_datetime: { type: "timestamptz", nullable: false, default: NOW(), index: ... }
	CheckDatetime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;type:timestamptz;index:IDX_health_checks_check_datetime;column:check_datetime" json:"check_datetime"`
}

// TableName ensures the table is named "health_checks" exactly as in your Node code
func (HealthCheck) TableName() string {
	return "health_checks"
}