package middleware

import (
	"aino_document/database"
	"aino_document/models"
	"aino_document/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	jose "github.com/dvsekhvalnov/jose2go"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

var db *sqlx.DB = database.Connection()

type JwtCustomClaims struct {
	// UserId   int    `json:"user_id"`
	UserUUID           string `json:"user_uuid"`
	AppRoleId          int    `json:"application_role_id"`
	DivisionTitle      string `json:"division_title"`
	DivisionCode       string `json:"division_code"`
	RoleCode           string `json:"role_code"`
	jwt.StandardClaims        // Embed the StandardClaims struct

}

func DecryptJWE(jweToken string, secretKey string) (string, error) {
	// Dekripsi token JWE
	decrypted, _, err := jose.Decode(jweToken, secretKey)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}

func ExtractClaims(jwtToken string) (JwtCustomClaims, error) {
	claims := &JwtCustomClaims{}
	secretKey := "secretJwToken" // Ganti dengan kunci yang benar

	token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return JwtCustomClaims{}, err
	}

	return *claims, nil
}

func SuperAdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		secretKey := "secretJwToken" // Ganti dengan kunci yang benar
		_, exists := utils.InvalidTokens[tokenString]
		if exists {
			return c.JSON(http.StatusUnauthorized, &models.Response{
				Code:    401,
				Message: "Token tidak valid atau Anda telah logout",
				Status:  false,
			})
		}
		// Periksa apakah tokenString tidak kosong
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "Token tidak ditemukan!",
				"status":  false,
			})
		}
		// _, exists := InvalidTokens[tokenString]
		// if exists {
		// 	return c.JSON(http.StatusUnauthorized, &models.Response{
		// 		Code:    401,
		// 		Message: "Token tidak valid atau Anda telah logout",
		// 		Status:  false,
		// 	})
		// }

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

		// Langkah 1: Mendekripsi token JWE
		decrypted, err := DecryptJWE(tokenOnly, secretKey)
		if err != nil {
			fmt.Println("Gagal mendekripsi token:", err)
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "Token tidak valid!",
				"status":  false,
			})
		}

		fmt.Println("Token yang sudah dideskripsi:", decrypted)

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
		if time.Now().Unix() > claims.StandardClaims.ExpiresAt {
			// Token telah kedaluwarsa
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "Sesi Anda sudah habis! Silahkan login kembali.",
				"status":  false,
			})
		}
		// Sekarang Anda memiliki data dalam struct JwtCustomClaims
		// Anda bisa mengakses UserId atau klaim lain sesuai kebutuhan
		// fmt.Println("UserID:", claims.UserId)

		userUUID := claims.UserUUID // Mengakses UserID langsung
		roleID := claims.AppRoleId
		divisionTitle := claims.DivisionTitle
		roleCode := claims.RoleCode
		divisionCode := claims.DivisionCode
		if roleCode != "" {
			log.Print(roleCode)
		}

		fmt.Println("User UUID:", userUUID)
		fmt.Println("Role Code:", roleCode)
		fmt.Println("Division title:", divisionTitle)
		fmt.Println("Division Code : ", divisionCode)
		c.Set("division_code", divisionCode)
		c.Set("user_uuid", userUUID)
		c.Set("application_role_id", roleID)
		c.Set("division_title", divisionTitle)
		c.Set("role_code", roleCode)

		if roleCode != "SA" {
			return c.JSON(http.StatusForbidden, &models.Response{
				Code:    403,
				Message: "Akses ditolak!",
				Status:  false,
			})
		}

		// Token JWE valid, Anda dapat melanjutkan dengan pengolahan berikutnya
		return next(c)
	}
}

func CheckRolePermission(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Ambil informasi pengguna dari token
		roleCode := c.Get("role_code").(string)

		// Ambil izin dari database berdasarkan role_code
		permissions, err := getRolePermissionsFromDB(roleCode)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
		}

		// Set izin dalam konteks
		c.Set("permissions", permissions)

		menuID := c.Param("id")

		// Panggil fungsi untuk mendapatkan required_permission dari DB
		requiredPermission, err := getRequiredPermissionFromDB(menuID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
		}

		c.Set("required_permission", requiredPermission)

		// Periksa izin
		if !hasPermission(permissions, requiredPermission) {
			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"error":               "Permission denied",
				"required_permission": requiredPermission,
				"user_permissions":    permissions,
			})
		}

		return next(c)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func hasPermission(userPermissions, requiredPermission string) bool {
	userPermissionsSlice := strings.Split(userPermissions, ", ")
	requiredPermissionSlice := strings.Split(requiredPermission, ", ")

	for _, rp := range requiredPermissionSlice {
		if !contains(userPermissionsSlice, rp) {
			return false
		}
	}

	return true
}

