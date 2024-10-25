package models

import (
	"database/sql"
	"time"
)

type Application struct {
	UUID        string         `json:"application_uuid" db:"application_uuid"`
	Id          int            `json:"application_id" db:"application_id"`
	Code        string         `json:"application_code" db:"application_code" validate:"required"`
	Title       string         `json:"application_title" db:"application_title" validate:"required"`
	Description string         `json:"application_description" db:"application_description"`
	Order       int            `json:"application_order" db:"application_order"`
	Show        bool           `json:"application_show" db:"application_show"`
	Created_by  string         `json:"created_by" db:"created_by"`
	Created_at  time.Time      `json:"created_at" db:"created_at"`
	Updated_by  sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at  sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by  sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at  sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type Applications struct {
	UUID        string         `json:"application_uuid" db:"application_uuid"`
	Order       int            `json:"application_order" db:"application_order"`
	Code        string         `json:"application_code" db:"application_code" validate:"required"`
	Title       string         `json:"application_title" db:"application_title" validate:"required"`
	Description string         `json:"application_description" db:"application_description"`
	Show        bool           `json:"application_show" db:"application_show"`
	Created_by  string         `json:"created_by" db:"created_by"`
	Created_at  time.Time      `json:"created_at" db:"created_at"`
	Updated_by  sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at  sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by  sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at  sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type ApplicationCodeTitle struct {
	UUID  string `json:"application_uuid" db:"application_uuid"`
	Code  string `json:"application_code" db:"application_code"`
	Title string `json:"application_title" db:"application_title"`
}
