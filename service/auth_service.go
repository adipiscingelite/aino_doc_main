package service

import (
	"aino_document/models"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
	"golang.org/x/crypto/bcrypt"
)

type ValidationError struct {
	Message string
	Field   string
	Tag     string
}

func (ve *ValidationError) Error() string {
	return ve.Message
}

func RegisterUser(userRegister models.Register, userUUID string) error {
	if len(userRegister.Password) < 8 {
		return &ValidationError{
			Message: "Password should be of 8 characters long",
			Field:   "password",
			Tag:     "strong_password",
		}
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRegister.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashedPasswordStr := base64.StdEncoding.EncodeToString(hashedPassword)
	fmt.Println(hashedPassword)
	fmt.Println(hashedPasswordStr)

	username, errP := GetUsernameByID(userUUID)
	if errP != nil {
		return errP
	}
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()

	user_id := currentTimestamp + int64(uniqueID)
	uuid := uuid.New()
	uuidString := uuid.String()
	_, errInsert := db.NamedExec("INSERT INTO user_ms (user_id, user_uuid, user_name, user_email, user_password, created_by) VALUES (:user_id, :user_uuid, :user_name, :user_email, :user_password, :created_by)", map[string]interface{}{
		"user_id":       user_id,
		"user_uuid":     uuidString,
		"user_name":     userRegister.Username,
		"user_email":    userRegister.Email,
		"user_password": hashedPasswordStr,
		"created_by":    username,
	})

	if errInsert != nil {
		log.Print(errInsert)
		return errInsert
	}

	err = db.Get(&user_id, "SELECT user_id FROM user_ms WHERE user_name = $1", userRegister.Username)
	if err != nil {
		return err
	}

	// Mendapatkan role_id yang baru saja diinsert
	log.Printf("Selected Role UUID: %s", userRegister.ApplicationRole.Role_UUID)

	var roleID int64
	err = db.Get(&roleID, "SELECT role_id FROM role_ms WHERE role_uuid = $1 AND deleted_at IS NULL", userRegister.ApplicationRole.Role_UUID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Role not found for role_uuid: %s", userRegister.ApplicationRole.Role_UUID)
			return err
		}
		log.Printf("Error getting role_id: %v", err)
		return err
	}
	log.Printf("Obtained Role ID: %d", roleID)
	var applicationID int64
	err = db.Get(&applicationID, "SELECT application_id FROM application_ms WHERE application_uuid = $1 AND deleted_at IS NULL", userRegister.ApplicationRole.Application_UUID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Application not found for application_uuid: %s", userRegister.ApplicationRole.Application_UUID)
			return err
		}
		log.Println("Error getting application_id:", err)
		return err
	}

	// Get division_id
	var divisionID int64
	err = db.Get(&divisionID, "SELECT division_id FROM division_ms WHERE division_uuid = $1 AND deleted_at IS NULL", userRegister.ApplicationRole.Division_UUID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Division not found for division_uuid: %s", userRegister.ApplicationRole.Division_UUID)
			return err
		}
		log.Println("Error fetching division_id:", err)
		return err
	}

	AppRoleId := currentTimestamp + int64(uniqueID)
	uudiNew := uuid.String()
	// Insert data ke application_role_ms
	_, err = db.Exec("INSERT INTO application_role_ms(application_role_uuid, application_role_id, application_id, role_id, created_by) VALUES ($1, $2, $3, $4, $5)",
		uudiNew, AppRoleId, applicationID, roleID, username)
	if err != nil {
		log.Println("Error inserting data into application_role_ms:", err)
		return err
	}

	// Get application_role_id
	var applicationRoleID int64
	err = db.Get(&applicationRoleID, "SELECT application_role_id FROM application_role_ms WHERE application_id = $1 AND role_id = $2",
		applicationID, roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Application Role not found for application_uuid: %s", userRegister.ApplicationRole.Application_UUID)
			return err
		}
		log.Println("Error getting application_role_id:", err)
		return err
	}
	log.Println("Application Role ID:", applicationRoleID)

	layout := "2006-01-02"
	birthday := userRegister.PersonalBirthday
	// Konversi input tanggal ke time.Time
	parsedDate, err := time.Parse(layout, birthday)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Println(parsedDate)

	if err != nil {
		log.Fatal("Format tanggal tidak valid:", err)
	}

	phoneNumber := userRegister.PersonalPhone
	// Parse nomor telepon
	num, err := phonenumbers.Parse(phoneNumber, "ID")
	if err != nil {
		fmt.Println("Error parsing phone number:", err)
		return err
	}

	// Format nomor telepon dalam format nasional
	formattedNum := phonenumbers.Format(num, phonenumbers.NATIONAL)

	// Tampilkan hasil
	fmt.Println("Nomor telepon yang diformat:", formattedNum)
	_, errInsert = db.NamedExec("INSERT INTO personal_data_ms (personal_id, personal_uuid, division_id, user_id, personal_name, personal_birthday, personal_gender, personal_phone, personal_address) VALUES (:personal_id, :personal_uuid, :division_id, :user_id, :personal_name, :personal_birthday, :personal_gender, :personal_phone, :personal_address)", map[string]interface{}{
		"personal_id":       currentTimestamp + int64(uniqueID),
		"personal_uuid":     uuidString,
		"division_id":       divisionID,
		"user_id":           user_id,
		"personal_name":     userRegister.PersonalName,
		"personal_birthday": userRegister.PersonalBirthday,
		"personal_gender":   userRegister.PersonalGender,
		"personal_phone":    formattedNum,
		"personal_address":  userRegister.PersonalAddress,
	})

	if errInsert != nil {
		log.Print(errInsert)
		return err
	}
	// Insert user_application_role_ms data
	_, err = db.Exec("INSERT INTO user_application_role_ms(user_application_role_uuid, user_id, application_role_id, division_id, created_by) VALUES ($1, $2, $3, $4, $5)", uuidString, user_id, applicationRoleID, divisionID, username)
	if err != nil {
		log.Println("Error inserting data into user_application_role_ms:", err)
		return err
	}

	return nil
}

