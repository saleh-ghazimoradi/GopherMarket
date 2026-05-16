package dto

type Email struct {
	To      string
	Subject string
	Body    string
}

type PasswordResetEmailEvent struct {
	Email     string `json:"email"`
	ResetLink string `json:"reset_link"`
}
