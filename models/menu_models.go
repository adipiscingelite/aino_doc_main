package models

import (
	"database/sql"
	"time"
)

type Menu struct {
	UUID             string         `json:"menu_uuid" db:"menu_uuid"`
	Title            string         `json:"menu_title" db:"menu_title" validate:"required"`
	Description      string         `json:"menu_description" db:"menu_description"`
	Permission       string         `json:"required_permission" db:"required_permission" validate:"required"`
	Application_UUID string         `json:"application_uuid" db:"application_uuid" validate:"required"`
	Created_by       string         `json:"created_by" db:"created_by"`
	Created_at       time.Time      `json:"created_at" db:"created_at"`
	Updated_by       sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at       sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by       sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at       sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}
