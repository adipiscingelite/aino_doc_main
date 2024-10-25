package utils

import (
	"github.com/golang-jwt/jwt/v5"
)

type JwtCustomClaims struct {
	IdUser string `json:"id_user"`
	jwt.RegisteredClaims
}
