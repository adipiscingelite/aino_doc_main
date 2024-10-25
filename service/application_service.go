package service

import (
	"aino_document/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	jose "github.com/dvsekhvalnov/jose2go"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type JwtCustomClaims struct {
	UserUUID string `json:"user_uuid"`
	// AppRoleId          int `json:"application_role_id"`
	// DivisionId         int `json:"division_id"`
	jwt.StandardClaims // Embed the StandardClaims struct

}

func DecryptJWE(jweToken string, secretKey string) (string, error) {
	// Dekripsi token JWE
	decrypted, _, err := jose.Decode(jweToken, secretKey)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}

func GetUsernameByID(userUUID string) (string, error) {
	var username string
	err := db.QueryRow("SELECT user_name from user_ms WHERE user_uuid = $1", userUUID).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func GetUserInfoFromToken(tokenStr string) (string, error) {
	secretKey := "secretJwToken" // Ganti dengan kunci yang benar

	decrypted, err := DecryptJWE(tokenStr, secretKey)
	if err != nil {
		fmt.Println("Gagal mendekripsi token:", err)
		return "", err
	}

	fmt.Println("Token yang sudah dideskripsi:", decrypted)

	var claims JwtCustomClaims
	errJ := json.Unmarshal([]byte(decrypted), &claims)
	if errJ != nil {
		fmt.Println("Gagal mengurai klaim:", errJ)
		return "", errJ
	}

	userUUID := claims.UserUUID // Mengakses UserID langsung
	return userUUID, nil
}

func AddApplication(addApplication models.Application, userUUID string) error {
	username, errP := GetUsernameByID(userUUID)
	if errP != nil {
		return errP
	}

	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()

	uuid := uuid.New()
	uuidString := uuid.String()

	app_id := currentTimestamp + int64(uniqueID)

	_, err := db.NamedExec("INSERT INTO application_ms (application_id, application_uuid, application_code, application_title, application_description, created_by) VALUES (:application_id, :application_uuid, :application_code, :application_title, :application_description, :created_by)", map[string]interface{}{
		"application_id":          app_id,
		"application_uuid":        uuidString,
		"application_code":        addApplication.Code,
		"application_title":       addApplication.Title,
		"application_description": addApplication.Description,
		"created_by":              username,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetAllApp() ([]models.Applications, error) {
	application := []models.Applications{}

	rows, err := db.Queryx("SELECT application_uuid, application_order, application_code, application_title, application_description, application_show, created_by, created_at, updated_by, updated_at, deleted_by, deleted_at FROM application_ms WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		place := models.Applications{}
		rows.StructScan(&place)
		application = append(application, place)
	}
	return application, nil
}

func ShowApplicationById(id string) (models.Applications, error) {
	var appid models.Applications

	err := db.Get(&appid, "SELECT application_uuid, application_order, application_code, application_title, application_description, application_show, created_by, created_at, updated_by, updated_at, deleted_by, deleted_at FROM application_ms WHERE application_uuid = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return models.Applications{}, err
	}
	return appid, nil

}

func GetAppById(id int) (models.Application, error) {
	var appid models.Application
	idStr := strconv.Itoa(id)

	err := db.Get(&appid, "SELECT * FROM application_ms WHERE application_order = $1", idStr)
	if err != nil {
		return models.Application{}, err
	}
	return appid, nil

}
func GetAppCodeAndTitle(uuid string) (models.ApplicationCodeTitle, error) {
	var app models.ApplicationCodeTitle

	err := db.Get(&app, "SELECT application_code, application_title FROM application_ms WHERE application_uuid = $1", uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			// Tidak ada baris yang sesuai
			log.Println("No rows found for app_uuid:", uuid)
			return models.ApplicationCodeTitle{}, err
		}

		// Terjadi kesalahan lain
		log.Println("Error getting app data by application_ms:", err)
		return models.ApplicationCodeTitle{}, err
	}

	return app, nil
}

func IsUniqueApp(uuid, code, title string) (bool, error) {
	var count int

	var exsitingAppCode, exsitingAppTitle string
	err := db.Get(&exsitingAppCode, "SELECT application_code FROM application_ms WHERE application_uuid = $1", uuid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	err = db.Get(&exsitingAppTitle, "SELECT application_title FROM application_ms WHERE application_uuid = $1", uuid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if code == exsitingAppCode && title == exsitingAppTitle {
		return true, nil
	}

	err = db.Get(&count, "SELECT COUNT(*) FROM application_ms WHERE (application_code = $1 OR application_title = $2) AND application_uuid != $3 AND deleted_at IS NULL", code, title, uuid)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func UpdateApp(updateApp models.Application, id string, userUUID string) (models.Application, error) {
	username, errUser := GetUsernameByID(userUUID)
	if errUser != nil {
		log.Print(errUser)
		return models.Application{}, errUser

	}

	currentTime := time.Now()

	_, err := db.NamedExec("UPDATE application_ms SET application_code = :application_code, application_title = :application_title, application_description = :application_description, updated_by = :updated_by, updated_at = :updated_at WHERE application_uuid = :id", map[string]interface{}{
		"application_code":        updateApp.Code,
		"application_title":       updateApp.Title,
		"application_description": updateApp.Description,
		"updated_by":              username,
		"updated_at":              currentTime,
		"id":                      id,
	})

	if err != nil {
		log.Print(err)
		return models.Application{}, err
	}

	return updateApp, nil

}

func RestoreSoftDeletedApp(appID int, userUUID string, addApplication models.Application) error {
	username, _ := GetUsernameByID(userUUID)
	log.Printf("Restoring app with ID: %d", appID)
	_, err := db.Exec("UPDATE application_ms SET created_at = NOW(), created_by = $2, updated_at = NULL, updated_by = '',  deleted_at = NULL, deleted_by = '', application_code = $3, application_title = $4 WHERE application_id = $1", appID, username, addApplication.Code, addApplication.Title)
	if err != nil {
		log.Printf("Error during application restore: %s", err)
		return err
	}

	return nil
}

func DeleteApp(id string, userUUID string) error {
	username, errU := GetUsernameByID(userUUID)
	if errU != nil {
		log.Print(errU)
		return errU
	}

	currentTime := time.Now()

	result, err := db.NamedExec("UPDATE application_ms SET deleted_by = :deleted_by, deleted_at = :deleted_at WHERE application_uuid = :id AND deleted_at IS NULL", map[string]interface{}{
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
