package service

import (
	"aino_document/models"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

func AddApplicationRole(addAppRole models.AddApplicationRole, userUUID string) error {
	username, errP := GetUsernameByID(userUUID)
	if errP != nil {
		return errP
	}
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()
	uuid := uuid.New()
	uuidString := uuid.String()
	app_role_id := currentTimestamp + int64(uniqueID)

	var roleID int64
	err := db.Get(&roleID, "SELECT role_id FROM role_ms WHERE role_uuid = $1", addAppRole.Role_UUID)
	if err != nil {
		log.Println("Error getting role_id:", err)
		return err
	}
	var applicationID int64
	err = db.Get(&applicationID, "SELECT application_id FROM application_ms WHERE application_uuid = $1", addAppRole.Application_UUID)
	if err != nil {
		log.Println("Error getting application_id:", err)
		return err
	}
	_, errInsert := db.NamedExec("INSERT INTO application_role_ms(application_role_id, application_role_uuid, application_id, role_id, created_by) VALUES(:application_role_id, :application_role_uuid, :application_id, :role_id, :created_by)", map[string]interface{}{
		"application_role_id":   app_role_id,
		"application_role_uuid": uuidString,
		"application_id":        applicationID,
		"role_id":               roleID,
		"created_by":            username,
	})

	if errInsert != nil {
		log.Print(errInsert)
		return errInsert
	}
	return nil
}

func GetAllAppRole() ([]models.ApplicationRole, error) {
	applicationRole := []models.ApplicationRole{}

	rows, err := db.Queryx("SELECT ar.application_role_uuid, a.application_title, r.role_title, ar.application_id, ar.role_id, ar.created_by, ar.created_at, ar.updated_by, ar.updated_at, ar.deleted_by, ar.deleted_at FROM application_role_ms ar JOIN application_ms a ON ar.application_id = a.application_id JOIN role_ms r ON ar.role_id = r.role_id WHERE ar.deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		place := models.ApplicationRole{}
		//rows.StructScan(&place)

		if err := rows.StructScan(&place); err != nil {
			log.Print(err)
			continue
		}
		applicationRole = append(applicationRole, place)
	}
	return applicationRole, nil
}

func GetAppRole(id string) (models.ApplicationRole, error) {
	var appRoleId models.ApplicationRole

	err := db.Get(&appRoleId, "SELECT ar.application_role_uuid, a.application_title, r.role_title, ar.application_id, ar.role_id, ar.created_by, ar.created_at, ar.updated_by, ar.updated_at, ar.deleted_by, ar.deleted_at FROM application_role_ms ar JOIN application_ms a ON ar.application_id = a.application_id JOIN role_ms r ON ar.role_id = r.role_id WHERE ar.application_role_uuid = $1 AND ar.deleted_at IS NULL", id)
	if err != nil {
		return models.ApplicationRole{}, err
	}
	return appRoleId, nil

}

func ListAppRoleById(id string) ([]models.ListAllAppRole, error) {
	//appRoleId := []models.ListAllAppRole{}

	rows, err := db.Queryx("SELECT r.role_uuid, r.role_title FROM application_role_ms ar JOIN application_ms a ON ar.application_id = a.application_id JOIN role_ms r ON ar.role_id = r.role_id WHERE a.application_uuid = $1 AND ar.deleted_at IS NULL", id)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	uniqueRoles := make([]models.ListAllAppRole, 0)

	for rows.Next() {
		var role models.ListAllAppRole
		if err := rows.StructScan(&role); err != nil {
			return nil, err
		}

		if _, ok := seen[role.Role_UUID]; !ok {
			seen[role.Role_UUID] = struct{}{}
			uniqueRoles = append(uniqueRoles, role)
		}
	}

	// Check if no rows were returned
	if len(uniqueRoles) == 0 {
		return nil, sql.ErrNoRows
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return uniqueRoles, nil

}

func UpdateAppRole(updateAppRole models.AddApplicationRole, id string, userUUID string) (models.AddApplicationRole, error) {
	username, errUser := GetUsernameByID(userUUID)
	if errUser != nil {
		log.Print(errUser)
		return models.AddApplicationRole{}, errUser

	}

	var roleID int64
	err := db.Get(&roleID, "SELECT role_id FROM role_ms WHERE role_uuid = $1", updateAppRole.Role_UUID)
	if err != nil {
		log.Println("Error getting role_id:", err)
		return models.AddApplicationRole{}, err
	}
	var applicationID int64
	err = db.Get(&applicationID, "SELECT application_id FROM application_ms WHERE application_uuid = $1", updateAppRole.Application_UUID)
	if err != nil {
		log.Println("Error getting application_id:", err)
		return models.AddApplicationRole{}, err
	}

	currentTime := time.Now()

	_, errUpdate := db.NamedExec("UPDATE application_role_ms SET application_id = :application_id, role_id = :role_id, updated_by = :updated_by, updated_at = :updated_at WHERE application_role_uuid = :id", map[string]interface{}{
		"application_id": applicationID,
		"role_id":        roleID,
		"updated_by":     username,
		"updated_at":     currentTime,
		"id":             id,
	})

	if errUpdate != nil {
		log.Print(errUpdate)
		return models.AddApplicationRole{}, err
	}

	return updateAppRole, nil
}

func DeleteAppRole(id string, userUUID string) error {
	username, errU := GetUsernameByID(userUUID)
	if errU != nil {
		log.Print(errU)
		return errU
	}

	currentTime := time.Now()

	result, err := db.NamedExec("UPDATE application_role_ms SET deleted_by = :deleted_by, deleted_at = :deleted_at WHERE application_role_uuid = :id AND deleted_at IS NULL", map[string]interface{}{
		"deleted_by": username,
		"deleted_at": currentTime,
		"id":         id,
	})
	if err != nil {
		log.Print(err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound // Mengembalikan error jika tidak ada rekaman yang cocok
	}

	return nil
}
