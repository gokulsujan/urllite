package service

import (
	"fmt"
	"net/http"
	"urllite/store"
	"urllite/types"
)

type userService struct {
	store store.Store
}

type UserService interface {
	Create(user *types.User) *types.ApplicationError
	GetUserByID(id string) (*types.User, error)
	GetUserByEmail(email string) (*types.User, error)
	GetUsers(types.UserFilter) ([]*types.User, error)
	UpdateUserByID(id string, user types.User) error
	DeleteUserByID(id string) error
}

func NewUserService() UserService {
	store := store.NewStore()
	return &userService{store: store}
}

func (u userService) Create(user *types.User) *types.ApplicationError {
	// Check for email id existence
	existingUser, err := u.store.GetUserByEmail(user.Email)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Error while checking for existing user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:          err,
		}
	}
	if existingUser != nil {
		return &types.ApplicationError{
			Message:        fmt.Sprintf("User with email %s already exists", user.Email),
			HttpStatusCode: http.StatusConflict,
			Err:          nil,
		}
	}

	// Create user
	err = u.store.CreateUser(user)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Error while creating user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:          err,
		}
	}

	return nil
}

func (u *userService) GetUserByID(id string) (*types.User, error) {

	return nil, nil
}

func (u *userService) GetUserByEmail(email string) (*types.User, error) {
	return nil, nil
}

func (u *userService) GetUsers(filter types.UserFilter) ([]*types.User, error) {
	return nil, nil
}

func (u *userService) UpdateUserByID(id string, user types.User) error {
	return nil
}

func (u *userService) DeleteUserByID(id string) error {
	return nil
}
