package helper

type ValidateError struct {
	Message string
	Fields  map[string]string
}

func (v *ValidateError) Error() string {
	return v.Message
}

func (v *ValidateError) Extensions() map[string]any {
	return map[string]any{
		"validationErrors": v.Fields,
	}
}

func NewValidateError(v *Validator) *ValidateError {
	return &ValidateError{
		Message: "input validation failed",
		Fields:  v.Errors,
	}
}
