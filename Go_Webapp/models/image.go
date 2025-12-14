package models

import (
	"time"
)

type Image struct {
	// Equivalent to: image_id: { primary: true, generated: true, update: false }
	ImageID uint `gorm:"primaryKey;autoIncrement;column:image_id;<-:create" json:"image_id"`

	// Equivalent to: product_id: { type: "int", update: false }
	// Note: We use uint assuming product_id is positive. 
	// If you later define a relationship, you might add a 'Product' struct field here.
	ProductID uint `gorm:"column:product_id;not null;<-:create" json:"product_id"`

	// Equivalent to: file_name: { type: "varchar", update: false }
	FileName string `gorm:"column:file_name;type:varchar;not null;<-:create" json:"file_name"`

	// Equivalent to: date_created: { default: NOW(), update: false }
	DateCreated time.Time `gorm:"column:date_created;type:timestamptz;default:CURRENT_TIMESTAMP;<-:create" json:"date_created"`

	// Equivalent to: s3_bucket_path: { type: "varchar", update: false }
	S3BucketPath string `gorm:"column:s3_bucket_path;type:varchar;not null;<-:create" json:"s3_bucket_path"`
}

// TableName ensures the table is named "image"
func (Image) TableName() string {
	return "image"
}