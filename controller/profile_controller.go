package controller

import (
	"aino_document/models"
	"aino_document/service"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func MyProfile(c echo.Context) error {

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

	id := c.Get("user_uuid").(string)
	_, errK := service.GetUserInfoFromToken(tokenOnly)
	if errK != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid token atau token tidak ditemukan!")
	}

	myProfile, err := service.MyProfile(id)

	if err != nil {
		if err == sql.ErrNoRows {
			response := models.Response{
				Code:    404,
				Message: "Profile tidak ditemukan!",
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

	return c.JSON(http.StatusOK, myProfile)
}

// func MyProfile(c echo.Context) error {
// 	// Mendapatkan klaim token dari konteks Echo
// 	claims := c.Get("claims").(JwtCustomClaims)

// 	// Mengambil UUID pengguna dari klaim token
// 	userUUID := claims.UserUUID

// 	// Panggil service untuk mendapatkan profil pengguna berdasarkan UUID
// 	myProfile, err := service.MyProfile(userUUID)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			response := models.Response{
// 				Code:    404,
// 				Message: "Role tidak ditemukan!",
// 				Status:  false,
// 			}
// 			return c.JSON(http.StatusNotFound, response)
// 		} else {
// 			return c.JSON(http.StatusInternalServerError, &models.Response{
// 				Code:    500,
// 				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
// 				Status:  false,
// 			})
// 		}
// 	}

// 	// Mengembalikan profil pengguna dalam respons JSON
// 	return c.JSON(http.StatusOK, myProfile)
// }
