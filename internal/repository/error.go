package repository

import "errors"

var (
	ErrsNotFound      = errors.New("record not found")
	ErrsAlreadyExists = errors.New("you can not register with this user")
)
