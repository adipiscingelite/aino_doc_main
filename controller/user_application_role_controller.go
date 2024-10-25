package controller

import (
	"aino_document/models"
	"aino_document/service"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/badoux/checkmail"
	"github.com/labstack/echo/v4"
)

func GetUserAppRole(c echo.Context) error {
	userAppRole, err := service.GetUserApplicationRole()

	if err != nil {
		log.Print(err)
		return c.JSON(http.StatusInternalServerError, &models.Response{
			Code:    500,
			Message: "Terjadi kesalahan internal server. Mohon coba beberapa saat lagi",
			Status:  false,
		})
	}
	return c.JSON(http.StatusOK, userAppRole)
}
func GetUserByDivision(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken"

	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak ditemukan!",
			"status":  false,
		})
	}

	// Periksa apakah tokenString mengandung "Bearer "
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	// Hapus "Bearer " dari tokenString
	tokenOnly := strings.TrimPrefix(tokenString, "Bearer ")

	//dekripsi token JWE
	decrypted, err := DecryptJWE(tokenOnly, secretKey)
	if err != nil {
		fmt.Println("Gagal mendekripsi token:", err)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	var claims JwtCustomClaims
	errJ := json.Unmarshal([]byte(decrypted), &claims)
	if errJ != nil {
		fmt.Println("Gagal mengurai klaim:", errJ)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}
	// code := c.Get("division_code").(string)
	// _, errK := service.GetUserInfoFromToken(tokenOnly)
	// if errK != nil {
	// 	return c.JSON(http.StatusUnauthorized, "Invalid token atau token tidak ditemukan!")
	// }

	// c.Set("code", code)
	// userAppRole, err := service.GetUserByDivision(code)
	title := c.Get("division_title").(string)
	_, errK := service.GetUserInfoFromToken(tokenOnly)
	if errK != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid token atau token tidak ditemukan!")
	}

	divisionTitle := title

	c.Set("title", divisionTitle)
	userAppRole, err := service.GetUserByDivision(divisionTitle)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Print(err)
			response := models.Response{
				Code:    404,
				Message: "User tidak ditemukan!",
				Status:  false,
			}
			return c.JSON(http.StatusNotFound, response)
		} else {
			log.Print(err)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
				Status:  false,
			})
		}
	}
	return c.JSON(http.StatusOK, userAppRole)
}

func ShowAppRoleById(c echo.Context) error {
	id := c.Param("id")

	var getAppRole models.Users

	getAppRole, err := service.GetSpecUseApplicationRole(id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print(err)
			response := models.Response{
				Code:    404,
				Message: "User tidak ditemukan!",
				Status:  false,
			}
			return c.JSON(http.StatusNotFound, response)
		} else {
			log.Print(err)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
				Status:  false,
			})
		}
	}

	return c.JSON(http.StatusOK, getAppRole)
}

