package rvalid

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	instance *Validator
	once     sync.Once
)

type Validator struct {
	v1 *validator.Validate
	v2 *validator.Validate
}

func Instance() *Validator {
	once.Do(func() {
		v1 := validator.New()
		v1.SetTagName("v")

		v2 := validator.New()
		v2.SetTagName("v2")

		instance = &Validator{v1: v1, v2: v2}
	})
	return instance
}

func (v *Validator) RegisterValidation(tag string, fn validator.Func) (err error) {
	if err = v.v1.RegisterValidation(tag, fn); err != nil {
		return
	}
	return v.v2.RegisterValidation(tag, fn)
}

func (v *Validator) Validate(obj any) error {
	return v.v1.Struct(obj)
}

func (v *Validator) ValidateV2(obj any) error {
	return v.v2.Struct(obj)
}

func Validate(obj any) error {
	return Instance().Validate(obj)
}

func ValidateV2(obj any) error {
	return Instance().ValidateV2(obj)
}
