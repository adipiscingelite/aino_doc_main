package service

import (
	"aino_document/models"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

func AddDocument(addDocument models.Document, userUUID string) error {

	username, errP := GetUsernameByID(userUUID)
	if errP != nil {
		return errP
	}
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()

	app_id := currentTimestamp + int64(uniqueID)

	uuid := uuid.New()
	uuidString := uuid.String()
	_, err := db.NamedExec("INSERT INTO document_ms (document_id, document_uuid, document_code, document_name, document_format_number, created_by) VALUES (:document_id, :document_uuid, :document_code, :document_name, :document_format_number, :created_by)", map[string]interface{}{
		"document_id":            app_id,
		"document_uuid":          uuidString,
		"document_code":          addDocument.Code,
		"document_name":          addDocument.Name,
		"document_format_number": addDocument.NumberFormat,
		"created_by":             username,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetAllDoc() ([]models.Document, error) {

	document := []models.Document{}
	rows, errSelect := db.Queryx("select document_uuid, document_order, document_code, document_name, document_format_number, created_by, created_at, updated_by, updated_at from document_ms WHERE deleted_at IS NULL")
	if errSelect != nil {
		return nil, errSelect
	}

	for rows.Next() {
		place := models.Document{}
		rows.StructScan(&place)
		document = append(document, place)
	}

	return document, nil
}
func ShowDocById(id string) (models.Document, error) {
	var document models.Document

	err := db.Get(&document, "SELECT document_uuid, document_order, document_code, document_name,document_format_number, created_by, created_at, updated_by, updated_at FROM document_ms WHERE document_uuid = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return models.Document{}, err
	}
	return document, nil

}

func GetDocCodeName(uuid string) (models.DocCodeName, error) {
	var docCodeName models.DocCodeName

	err := db.Get(&docCodeName, "SELECT document_code, document_name FROM document_ms WHERE document_uuid = $1", uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			// Tidak ada baris yang sesuai
			log.Println("No rows found for role_uuid:", uuid)
			return models.DocCodeName{}, err
		}

		// Terjadi kesalahan lain
		log.Println("Error getting role data by role_ms:", err)
		return models.DocCodeName{}, err
	}

	return docCodeName, nil
}

func IsUniqueDoc(uuid, code, name string) (bool, error) {
	var count int

	var exsitingDocCode, exsitingDocName string
	err := db.Get(&exsitingDocCode, "SELECT document_code FROM document_ms WHERE document_uuid = $1", uuid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	err = db.Get(&exsitingDocName, "SELECT document_name FROM document_ms WHERE document_uuid = $1", uuid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if code == exsitingDocCode && name == exsitingDocName {
		return true, nil
	}

	err = db.Get(&count, "SELECT COUNT(*) FROM document_ms WHERE (document_code = $1 OR document_name = $2) AND document_uuid != $3 AND deleted_at IS NULL", code, name, uuid)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func UpdateDocument(updateDoc models.Document, id string, userUUID string) (models.Document, error) {
	username, errUser := GetUsernameByID(userUUID)
	if errUser != nil {
		log.Print(errUser)
		return models.Document{}, errUser

	}

	currentTime := time.Now()

	_, err := db.NamedExec("UPDATE document_ms SET document_code = :document_code, document_name = :document_name, document_format_number = :document_format_number, updated_by = :updated_by, updated_at = :updated_at WHERE document_uuid = :id", map[string]interface{}{
		"document_code":          updateDoc.Code,
		"document_name":          updateDoc.Name,
		"document_format_number": updateDoc.NumberFormat,
		"updated_by":             username,
		"updated_at":             currentTime,
		"id":                     id,
	})
	if err != nil {
		log.Print(err)
		return models.Document{}, err
	}
	return updateDoc, nil
}

func DeleteDoc(id string, userUUID string) error {
	username, errUser := GetUsernameByID(userUUID)
	if errUser != nil {
		log.Print(errUser)
		return errUser
	}
	currentTime := time.Now()
	result, err := db.Exec("UPDATE document_ms SET deleted_by = :deleted_by, deleted_at = :deleted_at WHERE document_uuid = :id", map[string]interface{}{
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
