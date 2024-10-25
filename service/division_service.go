package service

import (
	"aino_document/database"
	"aino_document/models"
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB = database.Connection()

func GetAllDivision() ([]models.Divisions, error) {
	divisi := []models.Divisions{}

	rows, errSelect := db.Queryx("select division_uuid, division_order, division_code, division_title, division_show, created_by, created_at, updated_by, updated_at, deleted_by, deleted_at from division_ms WHERE deleted_at IS NULL")
	if errSelect != nil {
		return nil, errSelect
	}

	for rows.Next() {
		place := models.Divisions{}
		rows.StructScan(&place)
		divisi = append(divisi, place)
	}

	return divisi, nil
}

func ShowDivisionById(id string) (models.Divisions, error) {
	var divisiId models.Divisions

	err := db.Get(&divisiId, "SELECT division_uuid, division_order, division_code, division_title, division_show, created_by, created_at, updated_by, updated_at, deleted_by, deleted_at from division_ms WHERE division_uuid = $1 AND deleted_at IS NULL", id)
	if err != nil {
		log.Print(err)
		return models.Divisions{}, err
	}
	return divisiId, nil

}

func GetDivisionById(id int) (models.Division, error) {
	var divisiId models.Division
	idStr := strconv.Itoa(id)

	err := db.Get(&divisiId, "SELECT * FROM division_ms WHERE division_order = $1", idStr)
	if err != nil {
		return models.Division{}, err
	}
	return divisiId, nil

}

func AddDivision(addDivision models.Division, userUUID string) error {
	username, errP := GetUsernameByID(userUUID)
	if errP != nil {
		return errP
	}

	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()

	divisionid := currentTimestamp + int64(uniqueID)

	uuid := uuid.New()
	uuidString := uuid.String()
	_, err := db.NamedExec("INSERT INTO division_ms (division_id, division_uuid, division_code, division_title, created_by) VALUES (:division_id, :division_uuid, :division_code, :division_title, :created_by)", map[string]interface{}{
		"division_id":    divisionid,
		"division_uuid":  uuidString,
		"division_code":  addDivision.Code,
		"division_title": addDivision.Title,
		"created_by":     username,
	})
	if err != nil {
		log.Printf("error add %s", err)
		return err
	}
	return nil
}

func GetDivisionCodeAndTitle(uuid string) (models.DivisionCodeTitle, error) {
	var division models.DivisionCodeTitle

	err := db.Get(&division, "SELECT division_code, division_title FROM division_ms WHERE division_uuid = $1", uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			// Tidak ada baris yang sesuai
			log.Println("No rows found for role_uuid:", uuid)
			return models.DivisionCodeTitle{}, err
		}

		// Terjadi kesalahan lain
		log.Println("Error getting role data by role_ms:", err)
		return models.DivisionCodeTitle{}, err
	}

	return division, nil
}

func IsUniqueDivision(uuid, code, title string) (bool, error) {
	var count int

	var exsitingDivisionCode, exsitingDivisionTitle string
	err := db.Get(&exsitingDivisionCode, "SELECT division_code FROM division_ms WHERE division_uuid = $1", uuid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	err = db.Get(&exsitingDivisionTitle, "SELECT division_title FROM division_ms WHERE division_uuid = $1", uuid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if code == exsitingDivisionCode && title == exsitingDivisionTitle {
		return true, nil
	}

	err = db.Get(&count, "SELECT COUNT(*) FROM division_ms WHERE (division_code = $1 OR division_title = $2) AND division_uuid != $3 AND deleted_at IS NULL", code, title, uuid)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func UpdateDivision(updateDivision models.Division, id string, userUUID string) (models.Division, error) {
	username, errUser := GetUsernameByID(userUUID)
	if errUser != nil {
		log.Print(errUser)
		return models.Division{}, errUser

	}

	currentTime := time.Now()

	_, err := db.NamedExec("UPDATE division_ms SET division_code = :division_code, division_title = :division_title, updated_by = :updated_by, updated_at = :updated_at WHERE division_uuid = :id", map[string]interface{}{
		"division_code":  updateDivision.Code,
		"division_title": updateDivision.Title,
		"updated_by":     username,
		"updated_at":     currentTime,
		"id":             id,
	})
	if err != nil {
		log.Print(err)
		return models.Division{}, err
	}
	return updateDivision, nil
}

func RestoreSoftDeletedDivision(divisionID int, userUUID string, addDivision models.Division) error {
	username, errP := GetUsernameByID(userUUID)
	if errP != nil {
		return errP
	}

	log.Printf("Restoring division with ID: %d", divisionID)
	_, err := db.Exec("UPDATE division_ms SET created_at = NOW(), created_by = $2, updated_at = NULL, updated_by = '', deleted_at = NULL, deleted_by = '', division_code = $3, division_title = $4 WHERE division_id = $1", divisionID, username, addDivision.Code, addDivision.Title)
	if err != nil {
		log.Printf("Error during division restore: %s", err)
		return err
	}

	return nil
}

func DeleteDivision(id string, userUUID string) error {
	username, errUser := GetUsernameByID(userUUID)
	if errUser != nil {
		log.Print(errUser)
		return errUser
	}

	currentTime := time.Now()
	result, err := db.NamedExec("UPDATE division_ms SET deleted_by = :deleted_by, deleted_at = :deleted_at WHERE division_uuid = :id AND deleted_at IS NULL", map[string]interface{}{
		"deleted_by": username,
		"deleted_at": currentTime,
		"id":         id,
	})
	if err != nil {
		log.Print(err)
		return err
	}

	var deletedDivisionID string
	errk := db.Get(&deletedDivisionID, "SELECT division_id FROM division_ms WHERE division_uuid = $1 AND deleted_at IS NOT NULL", id)
	if errk != nil {
		log.Print(errk)
		return err
	}

	_, err = db.Exec("UPDATE user_application_role_ms SET division_id = 0 WHERE division_id = $1 AND deleted_at IS NULL", deletedDivisionID)
	if err != nil {
		log.Print(err)
		return err
	}
	// _, errK := db.Exec("UPDATE user_application_role_ms SET division_id = '' WHERE division_id IN (SELECT division_id FROM division_ms WHERE division_uuid = $1 AND deleted_at IS NULL) AND deleted_at IS NULL", id)
	// if errK != nil {
	// 	log.Printf("bang udah bang %s", errK)
	// 	return err
	// }

	// _, err = db.Exec("UPDATE user_application_role_ms SET division_id = NULL WHERE division_id = $1", divisionID, username)
	// if err != nil {
	// 	log.Printf("Error user app role: %s", err)
	// 	return err
	// }
	// _, err = db.Exec("UPDATE user_application_role_ms SET division_id = NULL WHERE division_id = $1", id)
	// if err != nil {
	// 	log.Printf("Error user app role: %s", err)
	// 	return err
	// }

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound // Mengembalikan error jika tidak ada rekaman yang cocok
	}

	return nil
}
