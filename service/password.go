package service

import (
	"net/http"
	"urllite/store"
	"urllite/types"

	"github.com/gocql/gocql"
	"golang.org/x/crypto/bcrypt"
)

type PasswordService interface {
	Create(password, user_id string) (*types.Password, *types.ApplicationError)
	GetPasswordByUserID(user_id string) (*types.Password, *types.ApplicationError)
	DeletePasswordByUserID(user_id string) *types.ApplicationError
	VerifyPassword(passwordStr string, password *types.Password) bool
}

type passwordService struct {
	store store.Store
}

func NewPasswordService() PasswordService {
	store := store.NewStore()
	return &passwordService{store: store}
}

func (s *passwordService) Create(password, user_id string) (*types.Password, *types.ApplicationError) {

	var newPassword types.Password
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to hash the password",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	newPassword.HashedPassword = string(hashed_password)
	userID, err := gocql.ParseUUID(user_id)
	newPassword.UserID = userID
	newPassword.Status = "active"
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Invalid user id",
			HttpStatusCode: http.StatusBadRequest,
			Err:            err,
		}
	}

	err = s.store.CreatePassword(&newPassword)
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to create password",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return &newPassword, nil
}

func (s *passwordService) GetPasswordByUserID(user_id string) (*types.Password, *types.ApplicationError) {
	password, err := s.store.GetPasswordByUserID(user_id)
	if err == gocql.ErrNotFound {
		return nil, &types.ApplicationError{
			Message:        "Password not found",
			HttpStatusCode: http.StatusNotFound,
			Err:            err,
		}
	}
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to find password",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return password, nil
}

func (s *passwordService) DeletePasswordByUserID(user_id string) *types.ApplicationError {
	password, err := s.store.GetPasswordByUserID(user_id)
	if err == gocql.ErrNotFound {
		return &types.ApplicationError{
			Message:        "Password not found",
			HttpStatusCode: http.StatusNotFound,
			Err:            err,
		}
	}
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to find password",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	err = s.store.DeletePassword(password)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to delete password",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return nil
}

func (s *passwordService) VerifyPassword(passwordStr string, password *types.Password) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password.HashedPassword), []byte(passwordStr))
	return err == nil
}
