package controller

import (
	"aino_document/models"
	"aino_document/service"
	"aino_document/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	jose "github.com/dvsekhvalnov/jose2go"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

	_ "github.com/dgrijalva/jwt-go"
)

type TokenCheck struct {
	Token string `json:"token"`
}

type JwtCustomClaims struct {
	UserID   int    `json:"user_id"`
	UserUUID string `json:"user_uuid"`
	UserName string `json:"user_name"`
	RoleCode string `json:"role_code"`
	// AppRoleId          int `json:"application_role_id"`
	DivisionTitle      string `json:"division_title"`
	DivisionCode       string `json:"division_code"`
	jwt.StandardClaims        // Embed the StandardClaims struct

}

// Simpan token yang tidak valid dalam bentuk set
var InvalidTokens = make(map[string]struct{})

// IsValidGender memeriksa apakah nilai gender yang diberikan valid
func IsValidGender(gender string) bool {
	return gender == string("Laki-laki") || gender == string("Perempuan")
}

func RegisterUser(c echo.Context) error {
	e := echo.New()
	e.Validator = &utils.CustomValidator{Validator: validator.New()}

	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken" // Ganti dengan kunci yang benar

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
	userUUID := c.Get("user_uuid").(string)
	_, errK := service.GetUserInfoFromToken(tokenOnly)
	if errK != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid token atau token tidak ditemukan!")
	}

	var userRegister models.Register

	if errBind := c.Bind(&userRegister); errBind != nil {
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data tidak valid!",
			Status:  false,
		})
	}

	errEmail := checkmail.ValidateFormat(userRegister.Email)
	if errEmail != nil {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Format email tidak valid",
			Status:  false,
		})
	}

	err := c.Validate(&userRegister)

	if err == nil {
		var existingUserID int
		err := db.QueryRow("SELECT user_id FROM user_ms WHERE (user_name = $1 OR user_email = $2) AND deleted_at IS NULL", userRegister.Username, userRegister.Email).Scan(&existingUserID)

		if err == nil {
			return c.JSON(http.StatusBadRequest, &models.Response{
				Code:    400,
				Message: "Username atau email telah digunakan!",
				Status:  false,
			})
		}

		re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
		if !re.MatchString(userRegister.PersonalPhone) {
			log.Println("Nomor telepon tidak valid:", userRegister.PersonalPhone)
			return c.JSON(http.StatusBadRequest, &models.Response{
				Code:    422,
				Message: "Nomor telepon tidak valid",
				Status:  false,
			})
		}

		if len(userRegister.PersonalPhone) < 11 || len(userRegister.PersonalPhone) > 15 {
			log.Println("Nomor telepon tidak sesuai panjang:", userRegister.PersonalPhone)
			return c.JSON(http.StatusUnprocessableEntity, &models.Response{
				Code:    422,
				Message: "Nomor telepon harus antara 11 dan 15 digit",
				Status:  false,
			})
		}
		if !IsValidGender(userRegister.PersonalGender) {
			log.Println("Gender tidak valid:", userRegister.PersonalGender)
			return c.JSON(http.StatusUnprocessableEntity, &models.Response{
				Code:    422,
				Message: "Gender tidak valid",
				Status:  false,
			})
		}

		addroleErr := service.RegisterUser(userRegister, userUUID)
		if addroleErr != nil {
			if strings.Contains(addroleErr.Error(), "parsing time") {
				log.Println("Format tanggal tidak valid:", addroleErr)
				return c.JSON(http.StatusUnprocessableEntity, &models.Response{
					Code:    422,
					Message: "Format tanggal tidak valid",
					Status:  false,
				})
			}
			if validationErr, ok := addroleErr.(*service.ValidationError); ok {
				if validationErr.Tag == "strong_password" {
					return c.JSON(http.StatusUnprocessableEntity, &models.Response{
						Code:    422,
						Message: "Password harus memiliki setidaknya 8 karakter",
						Status:  false,
					})
				}
			}
			log.Println(addroleErr)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server.",
				Status:  false,
			})

		}

		return c.JSON(http.StatusCreated, &models.Response{
			Code:    201,
			Message: "Berhasil membuat akun!",
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

func Login(c echo.Context) error {
    e := echo.New()
    e.Validator = &utils.CustomValidator{Validator: validator.New()}

    var loginbody models.Login
    if errBind := c.Bind(&loginbody); errBind != nil {
        return c.JSON(http.StatusBadRequest, &models.Response{
            Code:    400,
            Message: "Data tidak valid!",
            Status:  false,
        })
    }

    err := c.Validate(&loginbody)

    if err != nil {
        log.Println(err)
        return c.JSON(http.StatusInternalServerError, &models.Response{
            Code:    500,
            Message: "Terjadi kesalahan pada internal server. Coba beberapa saat lagi!",
            Status:  false,
        })
    }

    user_uuid, role_code, division_title, division_code, username, user_id, isAuthentication, err := service.Login(loginbody)
    if err != nil {
        fmt.Println("Gagal login:", err)
    }
    fmt.Println("isAuthentication:", isAuthentication)

    if !isAuthentication {
        fmt.Println("Authentication failed")
        return c.JSON(http.StatusUnauthorized, &models.Response{
            Code:    401,
            Message: "Akun tidak ada atau password salah",
            Status:  false,
        })
    }

    expiresAt := time.Now().Add(2 * time.Hour).Unix()
    claims := &JwtCustomClaims{
        UserUUID:       user_uuid,
        UserID:         user_id,
        UserName:       username,
        RoleCode:       role_code,
        DivisionTitle:  division_title,
        DivisionCode:   division_code,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expiresAt,
        },
    }

    // Mengonversi claims ke format JSON
    claimData, err := json.Marshal(claims)
    if err != nil {
        fmt.Println("Gagal mengonversi klaim:", err)
    }

    // Buat token JWT dengan enkripsi JWE
    secretKey := "secretJwToken"
    jweToken, err := jose.Encrypt(string(claimData), jose.PBES2_HS256_A128KW, jose.A128GCM, secretKey)
    if err != nil {
        fmt.Println("Gagal membuat token:", err)
    }

    // Mengonversi waktu kadaluarsa ke dalam format string
    expirationTime := time.Unix(expiresAt, 0).Format(time.RFC3339)

    fmt.Println("Token JWE:", jweToken)
    return c.JSON(http.StatusOK, &models.ResponseLogin{
        Code:             200,
        Message:          "Berhasil login",
        Status:           true,
        Token:            jweToken,
        TokenExpiration:  expirationTime, // Tambahkan waktu kadaluarsa di sini
    })
}


func Logout(c echo.Context) error {

	defer func() {
		if r := recover(); r != nil {
			log.Println("Panic occurred:", r)
			c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Internal Server Error",
				Status:  false,
			})
		}
	}()
	tokenString := c.Request().Header.Get("Authorization")

	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, &models.Response{
			Code:    401,
			Message: "Token tidak ditemukan",
			Status:  false,
		})
	}
	// _, exists := InvalidTokens[tokenString]
	// if exists {
	// 	return c.JSON(http.StatusUnauthorized, &models.Response{
	// 		Code:    401,
	// 		Message: "Token tidak valid atau Anda telah logout",
	// 		Status:  false,s
	// 	})
	// }

	// Panggil fungsi untuk menandai token sebagai tidak valid
	utils.InvalidateToken(tokenString)
	return c.JSON(http.StatusOK, &models.Response{
		Code:    200,
		Message: "Berhasil Logout!",
		Status:  true,
	})
}

