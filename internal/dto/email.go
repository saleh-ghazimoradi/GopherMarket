package dto

type Email struct {
	To      string
	Subject string
	Body    string
}

type PasswordResetEmailEvent struct {
	Email    string `json:"email"`
	ResetURL string `json:"reset_url"` // e.g. https://yourfrontend.com/reset-password
	Code     string `json:"code"`      // 8‑character code
}