func getRolePermissionsFromDB(roleCode string) (string, error) {
	var permissionsStr string

	err := db.Get(&permissionsStr, "SELECT permissions FROM role_ms WHERE role_code = $1", roleCode)
	if err != nil {
		log.Println("Error querying permissions from database:", err)
		return "", err
	}
	return permissionsStr, nil
}

func getRequiredPermissionFromDB(menuID string) (string, error) {
	var requiredPermission string
	err := db.Get(&requiredPermission, "SELECT required_permission FROM menu_ms WHERE menu_uuid = $1", menuID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Tidak ada baris yang sesuai
			log.Println("No rows found for menu_uuid:", menuID)
			return "", nil
		}
		log.Println("Error querying required_permission:", err)
		return "", err
	}
	return requiredPermission, nil
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		secretKey := "secretJwToken" // Ganti dengan kunci yang benar
		_, exists := utils.InvalidTokens[tokenString]
		if exists {
			return c.JSON(http.StatusUnauthorized, &models.Response{
				Code:    401,
				Message: "Token tidak valid atau Anda telah logout",
				Status:  false,
			})
		}
		// Periksa apakah tokenString tidak kosong
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

		// Langkah 1: Mendekripsi token JWE
		decrypted, err := DecryptJWE(tokenOnly, secretKey)
		if err != nil {
			fmt.Println("Gagal mendekripsi token:", err)
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "Token tidak valid!",
				"status":  false,
			})
		}

		fmt.Println("Token", tokenOnly)
		fmt.Println("Token yang sudah dideskripsi:", decrypted)

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

		if time.Now().Unix() > claims.StandardClaims.ExpiresAt {
			// Token telah kedaluwarsa
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "Sesi Anda sudah habis! Silahkan login kembali.",
				"status":  false,
			})
		}

		// Sekarang Anda memiliki data dalam struct JwtCustomClaims
		// Anda bisa mengakses UserId atau klaim lain sesuai kebutuhan
		// fmt.Println("UserID:", claims.UserId)

		userUUID := claims.UserUUID // Mengakses UserID langsung
		roleID := claims.AppRoleId
		divisionTitle := claims.DivisionTitle
		roleCode := claims.RoleCode
		if roleCode != "" {
			log.Print(roleCode)
		}

		fmt.Println("User UUID:", userUUID)
		fmt.Println("Role Code:", roleCode)

		c.Set("user_uuid", userUUID)
		c.Set("application_role_id", roleID)
		c.Set("division_title", divisionTitle)
		c.Set("role_code", roleCode)

		// Token JWE valid, Anda dapat melanjutkan dengan pengolahan berikutnya
		return next(c)
	}
}

func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		secretKey := "secretJwToken" // Ganti dengan kunci yang benar
		_, exists := utils.InvalidTokens[tokenString]
		if exists {
			return c.JSON(http.StatusUnauthorized, &models.Response{
				Code:    401,
				Message: "Token tidak valid atau Anda telah logout",
				Status:  false,
			})
		}
		// Periksa apakah tokenString tidak kosong
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

		// Langkah 1: Mendekripsi token JWE
		decrypted, err := DecryptJWE(tokenOnly, secretKey)
		if err != nil {
			fmt.Println("Gagal mendekripsi token:", err)
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "Token tidak valid!",
				"status":  false,
			})
		}

		fmt.Println("Token yang sudah dideskripsi:", decrypted)

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

		if time.Now().Unix() > claims.StandardClaims.ExpiresAt {
			// Token telah kedaluwarsa
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "Sesi Anda sudah habis! Silahkan login kembali.",
				"status":  false,
			})
		}
		// Sekarang Anda memiliki data dalam struct JwtCustomClaims
		// Anda bisa mengakses UserId atau klaim lain sesuai kebutuhan
		// fmt.Println("UserID:", claims.UserId)

		userUUID := claims.UserUUID // Mengakses UserID langsung
		roleID := claims.AppRoleId
		divisionTitle := claims.DivisionTitle
		roleCode := claims.RoleCode
		if roleCode != "" {
			log.Print(roleCode)
		}

		fmt.Println("User UUID:", userUUID)
		fmt.Println("Role Code:", roleCode)
		fmt.Println("Division title:", divisionTitle)

		c.Set("user_uuid", userUUID)
		c.Set("application_role_id", roleID)
		c.Set("division_title", divisionTitle)
		c.Set("role_code", roleCode)
		if roleCode != "A" {
			return c.JSON(http.StatusForbidden, &models.Response{
				Code:    403,
				Message: "Akses ditolak!",
				Status:  false,
			})
		}

		// Token JWE valid, Anda dapat melanjutkan dengan pengolahan berikutnya
		return next(c)
	}
}
