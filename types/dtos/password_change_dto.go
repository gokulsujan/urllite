package dtos

type PasswordChangeDTO struct {
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}

type PasswordChangeUsingOtpDTO struct {
	Email    string `json:"email"`
	Otp      string `json:"otp"`
	Password string `json:"password"`
}
