package server

import "github.com/go-playground/validator/v10"

type Validator struct {
	Validator *validator.Validate
}

func NewValidator() *Validator {
	val := validator.New()
	return &Validator{Validator: val}
}

func (v *Validator) Validate(i interface{}) error {
	return v.Validator.Struct(i)
}
