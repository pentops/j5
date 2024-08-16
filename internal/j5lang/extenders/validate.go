package extenders

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

type Value interface {
	AsString() string
	AsUint(bits int) (uint64, error)
}

type ValidateExtender struct {
	UnknownExtender
}

func (v ValidateExtender) FieldLevel(f *schema_j5pb.ObjectProperty, key string, value Value) (bool, error) {
	switch key {
	case "required":
		f.Required = true
		return true, nil

	default:
		return false, nil
	}
}

func (v ValidateExtender) String(f *schema_j5pb.StringField, key string, value Value) error {
	switch key {
	case "min_length":
		val, err := value.AsUint(64)
		if err != nil {
			return fmt.Errorf("error parsing min_length: %w", err)
		}
		f.Rules.MinLength = &val

	case "max_length":
		val, err := value.AsUint(64)
		if err != nil {
			return fmt.Errorf("error parsing max_length: %w", err)
		}
		f.Rules.MaxLength = &val

	case "pattern":
		val := value.AsString()
		f.Rules.Pattern = &val

	default:
		return fmt.Errorf("unexpected key (S) %s", key)
	}

	return nil
}
