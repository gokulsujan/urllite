package dtos

type PasswordChangeDTO struct {
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}
