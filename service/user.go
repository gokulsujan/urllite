package service

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"urllite/cache"
	"urllite/store"
	"urllite/types"
	"urllite/types/dtos"
	"urllite/utils"

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
	GenerateUserAccessToken(user *types.User, ctx context.Context) (string, *types.ApplicationError)
	SendEmailVerificationOtp(emailID string) *types.ApplicationError
	VerifyEmail(emailID, otpStr string) *types.ApplicationError
	MakeAdmin(user_id string) *types.ApplicationError
}

func NewUserService() UserService {
	store := store.NewStore()
	return &userService{store: store}
}

func (u userService) Create(user *types.User) *types.ApplicationError {
	if !utils.EmailValidation(user.Email) {
		return &types.ApplicationError{
			Message:        "Invalid Email ID",
			HttpStatusCode: http.StatusNotAcceptable,
		}
	}
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

func (u *userService) GenerateUserAccessToken(user *types.User, ctx context.Context) (string, *types.ApplicationError) {
	redisTokenKey := "access_token_" + user.ID.String()
	redicClient := cache.InitRedis(ctx)
	ok, err := redicClient.Exists(redisTokenKey)
	if err != nil {
		return "", &types.ApplicationError{
			Message:        "Unable to get token from redis",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	if ok {
		token, err := redicClient.Get(redisTokenKey)
		if err != nil {
			return "", &types.ApplicationError{
				Message:        "Unable to get token from redis",
				HttpStatusCode: http.StatusInternalServerError,
				Err:            err,
			}
		}

		return token, nil
	}

	claims := &dtos.JWTClaims{Username: user.Name, Email: user.Email, UserId: user.ID.String(), Role: user.Role,
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

	redicClient.Set(redisTokenKey, accessToken, 23*time.Hour)

	return accessToken, nil
}

func (u *userService) SendEmailVerificationOtp(emailID string) *types.ApplicationError {
	// Verify user
	user, appErr := u.GetUserByEmail(emailID)
	if appErr != nil {
		return appErr
	}

	if user == nil {
		return &types.ApplicationError{
			Message:        "User not found",
			HttpStatusCode: http.StatusNotFound,
		}
	}

	if user.IsEmailVerified() {
		return &types.ApplicationError{
			Message:        "User already verified",
			HttpStatusCode: http.StatusOK,
		}
	}

	// Otp generation
	var otp types.Otp
	otp.Key = "VERIFY_EMAIL_OTP_" + emailID
	rand.Seed(time.Now().UnixNano())
	otp.Otp = strconv.Itoa(rand.Intn(900000) + 100000)
	otp.ExpiredAt = time.Now().Add(10 * time.Minute)
	otp.UserID = user.ID
	otp.Status = "pending"
	u.store.CreateOtp(&otp)

	mailer := utils.NewMailer()
	err := mailer.SendOtpForEmailVerification(user, &otp)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable sent email",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	return nil
}

func (u *userService) VerifyEmail(emailID, otpStr string) *types.ApplicationError {
	// Verify user
	user, appErr := u.GetUserByEmail(emailID)
	if appErr != nil {
		return appErr
	}

	if user == nil {
		return &types.ApplicationError{
			Message:        "User not found",
			HttpStatusCode: http.StatusNotFound,
		}
	}

	if user.IsEmailVerified() {
		return &types.ApplicationError{
			Message:        "Email already verified",
			HttpStatusCode: http.StatusOK,
		}
	}

	// Getting otp
	key := "VERIFY_EMAIL_OTP_" + user.Email
	otps, err := u.store.GetOtpByUserIdAndOtp(user.ID.String(), key, otpStr)
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

	// Verify email
	user.VerifiedEmail = user.Email
	err = u.store.UpdateUser(user)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable update user verified email",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	appErr = u.UpdateUserByID(user.ID.String(), *user)
	if appErr != nil {
		return appErr
	}
	for _, otp := range otps {
		err = u.store.ChangeOtpStatus(otp, "verified")
		if err != nil {
			return &types.ApplicationError{
				Message:        "Unable verify otp",
				HttpStatusCode: http.StatusInternalServerError,
				Err:            err,
			}
		}
	}
	return nil

}

func (u *userService) MakeAdmin(user_id string) *types.ApplicationError {
	user, appErr := u.GetUserByID(user_id)
	if appErr != nil {
		return appErr
	}

	user.Role = "admin"
	err := u.store.UpdateUser(user)
	if err != nil {
		return &types.ApplicationError{
			Message:        "Unable to update user",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}

	return nil
}
