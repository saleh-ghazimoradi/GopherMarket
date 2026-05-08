package helper

import (
	"net/url"
	"regexp"
	"slices"
	"strings"
)

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	// Initialize map on the fly if it's nil (safety first)
	if v.Errors == nil {
		v.Errors = make(map[string]string)
	}

	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, n int) bool {
	return len([]rune(value)) <= n
}

func MinChars(value string, n int) bool {
	return len([]rune(value)) >= n
}

func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]struct{}) // Using struct{} saves memory vs bool

	for _, value := range values {
		uniqueValues[value] = struct{}{}
	}

	return len(values) == len(uniqueValues)
}

func IsURL(value string) bool {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func NewValidator() *Validator {
	return &Validator{Errors: make(map[string]string)}
}