func UpdateUserAppRole(c echo.Context) error {

	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken"

	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak ditemukan!",
			"status":  false,
		})
	}

	if !strings.HasPrefix(tokenString, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	tokenOnly := strings.TrimPrefix(tokenString, "Bearer ")

	// Mendekripsi token JWE
	decrypted, errDec := DecryptJWE(tokenOnly, secretKey)
	if errDec != nil {
		fmt.Println("Gagal mendekripsi token:", errDec)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	var claims JwtCustomClaims
	errJ := json.Unmarshal([]byte(decrypted), &claims)
	if errJ != nil {
		fmt.Println("Gagal mengurai klaim:", errJ)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}
	//userUUID := c.Get("user_uuid").(string)
	_, errK := service.GetUserInfoFromToken(tokenOnly)
	if errK != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid token atau token tidak ditemukan!")
	}

	var updateUserAppRole models.UpdateUser

	if errBind := c.Bind(&updateUserAppRole); errBind != nil {
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data tidak valid!",
			Status:  false,
		})
	}

	errEmail := checkmail.ValidateFormat(updateUserAppRole.Email)
	if errEmail != nil {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Format email tidak valid",
			Status:  false,
		})
	}

	err := c.Validate(&updateUserAppRole)

	if err == nil {
		userApplicationRoleUUID := c.Param("user_application_role_uuid")
		log.Println("user_application_role_uuid from URL:", userApplicationRoleUUID)

		prevUser, errGet := service.GetSpecUseApplicationRole(userApplicationRoleUUID)
		if errGet != nil {
			return c.JSON(http.StatusNotFound, &models.Response{
				Code:    404,
				Message: "Gagal mengupdate user. User tidak ditemukan!",
				Status:  false,
			})
		}
		existingUser, err := service.GetUserByUsernameAndEmail(userApplicationRoleUUID)
		if err != nil {
			log.Printf("Error getting existing user data: %v", err)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server.",
				Status:  false,
			})
		}

		if updateUserAppRole.Username != existingUser.UserName || updateUserAppRole.Email != existingUser.UserEmail {
			isUnique, err := service.IsUniqueUsernameOrEmail(userApplicationRoleUUID, updateUserAppRole.Username, updateUserAppRole.Email)
			if err != nil {
				log.Println("Error checking uniqueness:", err)
				return c.JSON(http.StatusInternalServerError, &models.Response{
					Code:    500,
					Message: "Terjadi kesalahan internal pada server.",
					Status:  false,
				})
			}

			if !isUnique {
				log.Println("Username atau email telah digunakan oleh data lain.")
				return c.JSON(http.StatusBadRequest, &models.Response{
					Code:    400,
					Message: "Username atau email telah digunakan!",
					Status:  false,
				})
			}
		}
		re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
		if !re.MatchString(updateUserAppRole.PersonalPhone) {
			log.Println("Nomor telepon tidak valid:", updateUserAppRole.PersonalPhone)
			return c.JSON(http.StatusUnprocessableEntity, &models.Response{
				Code:    422,
				Message: "Nomor telepon tidak valid",
				Status:  false,
			})
		}

		if len(updateUserAppRole.PersonalPhone) < 11 || len(updateUserAppRole.PersonalPhone) > 15 {
			log.Println("Nomor telepon tidak sesuai panjang:", updateUserAppRole.PersonalPhone)
			return c.JSON(http.StatusUnprocessableEntity, &models.Response{
				Code:    422,
				Message: "Nomor telepon harus antara 11 dan 15 digit",
				Status:  false,
			})
		}

		if !IsValidGender(updateUserAppRole.PersonalGender) {
			log.Println("Gender tidak valid:", updateUserAppRole.PersonalGender)
			return c.JSON(http.StatusUnprocessableEntity, &models.Response{
				Code:    422,
				Message: "Gender tidak valid",
				Status:  false,
			})
		}
		_, addroleErr := service.UpdateUserAppRole(updateUserAppRole, userApplicationRoleUUID)
		if addroleErr != nil {
			if strings.Contains(addroleErr.Error(), "parsing time") {
				log.Println("Format tanggal tidak valid:", addroleErr)
				return c.JSON(http.StatusBadRequest, &models.Response{
					Code:    400,
					Message: "Format tanggal tidak valid",
					Status:  false,
				})
			}
			log.Printf("Error updating user application role: %v", addroleErr)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server.",
				Status:  false,
			})
		}

		log.Println(prevUser)
		return c.JSON(http.StatusOK, &models.Response{
			Code:    200,
			Message: "Berhasil mengupdate akun!",
			Status:  true,
		})
	} else {
		log.Println(err)
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Data tidak boleh kosong!",
			Status:  false,
		})
	}
}

func DeleteUserAppRole(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken"

	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak ditemukan!",
			"status":  false,
		})
	}

	// Periksa apakah tokenString mengandung "Bearer "
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	// Hapus "Bearer " dari tokenString
	tokenOnly := strings.TrimPrefix(tokenString, "Bearer ")

	//dekripsi token JWE
	decrypted, err := DecryptJWE(tokenOnly, secretKey)
	if err != nil {
		fmt.Println("Gagal mendekripsi token:", err)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	var claims JwtCustomClaims
	errJ := json.Unmarshal([]byte(decrypted), &claims)
	if errJ != nil {
		fmt.Println("Gagal mengurai klaim:", errJ)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}
	userUUID := c.Get("user_uuid").(string)
	_, errK := service.GetUserInfoFromToken(tokenOnly)
	if errK != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid token atau token tidak ditemukan!")
	}

	id := c.Param("id")

	if err == nil {
		errService := service.DeleteUserAppRole(id, userUUID)
		if errService == service.ErrNotFound {
			log.Printf("error kenapa %s", errService)
			return c.JSON(http.StatusNotFound, &models.Response{
				Code:    404,
				Message: "Gagal menghapus user. User tidak ditemukan!",
				Status:  false,
			})
		}

		return c.JSON(http.StatusOK, &models.Response{
			Code:    200,
			Message: "User berhasil dihapus!",
			Status:  true,
		})
	} else {
		log.Println("Kesalahan sebelum pembaruan:", err)
		return c.JSON(http.StatusInternalServerError, &models.Response{
			Code:    500,
			Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi",
			Status:  false,
		})
	}
}

func GetAllPersonal(c echo.Context) error {
	documents, err := service.GetAllPersonalName()
	if err != nil {
		log.Print(err)
		response := models.Response{
			Code:    500,
			Message: "Terjadi kesalahan internal server. Mohon coba beberapa saat lagi",
			Status:  false,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusOK, documents)

}
