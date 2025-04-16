package dtos

type SignupDTO struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Mobile          string `json:"mobile"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}
