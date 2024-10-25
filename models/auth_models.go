package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Login struct {
	Username         string `json:"user_name" db:"user_name"`
	Password         string `json:"user_password" db:"user_password" validate:"required"`
	Application_Role int    `json:"application_role_id" db:"application_role_id"`
	Division         int    `json:"division_id" db:"division_id"`
	Email            string `json:"user_email" db:"user_email" validate:"required"`
}

type Register struct {
	Id               int            `json:"user_id" db:"user_id"`
	UUID             uuid.UUID      `json:"user_uuid" db:"user_uuid"`
	Username         string         `json:"user_name" db:"user_name" validate:"required"`
	Password         string         `json:"user_password" db:"user_password" validate:"required"`
	Email            string         `json:"user_email" db:"user_email" validate:"required,email"`
	PersonalUUID     string         `json:"personal_uuid" db:"personal_uuid"`
	PersonalName     string         `json:"personal_name" db:"personal_name" validate:"required"`
	PersonalBirthday string         `json:"personal_birthday" db:"personal_birthday" validate:"required"`
	PersonalGender   string         `json:"personal_gender" db:"personal_gender" validate:"required"`
	PersonalPhone    string         `json:"personal_phone" db:"personal_phone" validate:"required"`
	PersonalAddress  string         `json:"personal_address" db:"personal_address" validate:"required"`
	Created_by       string         `json:"created_by" db:"created_by"`
	Created_at       time.Time      `json:"created_at" db:"created_at"`
	Updated_by       sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at       sql.NullTime   `json:"updated_at" db:"updated_by"`
	Deleted_by       sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at       sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	ApplicationRole  struct {
		Application_UUID string `json:"application_uuid" db:"application_uuid" validate:"required"`
		Role_UUID        string `json:"role_uuid" db:"role_uuid" validate:"required"`
		Division_UUID    string `json:"division_uuid" db:"division_uuid" validate:"required"`
	} `json:"applicationRole"`
}

type UserAppRole struct {
	User_id             int    `json:"user_id" db:"user_id"`
	Application_role_id int    `json:"application_role_id" db:"application_role_id"`
	Division_id         int    `json:"division_id" db:"division_id"`
	Division_code       string `json:"division_code" db:"division_code"`
	Application_id      int    `json:"application_Id" db:"application_id"`
}

type UpdateUser struct {
	Username         string         `json:"user_name" db:"user_name" validate:"required"`
	Email            string         `json:"user_email" db:"user_email" validate:"required,email"`
	PersonalName     string         `json:"personal_name" db:"personal_name" validate:"required"`
	PersonalBirthday string         `json:"personal_birthday" db:"personal_birthday" validate:"required"`
	PersonalGender   string         `json:"personal_gender" db:"personal_gender" validate:"required"`
	PersonalPhone    string         `json:"personal_phone" db:"personal_phone" validate:"required"`
	PersonalAddress  string         `json:"personal_address" db:"personal_address" validate:"required"`
	Created_by       string         `json:"created_by" db:"created_by"`
	Created_at       time.Time      `json:"created_at" db:"created_at"`
	Updated_by       sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at       sql.NullTime   `json:"updated_at" db:"updated_by"`
	Deleted_by       sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at       sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	ApplicationRole  struct {
		Application_role_UUID string `json:"application_role_uuid" db:"application_role_uuid"`
		Application_UUID      string `json:"application_uuid" db:"application_uuid" validate:"required"`
		Role_UUID             string `json:"role_uuid" db:"role_uuid" validate:"required"`
		Division_UUID         string `json:"division_uuid" db:"division_uuid" validate:"required"`
	} `json:"applicationRole"`
}
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}
