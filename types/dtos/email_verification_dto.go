package dtos

type EmailVerificationDTO struct {
	Email string `json:"email"`
	Otp   string `json:"otp"`
}
