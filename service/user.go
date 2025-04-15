package service

import (
	"fmt"
	"net/http"
	"strings"
	"urllite/store"
	"urllite/types"

	"github.com/gocql/gocql"
)

type userService struct {
	store store.Store
}

type UserService interface {
	Create(user *types.User) *types.ApplicationError
	GetUserByID(id string) (*types.User, *types.ApplicationError)
	GetUserByEmail(email string) (*types.User, error)
	GetUsers(types.UserFilter) ([]*types.User, *types.ApplicationError)
	UpdateUserByID(id string, user types.User) *types.ApplicationError
	DeleteUserByID(id string) *types.ApplicationError
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

func (u *userService) GetUserByEmail(email string) (*types.User, error) {
	return u.store.GetUserByEmail(email)
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
