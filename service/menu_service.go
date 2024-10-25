package service

import (
	"aino_document/models"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

func AddMenu(addMenu models.Menu, userUUID string) error {
	username, errP := GetUsernameByID(userUUID)
	if errP != nil {
		return errP
	}

	var exsitingMenuId int
	err := db.QueryRow("SELECT menu_id FROM menu_ms WHERE menu_title = $1 AND deleted_at IS NULL", addMenu.Title).Scan(&exsitingMenuId)

	if err == nil {
		// Duplikat ditemukan, kembalikan kesalahan
		log.Print("Role dengan judul atau kode yang sama sudah ada")
	}

	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()

	menuid := currentTimestamp + int64(uniqueID)

	var applicationID int64
	err = db.Get(&applicationID, "SELECT application_id FROM application_ms WHERE application_uuid = $1", addMenu.Application_UUID)
	if err != nil {
		log.Println("Error getting application_id:", err)
		return err
	}
	uuid := uuid.New()
	uuidString := uuid.String()
	permissionString := addMenu.Permission
	if !strings.Contains(permissionString, ",") {
		// Menambahkan koma jika belum ada
		permissionString = strings.Replace(permissionString, " ", ", ", -1)
	}

	_, err = db.NamedExec("INSERT INTO menu_ms(menu_id, menu_uuid, application_id, menu_title, menu_description, required_permission, created_by)VALUES(:menu_id, :menu_uuid, :application_id, :menu_title, :menu_description, :required_permission, :created_by)", map[string]interface{}{
		"menu_id":             menuid,
		"menu_uuid":           uuidString,
		"application_id":      applicationID,
		"menu_title":          addMenu.Title,
		"menu_description":    addMenu.Description,
		"required_permission": permissionString,
		"created_by":          username,
	})

	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func GetAllMenu() ([]models.Menu, error) {
	allMenu := []models.Menu{}

	rows, err := db.Queryx("SELECT menu_uuid, menu_title, required_permission, menu_description, created_by, created_at, updated_by, updated_at, deleted_by, deleted_at FROM menu_ms WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		place := models.Menu{}
		err := rows.StructScan(&place)
		if err != nil {
			return nil, err
		}
		allMenu = append(allMenu, place)

	}
	return allMenu, nil
}

func ShowMenuById(id string) (models.Menu, error) {
	var menuUUID models.Menu

	err := db.Get(&menuUUID, "SELECT  menu_uuid, menu_title, required_permission, menu_description, created_by, created_at, updated_by, updated_at, deleted_by, deleted_at FROM menu_ms WHERE menu_uuid = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return models.Menu{}, err
	}
	return menuUUID, nil
}