func ChangePassword(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken" // Ganti dengan kunci yang benar

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
	userUUIDRaw := c.Get("user_uuid")
	if userUUIDRaw == nil {
		// Handle jika nilai "user_uuid" nil
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "User UUID not found!",
			"status":  false,
		})
	}

	userUUID, ok := userUUIDRaw.(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Invalid User UUID format!",
			"status":  false,
		})
	}

	_, errK := service.GetUserInfoFromToken(tokenOnly)
	if errK != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid token atau token tidak ditemukan!")
	}

	var passUpdate models.ChangePasswordRequest
	if err := c.Bind(&passUpdate); err != nil {
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Invalid request data",
			Status:  false,
		})
	}

	if err := c.Validate(&passUpdate); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Invalid data! Password tidak boleh kosong!",
			Status:  false,
		})
	}

	errS := service.ChangePassword(passUpdate, userUUID)
	if errS != nil {

		if validationErr, ok := errS.(*service.ValidationError); ok {
			if validationErr.Tag == "strong_password" {
				return c.JSON(http.StatusUnprocessableEntity, &models.Response{
					Code:    422,
					Message: "Password harus memiliki setidaknya 8 karakter",
					Status:  false,
				})
			}
		}

		if passUpdate.OldPassword == passUpdate.NewPassword {
			return c.JSON(http.StatusBadRequest, &models.Response{
				Code:    400,
				Message: "Password baru tidak boleh sama dengan password lama!",
				Status:  false,
			})
		}
		return c.JSON(http.StatusUnauthorized, &models.Response{
			Code:    401,
			Message: "Password lama salah!",
			Status:  false,
		})
	}

	return c.JSON(http.StatusOK, &models.Response{
		Code:    200,
		Message: "Password berhasil diubah! Silahkan login kembali!",
		Status:  true,
	})

}