func Login(userLogin models.Login) (string, string, string, string, string, int, bool, error) {
	var isAuthentication bool
	var user_id int
	var user_uuid string
	var role_code string
	// var application_role_id int
	var division_title string
	var division_code string
	var username string

	rows, err := db.Query("SELECT CASE WHEN COUNT(*) > 0 THEN 'true' ELSE 'false' END FROM user_ms WHERE user_email = $1 AND user_password = $2", userLogin.Email, userLogin.Password)
	if err != nil {
		return "", "", "", "", "", 0, false, err
	}

	defer rows.Close()

	rows, err = db.Query("SELECT user_uuid, user_password, user_id  from user_ms where user_email = $1", userLogin.Email)
	if err != nil {
		fmt.Println("Error querying users:", err)
		return "", "", "", "", "", 0, false, err
	}

	defer rows.Close()

	var dbPasswordBase64 string
	if rows.Next() {
		err = rows.Scan(&user_uuid, &dbPasswordBase64, &user_id)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			return "", "", "", "", "", 0, false, err
		}
		dbPassword, errBycript := base64.StdEncoding.DecodeString(dbPasswordBase64)

		if errBycript != nil {
			fmt.Println("Password comparison failed:", errBycript)
			return "", "", "", "", "", 0, false, errBycript
		}
		errBycript = bcrypt.CompareHashAndPassword(dbPassword, []byte(userLogin.Password))
		if errBycript != nil {
			fmt.Println("Password comparison failed:", errBycript)
			return "", "", "", "", "", 0, false, errBycript
		}
		isAuthentication = true
	}

	rows, err = db.Query("SELECT user_id FROM user_ms WHERE user_email = $1", userLogin.Email)
	if err != nil {
		fmt.Println("Error querying user:", err)
		return "", "", "", "", "", 0, false, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&user_id)
		if err != nil {
			fmt.Println("Error scanning user row:", err)
			return "", "", "", "", "", 0, false, err
		}
	}

	if isAuthentication {
		// Query untuk mendapatkan division_code
		rows, err := db.Query("SELECT d.division_title, d.division_code FROM division_ms d JOIN user_application_role_ms uar ON d.division_id = uar.division_id JOIN user_ms u ON uar.user_id = u.user_id WHERE u.user_uuid = $1", user_uuid)
		if err != nil {
			fmt.Println("Error querying division title:", err)
			return "", "", "", "", "", 0, false, err
		}
		defer rows.Close()

		// Periksa hasil query division_code
		if rows.Next() {
			err = rows.Scan(&division_title, &division_code)
			if err != nil {
				fmt.Println("Error scanning division_title and division_code rows:", err)
				return "", "", "", "", "", 0, false, err
			}
		} else {
			fmt.Println("No division title and division code found for user UUID:", user_uuid)
		}

		// Query untuk mendapatkan role_code
		rows, err = db.Query("SELECT r.role_code FROM role_ms r JOIN application_role_ms ar ON r.role_id = ar.role_id JOIN user_application_role_ms uar ON ar.application_role_id = uar.application_role_id JOIN user_ms u ON u.user_id = uar.user_id WHERE u.user_uuid = $1", user_uuid)
		if err != nil {
			fmt.Println("Error querying role code:", err)
			return "", "", "", "", "", 0, false, err
		}
		defer rows.Close()

		// Periksa hasil query role_code
		if rows.Next() {
			err = rows.Scan(&role_code)
			if err != nil {
				fmt.Println("Error scanning role_code row:", err)
				return "", "", "", "", "", 0, false, err
			}
		}
		// Query untuk mendapatkan username
		rows, err = db.Query("SELECT user_name FROM user_ms WHERE user_uuid = $1", user_uuid)
		if err != nil {
			fmt.Println("Error querying username:", err)
			return "", "", "", "", "", 0, false, err
		}
		defer rows.Close()

		// Periksa hasil query username
		if rows.Next() {
			err = rows.Scan(&username)
			if err != nil {
				fmt.Println("Error scanning username row:", err)
				return "", "", "", "", "", 0, false, err
			}
		}

		// if rows.Next() {
		// 	err = rows.Scan(&application_role_id, &division_id)
		// 	if err != nil {
		// 		fmt.Println("Error scanning role row:", err)
		// 		return 0, false, 0, 0, err
		// 	}
		// }
		return user_uuid, role_code, division_title, division_code, username, user_id, isAuthentication, nil
	}
	return "", "", "", "", "", 0, false, nil // Jika tidak ada authentikasi yang berhasil

}

