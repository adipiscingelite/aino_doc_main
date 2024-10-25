package controller

import (
	"aino_document/database"
	"aino_document/models"
	"aino_document/service"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

var db *sqlx.DB = database.Connection()

func AddRole(c echo.Context) error {
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

	var addRole models.Role

	if err := c.Bind(&addRole); err != nil {
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data tidak valid!",
			Status:  false,
		})
	}

	errVal := c.Validate(&addRole)

	if errVal == nil {
		var existingRoleID int
		err := db.QueryRow("SELECT role_id FROM role_ms WHERE (role_title = $1 OR role_code = $2) AND deleted_at IS NULL", addRole.Title, addRole.Code).Scan(&existingRoleID)

		if err == nil {
			return c.JSON(http.StatusBadRequest, &models.Response{
				Code:    400,
				Message: "Gagal menambahkan role. Role sudah ada!",
				Status:  false,
			})
		} else {
			addroleErr := service.AddRole(addRole, userUUID)
			if addroleErr != nil {
				return c.JSON(http.StatusInternalServerError, &models.Response{
					Code:    500,
					Message: "Terjadi kesalahan internal pada server.",
					Status:  false,
				})
			}

			return c.JSON(http.StatusCreated, &models.Response{
				Code:    201,
				Message: "Berhasil menambahkan role!",
				Status:  true,
			})
		}
	} else {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Data tidak boleh kosong!",
			Status:  false,
		})
	}

}

func GetAllRole(c echo.Context) error {
	role, err := service.GetAllRole()

	if err != nil {
		log.Print(err)
		response := models.Response{
			Code:    500,
			Message: "Terjadi kesalahan internal server. Mohon coba beberapa saat lagi",
			Status:  false,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusOK, role)
}

func ShowRoleById(c echo.Context) error {
	id := c.Param("id")

	var getRole models.Roles

	getRole, err := service.ShowRoleById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			response := models.Response{
				Code:    404,
				Message: "Role tidak ditemukan!",
				Status:  false,
			}
			return c.JSON(http.StatusNotFound, response)
		} else {
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
				Status:  false,
			})
		}
	}

	return c.JSON(http.StatusOK, getRole)
}

func GetRoleById(c echo.Context) error {
	id := c.Param("id")

	var getRole models.Role

	getRole, err := service.GetRoleById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			response := models.Response{
				Code:    404,
				Message: "Role tidak ditemukan!",
				Status:  false,
			}
			return c.JSON(http.StatusNotFound, response)
		} else {
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
				Status:  false,
			})
		}
	}

	return c.JSON(http.StatusOK, getRole)
}

func UpdateRole(c echo.Context) error {
	
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

	// dekripsi token JWE
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

	perviousContent, errGet := service.ShowRoleById(id)
	if errGet != nil {
		return c.JSON(http.StatusNotFound, &models.Response{
			Code:    404,
			Message: "Gagal mengupdate role. Role tidak ditemukan!",
			Status:  false,
		})
	}

	var editRole models.Role
	if err := c.Bind(&editRole); err != nil {
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data invalid!",
			Status:  false,
		})
	}

	errValidate := c.Validate(&editRole)
	if errValidate != nil {
		return c.JSON(http.StatusInternalServerError, &models.Response{
			Code:    422,
			Message: "Data tidak boleh kosong!",
			Status:  false,
		})
	}
	if err == nil {
		exsitingRole, err := service.GetRoleCodeAndTitle(id)
		if err != nil {
			log.Printf("Error getting existing user data: %v", err)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server.",
				Status:  false,
			})
		}

		if editRole.Code != exsitingRole.Code || editRole.Title != exsitingRole.Title {
			isUnique, err := service.IsUniqueRole(id, editRole.Code, editRole.Title)
			if err != nil {
				log.Println("Error checking uniqueness:", err)
				return c.JSON(http.StatusInternalServerError, &models.Response{
					Code:    500,
					Message: "Terjadi kesalahan internal pada server.",
					Status:  false,
				})
			}

			if !isUnique {
				log.Println("Role sudah ad! Role tidak boleh sama!")
				return c.JSON(http.StatusBadRequest, &models.Response{
					Code:    400,
					Message: "Role sudah ada! Role tidak boleh sama!",
					Status:  false,
				})
			}
		}

		_, errService := service.UpdateRole(editRole, id, userUUID)
		if errService != nil {
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi",
				Status:  false,
			})
		}

		log.Println(perviousContent)
		return c.JSON(http.StatusOK, &models.Response{
			Code:    200,
			Message: "Role berhasil diperbarui!",
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

func DeleteRole(c echo.Context) error {
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
		errService := service.DeleteRole(id, userUUID)
		if errService == service.ErrNotFound {
			return c.JSON(http.StatusNotFound, &models.Response{
				Code:    404,
				Message: "Gagal menghapus role. Role tidak ditemukan!",
				Status:  false,
			})
		}

		return c.JSON(http.StatusOK, &models.Response{
			Code:    200,
			Message: "Role berhasil dihapus!",
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
