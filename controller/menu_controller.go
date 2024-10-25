package controller

import (
	"aino_document/models"
	"aino_document/service"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func ValidateToken(c echo.Context, secretKey string) (*JwtCustomClaims, string, error) {
	tokenString := c.Request().Header.Get("Authorization")

	if tokenString == "" {
		return nil, "", echo.NewHTTPError(http.StatusUnauthorized, "Token tidak ditemukan!")
	}

	// Periksa apakah tokenString mengandung "Bearer "
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return nil, "", echo.NewHTTPError(http.StatusUnauthorized, "Token tidak valid!")
	}

	// Hapus "Bearer " dari tokenString
	tokenOnly := strings.TrimPrefix(tokenString, "Bearer ")

	// Dekripsi token JWE
	decrypted, err := DecryptJWE(tokenOnly, secretKey)
	if err != nil {
		return nil, "", echo.NewHTTPError(http.StatusUnauthorized, "Token tidak valid!")
	}

	var claims JwtCustomClaims
	errJ := json.Unmarshal([]byte(decrypted), &claims)
	if errJ != nil {
		return nil, "", echo.NewHTTPError(http.StatusUnauthorized, "Token tidak valid!")
	}

	userUUID := c.Get("user_uuid").(string)
	_, errK := service.GetUserInfoFromToken(tokenOnly)
	if errK != nil {
		return nil, "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid token atau token tidak ditemukan!")
	}

	return &claims, userUUID, nil
}

func AddMenu(c echo.Context) error {
	secretKey := "secretJwToken"

	_, userUUID, err := ValidateToken(c, secretKey)
	if err != nil {
		return c.JSON(err.(*echo.HTTPError).Code, map[string]interface{}{
			"code":    err.(*echo.HTTPError).Code,
			"message": err.Error(),
			"status":  false,
		})
	}

	var addMenu models.Menu

	if err := c.Bind(&addMenu); err != nil {
		log.Print(err)
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data tidak valid!",
			Status:  false,
		})
	}

	errVal := c.Validate(&addMenu)

	if errVal == nil {
		var existingRoleID int
		err := db.QueryRow("SELECT menu_id FROM menu_ms WHERE (menu_title = $1) AND deleted_at IS NULL", addMenu.Title).Scan(&existingRoleID)

		if err == nil {
			return c.JSON(http.StatusBadRequest, &models.Response{
				Code:    400,
				Message: "Gagal menambahkan menu. Menu sudah ada!",
				Status:  false,
			})
		} else {
			addmenuErr := service.AddMenu(addMenu, userUUID)
			if addmenuErr != nil {
				return c.JSON(http.StatusInternalServerError, &models.Response{
					Code:    500,
					Message: "Terjadi kesalahan internal pada server.",
					Status:  false,
				})
			}

			return c.JSON(http.StatusCreated, &models.Response{
				Code:    201,
				Message: "Berhasil menambahkan Menu!",
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

func GetAllMenu(c echo.Context) error {
	menu, err := service.GetAllMenu()

	if err != nil {
		log.Print(err)
		response := models.Response{
			Code:    500,
			Message: "Terjadi kesalahan internal server. Mohon coba beberapa saat lagi",
			Status:  false,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusOK, menu)
}

func ShowMenuById(c echo.Context) error {
	id := c.Param("id")

	var getMenu models.Menu

	getMenu, err := service.ShowMenuById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			response := models.Response{
				Code:    404,
				Message: "Menu tidak ditemukan!",
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

	return c.JSON(http.StatusOK, getMenu)
}
