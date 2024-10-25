package models

import (
	"database/sql"
	"time"
)

type ApplicationRole struct {
	Application_role_id int            `json:"-" db:"application_role_id"`
	UUID                string         `json:"application_role_uuid" db:"application_role_uuid"`
	Application_ID      string         `json:"-" db:"application_id" validate:"required"`
	Role_ID             string         `json:"-" db:"role_id" validate:"required"`
	Application_title   string         `json:"application_title" db:"application_title"`
	Role_title          string         `json:"role_title" db:"role_title"`
	Created_by          string         `json:"created_by" db:"created_by"`
	Created_at          time.Time      `json:"created_at" db:"created_at"`
	Updated_by          sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at          sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by          sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at          sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type AddApplicationRole struct {
	Application_role_id int            `json:"application_role_id" db:"application_role_id"`
	Application_UUID    string         `json:"application_uuid" db:"application_uuid" validate:"required"`
	Role_UUID           string         `json:"role_uuid" db:"role_uuid" validate:"required"`
	Created_by          string         `json:"created_by" db:"created_by"`
	Created_at          time.Time      `json:"created_at" db:"created_at"`
	Updated_by          sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at          sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by          sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at          sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type AllAppRole struct {
	Application_UUID  string         `json:"application_role_uuid" db:"application_role_uuid"`
	Application_title string         `json:"application_title" db:"application_title"`
	Role_title        string         `json:"role_title" db:"role_title"`
	Created_by        string         `json:"created_by" db:"created_by"`
	Created_at        time.Time      `json:"created_at" db:"created_at"`
	Updated_by        sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at        sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by        sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at        sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type ListAllAppRole struct {
	Role_UUID  string `json:"role_uuid" db:"role_uuid"`
	Role_title string `json:"role_title" db:"role_title"`
}
