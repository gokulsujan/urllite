package dtos

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	Username string
	UserId   string
	Email    string
	Role     string
	jwt.RegisteredClaims
}
