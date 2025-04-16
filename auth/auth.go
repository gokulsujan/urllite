package auth

import (
	"net/http"
	"os"
	"urllite/store"
	"urllite/types"
	"urllite/types/dtos"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func UserAuthentication(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	jwtKey := os.Getenv("ACCESS_TOKEN_SECRET_KEY")

	tokenString := authHeader[7:]
	claims := &dtos.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	c.Set("current_username", claims.Username)
	c.Set("current_user_email", claims.Email)
	c.Set("current_user_id", claims.UserId)
	c.Next()
}

func CurrentUserFromContext(c *gin.Context) *types.User {
	store := store.NewStore()

	userIDFromContext, ok := c.Get("current_user_id")
	if !ok {
		return nil
	}

	user, _ := store.GetUserByID(string(userIDFromContext.(string)))
	return user
}
