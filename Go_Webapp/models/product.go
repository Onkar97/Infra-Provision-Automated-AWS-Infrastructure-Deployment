package models

import (
	"time"
)

type Product struct {
	// Equivalent to: id: { primary: true, generated: true, update: false }
	ID uint `gorm:"primaryKey;autoIncrement;column:id;<-:create" json:"id"`

	// Equivalent to: name: { type: "varchar", nullable: false }
	Name string `gorm:"column:name;type:varchar;not null" json:"name"`

	// Equivalent to: description: { type: "varchar", nullable: false }
	Description string `gorm:"column:description;type:varchar;not null" json:"description"`

	// Equivalent to: sku: { type: "varchar", nullable: false }
	Sku string `gorm:"column:sku;type:varchar;not null" json:"sku"`

	// Equivalent to: manufacturer: { type: "varchar", nullable: false }
	Manufacturer string `gorm:"column:manufacturer;type:varchar;not null" json:"manufacturer"`

	// Equivalent to: quantity: { type: "int", nullable: false, minimum: 0 }
	// Plus check constraint: "quantity" >= 0 AND "quantity" <= 100
	Quantity int `gorm:"column:quantity;type:int;not null;check:quantity >= 0 AND quantity <= 100" json:"quantity"`

	// Equivalent to: date_added: { default: NOW(), update: false }
	DateAdded time.Time `gorm:"column:date_added;type:timestamptz;default:CURRENT_TIMESTAMP;<-:create" json:"date_added"`

	// Equivalent to: date_last_updated: { default: NOW(), update: false, nullable: true }
	// Note: I kept your specific requirement (update: false) using <-:create
	DateLastUpdated time.Time `gorm:"column:date_last_updated;type:timestamptz;default:CURRENT_TIMESTAMP;<-:create" json:"date_last_updated"`

	// Equivalent to: owner_user_id: { type: "int", update: false }
	OwnerUserID uint `gorm:"column:owner_user_id;not null;<-:create" json:"owner_user_id"`
}

// TableName ensures the table is named "product"
func (Product) TableName() string {
	return "product"
}