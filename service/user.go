package service

import (
	"urllite/store"
	"urllite/types"
)

type userService struct {
	store store.Store
}

type UserService interface {
	Create(user *types.User) error
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

func (u userService) Create(user *types.User) error {
	return u.store.CreateUser(user)
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