func ChangePassword(changePassword models.ChangePasswordRequest, userUUID string) error {
	var dbPassword string
	err := db.Get(&dbPassword, "SELECT user_password FROM user_ms WHERE user_uuid = $1", userUUID)
	if err != nil {
		return err
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(dbPassword)
	if err != nil {
		return err
	}

	if len(changePassword.NewPassword) < 8 {
		return &ValidationError{
			Message: "Password should be of 8 characters long",
			Field:   "password",
			Tag:     "strong_password",
		}
	}
	errBycript := bcrypt.CompareHashAndPassword(decodedPassword, []byte(changePassword.OldPassword))
	if errBycript != nil {
		fmt.Println("Error comparing old passwords:", errBycript)
		return errBycript
	}

	if changePassword.OldPassword == changePassword.NewPassword {
		return errors.New("new password must be different from old password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(changePassword.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error generating hashed password:", err)
		return err
	}

	hashedPasswordStr := base64.StdEncoding.EncodeToString(hashedPassword)
	_, err = db.NamedExec("UPDATE user_ms SET user_password = :user_password WHERE user_uuid = :user_uuid", map[string]interface{}{
		"user_password": hashedPasswordStr,
		"user_uuid":     userUUID,
	})

	if err != nil {
		fmt.Println("Error updating password in database:", err)
		return err
	}
	return nil
}
