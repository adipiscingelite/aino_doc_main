package models

import (
	"database/sql"
	"time"
)

type Division struct {
	Id         int            `json:"division_id" db:"division_id"`
	UUID       string         `json:"division_uuid" db:"division_uuid"`
	Code       string         `json:"division_code" db:"division_code" validate:"required"`
	Title      string         `json:"division_title" db:"division_title" validate:"required"`
	Order      int            `json:"division_order" db:"division_order"`
	Show       bool           `json:"division_show" db:"division_show"`
	Created_by string         `json:"created_by" db:"created_by"`
	Created_at time.Time      `json:"created_at" db:"created_at"`
	Updated_by sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type Divisions struct {
	UUID       string         `json:"division_uuid" db:"division_uuid"`
	Order      int            `json:"division_order" db:"division_order"`
	Code       string         `json:"division_code" db:"division_code" validate:"required"`
	Title      string         `json:"division_title" db:"division_title" validate:"required"`
	Show       bool           `json:"division_show" db:"division_show"`
	Created_by string         `json:"created_by" db:"created_by"`
	Created_at time.Time      `json:"created_at" db:"created_at"`
	Updated_by sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type DivisionCodeTitle struct {
	UUID  string `json:"division_uuid" db:"division_uuid"`
	Code  string `json:"division_code" db:"division_code"`
	Title string `json:"division_title" db:"division_title"`
}
