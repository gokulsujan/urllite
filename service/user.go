package service

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"urllite/store"
	"urllite/types"
	"urllite/types/dtos"

	"github.com/gocql/gocql"
	"github.com/golang-jwt/jwt/v5"
)

type userService struct {
	store store.Store
}

type UserService interface {
	Create(user *types.User) *types.ApplicationError
	GetUserByID(id string) (*types.User, *types.ApplicationError)
	GetUserByEmail(email string) (*types.User, *types.ApplicationError)
	GetUsers(types.UserFilter) ([]*types.User, *types.ApplicationError)
	UpdateUserByID(id string, user types.User) *types.ApplicationError
	DeleteUserByID(id string) *types.ApplicationError
	GenerateUserAccessToken(user *types.User) (string, *types.ApplicationError)
}

func NewUserService() UserService {
	store := store.NewStore()
	return &userService{store: store}
}

func (u userService) Create(user *types.User) *types.ApplicationError {
	// Check for email id existence
	existingUser, err := u.store.GetUserByEmail(user.Email)
	if err != nil && err != gocql.ErrNotFound {
		return &types.ApplicationError{
			Message:        "Error while checking for existing user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	if existingUser != nil {
		return &types.ApplicationError{
			Message:        fmt.Sprintf("User with email %s already exists", user.Email),
			HttpStatusCode: http.StatusConflict,
			Err:            nil,
		}
	}

	// Create user
	err = u.store.CreateUser(user)
	if err != nil && err != gocql.ErrNotFound {
		return &types.ApplicationError{
			Message:        "Error while creating user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return nil
}

func (u *userService) GetUserByID(id string) (*types.User, *types.ApplicationError) {
	user, err := u.store.GetUserByID(id)
	if err == gocql.ErrNotFound {
		return nil, &types.ApplicationError{
			Message:        "User not found ",
			HttpStatusCode: http.StatusNotFound,
			Err:            nil,
		}

	}
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to fing user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	return user, nil
}

func (u *userService) GetUserByEmail(email string) (*types.User, *types.ApplicationError) {
	user, err := u.store.GetUserByEmail(email)
	if err == gocql.ErrNotFound {
		return nil, &types.ApplicationError{
			Message:        "No user found",
			HttpStatusCode: http.StatusNotFound,
		}
	} else if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to find user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return user, nil
}

func (u *userService) GetUsers(filter types.UserFilter) ([]*types.User, *types.ApplicationError) {
	users, err := u.store.SearchUsers(filter)
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to search users",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	return users, nil
}

func (u *userService) UpdateUserByID(id string, user types.User) *types.ApplicationError {
	existingUser, err := u.store.GetUserByID(id)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to find user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	if existingUser == nil {
		return &types.ApplicationError{
			Message:        "No user found",
			HttpStatusCode: http.StatusNotFound,
		}
	}

	if strings.TrimSpace(user.Name) != "" {
		existingUser.Name = strings.TrimSpace(user.Name)
	}

	if strings.TrimSpace(user.Email) != "" {
		existingUser.Email = strings.TrimSpace(user.Email)
	}

	if strings.TrimSpace(user.Mobile) != "" {
		existingUser.Mobile = strings.TrimSpace(user.Mobile)
	}

	if strings.TrimSpace(user.VerifiedEmail) != "" {
		existingUser.VerifiedEmail = strings.TrimSpace(user.VerifiedEmail)
	}

	err = u.store.UpdateUser(existingUser)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to update user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	return nil
}

func (u *userService) DeleteUserByID(id string) *types.ApplicationError {
	existingUser, err := u.store.GetUserByID(id)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to find user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	if existingUser == nil {
		return &types.ApplicationError{
			Message:        "No user found",
			HttpStatusCode: http.StatusNotFound,
		}
	}

	err = u.store.DeleteUser(existingUser)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to delete the user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return nil
}

func (u *userService) GenerateUserAccessToken(user *types.User) (string, *types.ApplicationError) {
	claims := &dtos.JWTClaims{Username: user.Name, Email: user.Email, UserId: user.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(os.Getenv("ACCESS_TOKEN_SECRET_KEY"))
	accessToken, err := token.SignedString(jwtKey)
	if err != nil {
		return "", &types.ApplicationError{
			Message:        "Unable to generate token",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return accessToken, nil
}
