package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"urllite/store"
	"urllite/types"
	"urllite/types/dtos"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func UserAuthentication(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "No token found"})
		c.Abort()
		return
	}

	jwtKey := os.Getenv("ACCESS_TOKEN_SECRET_KEY")

	tokenString := strings.TrimSpace(authHeader[7:])
	claims := &dtos.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	if err != nil {
		// Token might be expired or malformed
		if errors.Is(err, jwt.ErrTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "Token expired"})
		} else if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "Invalid token"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "Invalid token"})
		}
		c.Abort()
		return
	}

	c.Set("current_username", claims.Username)
	c.Set("current_user_email", claims.Email)
	c.Set("current_user_id", claims.UserId)
	c.Set("current_user_role", claims.Role)
	c.Next()
}

func AdminAuthentication(c *gin.Context) {
	currentUserRole, ok := c.Get("current_user_role")
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "failed", "Message": "role not found in the context"})
		return
	}

	if currentUserRole.(string) == "admin" {
		c.Next()
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "failed", "Message": "You are not an admin"})
		return
	}
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
