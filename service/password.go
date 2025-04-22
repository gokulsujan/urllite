package service

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"urllite/store"
	"urllite/types"
	"urllite/utils"

	"github.com/gocql/gocql"
	"golang.org/x/crypto/bcrypt"
)

type PasswordService interface {
	Create(password, user_id string) (*types.Password, *types.ApplicationError)
	GetPasswordByUserID(user_id string) (*types.Password, *types.ApplicationError)
	DeletePasswordByUserID(user_id string) *types.ApplicationError
	VerifyPassword(passwordStr string, password *types.Password) bool
	ChangePassword(email, currentPassword, newPassword string) *types.ApplicationError
	SendForgetPasswordOtp(email string) *types.ApplicationError
	ChangePasswordUsingOtp(email, otp, newPassword string) *types.ApplicationError
	VerifyForgetPasswordOtp(email, otpStr string) *types.ApplicationError
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

func (s *passwordService) ChangePassword(email, currentPassword, newPassword string) *types.ApplicationError {
	user, err := s.store.GetUserByEmail(email)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Invalid user",
			HttpStatusCode: http.StatusBadRequest,
		}

	}

	password, err := s.store.GetPasswordByUserID(user.ID.String())
	if err != nil {
		if err == gocql.ErrNotFound {
			return &types.ApplicationError{
				Message:        "Password not availble for the user",
				HttpStatusCode: http.StatusNotFound,
			}
		}
		return &types.ApplicationError{
			Message:        "Unable to find password",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(password.HashedPassword), []byte(currentPassword))
	if err != nil {
		return &types.ApplicationError{
			Message:        "Incorrect password.",
			HttpStatusCode: http.StatusNotAcceptable,
		}

	}

	newHashedPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to change password",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	password.HashedPassword = string(newHashedPass)

	err = s.store.UpdatePassword(password)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to change password",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return nil
}

func (s *passwordService) SendForgetPasswordOtp(email string) *types.ApplicationError {
	user, err := s.store.GetUserByEmail(email)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to find user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	otp := types.Otp{UserID: user.ID, Key: "FORGET_PASS_" + email, Status: "pending", ExpiredAt: (time.Now().Add(10 * time.Minute))}
	rand.Seed(time.Now().UnixNano())
	otp.Otp = strconv.Itoa(rand.Intn(900000) + 100000)
	_, err = s.store.CreateOtp(&otp)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to create otp",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	mailer := utils.NewMailer()
	err = mailer.SendOtpForEmailVerification(user, &otp)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable sent email",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	return nil
}

func (s *passwordService) VerifyForgetPasswordOtp(email, otpStr string) *types.ApplicationError {
	user, err := s.store.GetUserByEmail(email)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to find user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	if user == nil {
		return &types.ApplicationError{
			Message:        "User not found",
			HttpStatusCode: http.StatusNotFound,
		}
	}

	// Getting otp
	key := "FORGET_PASS_" + user.Email
	otps, err := s.store.GetOtpByUserIdAndOtp(user.ID.String(), key, otpStr)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to find otp",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	if otps == nil {
		return &types.ApplicationError{
			Message:        "Not a valid otp",
			HttpStatusCode: http.StatusBadRequest,
		}

	}
	return nil
}

func (s *passwordService) ChangePasswordUsingOtp(email, otp, newPassword string) *types.ApplicationError {
	user, err := s.store.GetUserByEmail(email)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to find user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	if user == nil {
		return &types.ApplicationError{
			Message:        "No user found",
			HttpStatusCode: http.StatusUnprocessableEntity,
			Err:            err,
		}
	}

	otps, err := s.store.GetOtpByUserIdAndOtp(user.ID.String(), "FORGET_PASS_"+email, otp)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to find otp",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	if otps == nil {
		return &types.ApplicationError{
			Message:        "Invalid Otp",
			HttpStatusCode: http.StatusUnprocessableEntity,
			Err:            err,
		}
	}

	for _, otp := range otps {
		err = s.store.ChangeOtpStatus(otp, "verified")
		if err != nil {
			return &types.ApplicationError{
				Message:        "Unable to verify otp",
				HttpStatusCode: http.StatusInternalServerError,
				Err:            err,
			}
		}
	}

	password, appErr := s.GetPasswordByUserID(user.ID.String())
	if appErr != nil {
		return appErr
	}

	newHashedPass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	password.HashedPassword = string(newHashedPass)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to find password",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	err = s.store.UpdatePassword(password)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to update password",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	return nil
}
