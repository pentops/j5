package j5validate

import (
	"fmt"

	"github.com/pentops/j5/lib/j5reflect"
)

var Global = NewValidator()

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(root j5reflect.Root) error {
	return fmt.Errorf("not implemented")
}
