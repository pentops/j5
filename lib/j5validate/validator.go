package j5validate

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/j5types/date_j5t"
	"github.com/pentops/j5/j5types/decimal_j5t"
	"github.com/pentops/j5/lib/id62"
	"github.com/pentops/j5/lib/j5reflect"
	"github.com/pentops/j5/lib/j5schema"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var Global = NewValidator()

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(root j5reflect.Root) error {
	e, err := v.validateRoot(root)
	if err != nil {
		return err
	}
	if len(e) == 0 {
		return nil
	}
	return e
}

type Errors []Error

func (e Errors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	msg := "validation errors:\n"
	for _, err := range e {
		msg += fmt.Sprintf("- %s: %s\n", err.clientPath, err.Message)
	}
	return msg
}

func (e *Errors) mergeAt(path string, subErrors Errors) {
	if len(subErrors) == 0 {
		return
	}
	newErrors := make(Errors, 0, len(subErrors))
	for _, err := range subErrors {
		err.clientPath = append([]string{path}, err.clientPath...)
		newErrors = append(newErrors, err)
	}
	*e = append(*e, newErrors...)
}

type Error struct {
	clientPath []string
	Message    string
}

func (e Error) JSONPath() string {
	if len(e.clientPath) == 0 {
		return "-"
	}
	out := strings.Join(e.clientPath, ".")
	return strings.ReplaceAll(out, ".[", "[")

}

func (v *Validator) validateRoot(root j5reflect.Root) (Errors, error) {
	switch elem := root.(type) {
	case j5reflect.Object:
		e, _, err := v.validatePropSet(elem)
		return e, err

	case j5reflect.Oneof:
		e, _, err := v.validatePropSet(elem)
		return e, err

	default:
		return nil, fmt.Errorf("unsupported root schema type %T", elem)
	}
}

func (v *Validator) validatePropSet(ps j5reflect.PropertySet) (Errors, int, error) {
	var errs Errors
	count := 0
	err := ps.RangeProperties(func(prop j5reflect.Property) error {
		schema := prop.Schema()

		if !prop.IsSet() {
			if schema.Required {
				errs = append(errs, Error{
					clientPath: []string{schema.JSONName},
					Message:    "required field is not set",
				})
			}

			switch st := schema.Schema.(type) {
			case *j5schema.ArrayField:
				if st.Rules != nil && st.Rules.MinItems != nil && *st.Rules.MinItems > 0 {
					errs = append(errs, Error{
						clientPath: []string{schema.JSONName},
						Message:    fmt.Sprintf("array field %s is required to have at least %d items, but is not set", schema.JSONName, *st.Rules.MinItems),
					})
				}
			}

			return nil
		}

		field, err := prop.Field()
		if err != nil {
			return err
		}
		count++

		validationErr, err := v.validateField(field, schema.Schema)
		if err != nil {
			return fmt.Errorf("error validating field %s: %w", schema.JSONName, err)
		}

		errs.mergeAt(schema.JSONName, validationErr)
		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("error validating properties: %w", err)
	}
	return errs, count, nil

}

func (v *Validator) validateField(field j5reflect.Field, schema j5schema.FieldSchema) (Errors, error) {

	switch st := schema.(type) {
	case *j5schema.ObjectField:

		obj, ok := field.AsObject()
		if !ok {
			return nil, fmt.Errorf("expected object field, got %T", field)
		}

		errs, subCount, err := v.validatePropSet(obj)
		if err != nil {
			return nil, err
		}

		if st.Rules != nil {
			if st.Rules.MinProperties != nil {
				if subCount < int(*st.Rules.MinProperties) {
					errs = append(errs, Error{
						Message: fmt.Sprintf("minimum properties %d not met, got %d", *st.Rules.MinProperties, subCount),
					})
				}
			}
			if st.Rules.MaxProperties != nil {
				if subCount > int(*st.Rules.MaxProperties) {
					errs = append(errs, Error{
						Message: fmt.Sprintf("maximum properties %d exceeded, got %d", *st.Rules.MaxProperties, subCount),
					})
				}
			}

		}

		return errs, nil

	case *j5schema.OneofField:
		obj, ok := field.AsOneof()
		if !ok {
			return nil, fmt.Errorf("expected oneof field, got %T", field)
		}

		errs, subCount, err := v.validatePropSet(obj)
		if err != nil {
			return nil, err
		}

		if subCount > 1 {
			errs = append(errs, Error{
				Message: "oneof field has multiple types set, expected only one",
			})
		}

		return errs, nil

	case *j5schema.AnyField:
		// OK, no rules for any, however sub-validation is not being run.
		return nil, nil

	case *j5schema.PolymorphField:
		// OK, no rules for any, however sub-validation is not being run.
		return nil, nil

	case *j5schema.ArrayField:

		arrayField, ok := field.AsArray()
		if !ok {
			return nil, fmt.Errorf("expected array field  got %T", field)
		}

		errs := Errors{}

		count := 0
		err := arrayField.RangeValues(func(idx int, item j5reflect.Field) error {
			count++
			subErr, err := v.validateField(item, st.ItemSchema)
			if err != nil {
				return err
			}
			errs.mergeAt(fmt.Sprintf("[%d]", idx), subErr)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error validating array field: %w", err)
		}

		if st.Rules != nil {
			if st.Rules.MinItems != nil {
				if count < int(*st.Rules.MinItems) {
					errs = append(errs, Error{
						Message: fmt.Sprintf("minimum items %d not met, got %d", *st.Rules.MinItems, count),
					})
				}
			}
			if st.Rules.MaxItems != nil {
				if count > int(*st.Rules.MaxItems) {
					errs = append(errs, Error{
						Message: fmt.Sprintf("maximum items %d exceeded, got %d", *st.Rules.MaxItems, count),
					})
				}
			}
		}

		return errs, nil

	case *j5schema.MapField:

		mapField, ok := field.AsMap()
		if !ok {
			return nil, fmt.Errorf("expected map field, got %T", field)
		}

		errs := Errors{}
		count := 0
		err := mapField.Range(func(key string, item j5reflect.Field) error {
			count++
			subErr, err := v.validateField(item, st.ItemSchema)
			if err != nil {
				return fmt.Errorf("error validating map field %s: %w", key, err)
			}
			errs.mergeAt(key, subErr)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error validating map field: %w", err)
		}

		if st.Rules != nil {
			if st.Rules.MinPairs != nil {
				if count < int(*st.Rules.MinPairs) {
					errs = append(errs, Error{
						Message: fmt.Sprintf("minimum items %d not met, got %d", *st.Rules.MinPairs, count),
					})
				}
			}
			if st.Rules.MaxPairs != nil {
				if count > int(*st.Rules.MaxPairs) {
					errs = append(errs, Error{
						Message: fmt.Sprintf("maximum items %d exceeded, got %d", *st.Rules.MaxPairs, count),
					})
				}
			}
		}

		return errs, nil

	case *j5schema.EnumField:

		enumField, ok := field.AsEnum()
		if !ok {
			return nil, fmt.Errorf("expected enum field, got %T", field)
		}

		value, err := enumField.GetValue()
		if err != nil {
			return nil, err
		}
		gotValueStr := value.Name()

		errs := Errors{}
		if st.Rules != nil {
			if st.Rules.In != nil {
				if !slices.Contains(st.Rules.In, gotValueStr) {
					errs = append(errs, Error{
						Message: fmt.Sprintf("enum value %s is not in allowed values %v", gotValueStr, st.Rules.In),
					})
				}
			}
			if st.Rules.NotIn != nil {
				if slices.Contains(st.Rules.NotIn, gotValueStr) {
					errs = append(errs, Error{
						Message: fmt.Sprintf("enum value %s is in disallowed values %v", gotValueStr, st.Rules.NotIn),
					})
				}
			}
		}

		return errs, nil

	case *j5schema.ScalarSchema:

		scalarField, ok := field.AsScalar()
		if !ok {
			return nil, fmt.Errorf("expected scalar field, got %T", field)
		}

		gotValue, err := scalarField.ToGoValue()
		if err != nil {
			return nil, fmt.Errorf("error getting scalar value: %w", err)
		}

		return validateScalar(gotValue, st)
	default:
		return nil, fmt.Errorf("unsupported schema type %T", st)

	}
}

func validateScalar(gotValue any, schema *j5schema.ScalarSchema) (Errors, error) {
	j5Schema := schema.ToJ5Field()
	switch st := j5Schema.Type.(type) {
	case *schema_j5pb.Field_String_:
		if st.String_.Rules == nil {
			return nil, nil
		}
		gotString, ok := gotValue.(string)
		if !ok {
			if gotValue == nil {
				gotString = ""
			} else {
				return nil, fmt.Errorf("expected string value for scalar field, got %T", gotValue)
			}
		}
		rules := st.String_.Rules
		if rules.MinLength != nil {
			if len(gotString) < int(*st.String_.Rules.MinLength) {
				return Errors{{
					Message: fmt.Sprintf("minimum length %d not met, got %d", *st.String_.Rules.MinLength, len(gotString)),
				}}, nil
			}
		}
		if rules.MaxLength != nil {
			if len(gotString) > int(*st.String_.Rules.MaxLength) {
				return Errors{{
					Message: fmt.Sprintf("maximum length %d exceeded, got %d", *st.String_.Rules.MaxLength, len(gotString)),
				}}, nil
			}
		}

		if rules.Pattern != nil {
			pattern, err := regexp.Compile(*rules.Pattern)
			if err != nil {
				return nil, fmt.Errorf("error compiling pattern %q: %w", *rules.Pattern, err)
			}
			matched := pattern.MatchString(gotString)
			if !matched {
				return Errors{{
					Message: fmt.Sprintf("string %q does not match pattern %q", gotString, *rules.Pattern),
				}}, nil
			}
		}

		return nil, nil

	case *schema_j5pb.Field_Bool:
		if st.Bool.Rules == nil {
			return nil, nil
		}

		gotBool, ok := gotValue.(bool)
		if !ok {
			if gotValue != nil {
				// default bool value is false
				return nil, fmt.Errorf("expected boolean value for scalar field, got %T", gotValue)
			}
		}

		if st.Bool.Rules.Const != nil {
			if *st.Bool.Rules.Const != gotBool {
				return Errors{{
					Message: fmt.Sprintf("boolean value must be %t (got %v)", *st.Bool.Rules.Const, gotValue),
				}}, nil
			}
		}

		return nil, nil

	case *schema_j5pb.Field_Bytes:
		if st.Bytes.Rules == nil {
			return nil, nil
		}

		gotBytes, ok := gotValue.([]byte)
		if !ok {
			if gotValue == nil {
				gotBytes = []byte{}
			} else {
				return nil, fmt.Errorf("expected bytes value for scalar field, got %T", gotValue)
			}
		}

		if st.Bytes.Rules.MinLength != nil {
			if len(gotBytes) < int(*st.Bytes.Rules.MinLength) {
				return Errors{{
					Message: fmt.Sprintf("minimum length %d not met, got %d", *st.Bytes.Rules.MinLength, len(gotBytes)),
				}}, nil
			}
		}
		if st.Bytes.Rules.MaxLength != nil {
			if len(gotBytes) > int(*st.Bytes.Rules.MaxLength) {
				return Errors{{
					Message: fmt.Sprintf("maximum length %d exceeded, got %d", *st.Bytes.Rules.MaxLength, len(gotBytes)),
				}}, nil
			}
		}

	case *schema_j5pb.Field_Date:
		if st.Date.Rules == nil {
			return nil, nil
		}

		gotDate, ok := gotValue.(*date_j5t.Date)
		if !ok {
			if gotValue == nil {
				gotDate = date_j5t.NewDate(0, 0, 0) // default date value
			} else {
				return nil, fmt.Errorf("expected date value for scalar field, got %T", gotValue)
			}
		}

		if st.Date.Rules.Minimum != nil {
			mustMin, err := date_j5t.DateFromString(*st.Date.Rules.Minimum)
			if err != nil {
				return nil, fmt.Errorf("error parsing minimum date %q: %w", *st.Date.Rules.Minimum, err)
			}
			if st.Date.Rules.ExclusiveMinimum != nil && *st.Date.Rules.ExclusiveMinimum {
				if gotDate.Before(mustMin) {
					return Errors{{
						Message: fmt.Sprintf("date %s is before exclusive minimum %s", gotDate, mustMin),
					}}, nil
				}
			} else {
				if gotDate.Before(mustMin) || gotDate.Equals(mustMin) {
					return Errors{{
						Message: fmt.Sprintf("date %s is before or equal to minimum %s", gotDate, mustMin),
					}}, nil
				}
			}
		}

		if st.Date.Rules.Maximum != nil {
			mustMax, err := date_j5t.DateFromString(*st.Date.Rules.Maximum)
			if err != nil {
				return nil, fmt.Errorf("error parsing maximum date %q: %w", *st.Date.Rules.Maximum, err)
			}
			if st.Date.Rules.ExclusiveMaximum != nil && *st.Date.Rules.ExclusiveMaximum {
				if gotDate.After(mustMax) {
					return Errors{{
						Message: fmt.Sprintf("date %s is after exclusive maximum %s", gotDate, mustMax),
					}}, nil
				}
			} else {
				if gotDate.After(mustMax) || gotDate.Equals(mustMax) {
					return Errors{{
						Message: fmt.Sprintf("date %s is after or equal to maximum %s", gotDate, mustMax),
					}}, nil
				}
			}

		}

		return nil, nil

	case *schema_j5pb.Field_Timestamp:
		if st.Timestamp.Rules == nil {
			return nil, nil
		}

		gotTimestamp, ok := gotValue.(*timestamppb.Timestamp)
		if !ok {
			if gotValue == nil {
				gotTimestamp = timestamppb.New(time.Time{}) // default timestamp value
			} else {
				return nil, fmt.Errorf("expected timestamp value for scalar field, got %T", gotValue)
			}
		}

		timeVal := gotTimestamp.AsTime()
		if st.Timestamp.Rules.Minimum != nil {
			if st.Timestamp.Rules.ExclusiveMinimum != nil && *st.Timestamp.Rules.ExclusiveMinimum {
				if timeVal.Before(st.Timestamp.Rules.Minimum.AsTime()) {
					return Errors{{
						Message: fmt.Sprintf("timestamp %s is before exclusive minimum %s", gotTimestamp, st.Timestamp.Rules.Minimum),
					}}, nil
				}
			} else {
				if timeVal.Before(st.Timestamp.Rules.Minimum.AsTime()) || timeVal.Equal(st.Timestamp.Rules.Minimum.AsTime()) {
					return Errors{{
						Message: fmt.Sprintf("timestamp %s is before or equal to minimum %s", gotTimestamp, st.Timestamp.Rules.Minimum),
					}}, nil
				}
			}
		}

		if st.Timestamp.Rules.Maximum != nil {
			if st.Timestamp.Rules.ExclusiveMaximum != nil && *st.Timestamp.Rules.ExclusiveMaximum {
				if timeVal.After(st.Timestamp.Rules.Maximum.AsTime()) {
					return Errors{{
						Message: fmt.Sprintf("timestamp %s is after exclusive maximum %s", gotTimestamp, st.Timestamp.Rules.Maximum),
					}}, nil
				}
			} else {
				if timeVal.After(st.Timestamp.Rules.Maximum.AsTime()) || timeVal.Equal(st.Timestamp.Rules.Maximum.AsTime()) {
					return Errors{{
						Message: fmt.Sprintf("timestamp %s is after or equal to maximum %s", gotTimestamp, st.Timestamp.Rules.Maximum),
					}}, nil
				}
			}
		}

		return nil, nil

	case *schema_j5pb.Field_Decimal:
		if st.Decimal.Rules == nil {
			return nil, nil
		}

		gotDecimalT, ok := gotValue.(*decimal_j5t.Decimal)
		if !ok {
			if gotValue == nil {
				gotDecimalT = decimal_j5t.Zero()
			} else {
				return nil, fmt.Errorf("expected decimal value for scalar field, got %T", gotValue)
			}
		}

		gotDecimal, err := gotDecimalT.ToShop()
		if err != nil {
			return nil, fmt.Errorf("error converting decimal value: %w", err)
		}

		if st.Decimal.Rules.Minimum != nil {
			mustMin, err := decimal.NewFromString(*st.Decimal.Rules.Minimum)
			if err != nil {
				return nil, fmt.Errorf("error parsing minimum decimal %q: %w", *st.Decimal.Rules.Minimum, err)
			}
			if st.Decimal.Rules.ExclusiveMinimum != nil && *st.Decimal.Rules.ExclusiveMinimum {
				if gotDecimal.LessThan(mustMin) {
					return Errors{{
						Message: fmt.Sprintf("decimal %s is less than exclusive minimum %s", gotDecimal, mustMin),
					}}, nil
				}
			} else {
				if gotDecimal.LessThan(mustMin) || gotDecimal.Equal(mustMin) {
					return Errors{{
						Message: fmt.Sprintf("decimal %s is less than or equal to minimum %s", gotDecimal, mustMin),
					}}, nil
				}
			}
		}

		if st.Decimal.Rules.Maximum != nil {
			mustMax, err := decimal.NewFromString(*st.Decimal.Rules.Maximum)
			if err != nil {
				return nil, fmt.Errorf("error parsing maximum decimal %q: %w", *st.Decimal.Rules.Maximum, err)
			}
			if st.Decimal.Rules.ExclusiveMaximum != nil && *st.Decimal.Rules.ExclusiveMaximum {
				if gotDecimal.GreaterThan(mustMax) {
					return Errors{{
						Message: fmt.Sprintf("decimal %s is greater than exclusive maximum %s", gotDecimal, mustMax),
					}}, nil
				}
			} else {
				if gotDecimal.GreaterThan(mustMax) || gotDecimal.Equal(mustMax) {
					return Errors{{
						Message: fmt.Sprintf("decimal %s is greater than or equal to maximum %s", gotDecimal, mustMax),
					}}, nil
				}
			}
		}

	case *schema_j5pb.Field_Float:
		if st.Float.Rules == nil {
			return nil, nil
		}

		var val64 float64

		switch gotValue := gotValue.(type) {
		case float32:
			val64 = float64(gotValue)
		case float64:
			val64 = gotValue
		case nil:
			val64 = 0.0 // default float value
		default:
			return nil, fmt.Errorf("expected float value for scalar field, got %T", gotValue)
		}

		if st.Float.Rules.Minimum != nil {
			if st.Float.Rules.ExclusiveMinimum != nil && *st.Float.Rules.ExclusiveMinimum {
				if val64 <= *st.Float.Rules.Minimum {
					return Errors{{
						Message: fmt.Sprintf("float value %f is less than exclusive minimum %f", val64, *st.Float.Rules.Minimum),
					}}, nil
				}
			} else {
				if val64 < *st.Float.Rules.Minimum {
					return Errors{{
						Message: fmt.Sprintf("float value %f is less than minimum %f", val64, *st.Float.Rules.Minimum),
					}}, nil
				}
			}
		}

		if st.Float.Rules.Maximum != nil {
			if st.Float.Rules.ExclusiveMaximum != nil && *st.Float.Rules.ExclusiveMaximum {
				if val64 >= *st.Float.Rules.Maximum {
					return Errors{{
						Message: fmt.Sprintf("float value %f is greater than exclusive maximum %f", val64, *st.Float.Rules.Maximum),
					}}, nil
				}
			} else {
				if val64 > *st.Float.Rules.Maximum {
					return Errors{{
						Message: fmt.Sprintf("float value %f is greater than maximum %f", val64, *st.Float.Rules.Maximum),
					}}, nil
				}
			}
		}

	case *schema_j5pb.Field_Integer:
		if st.Integer.Rules == nil {
			return nil, nil
		}

		var val64 int64
		switch gotValue := gotValue.(type) {
		case int32:
			val64 = int64(gotValue)
		case int64:
			val64 = gotValue
		case uint64:
			val64 = int64(gotValue)
		case uint32:
			val64 = int64(gotValue)
		case nil:
			val64 = 0 // default integer value
		default:
			return nil, fmt.Errorf("expected integer value for scalar field, got %T", gotValue)
		}

		if st.Integer.Rules.Minimum != nil {
			if st.Integer.Rules.ExclusiveMinimum != nil && *st.Integer.Rules.ExclusiveMinimum {
				if val64 <= *st.Integer.Rules.Minimum {
					return Errors{{
						Message: fmt.Sprintf("integer value %d is less than exclusive minimum %d", val64, *st.Integer.Rules.Minimum),
					}}, nil
				}
			} else {
				if val64 < *st.Integer.Rules.Minimum {
					return Errors{{
						Message: fmt.Sprintf("integer value %d is less than minimum %d", val64, *st.Integer.Rules.Minimum),
					}}, nil
				}
			}
		}

		if st.Integer.Rules.Maximum != nil {
			if st.Integer.Rules.ExclusiveMaximum != nil && *st.Integer.Rules.ExclusiveMaximum {
				if val64 >= *st.Integer.Rules.Maximum {
					return Errors{{
						Message: fmt.Sprintf("integer value %d is greater than exclusive maximum %d", val64, *st.Integer.Rules.Maximum),
					}}, nil
				}
			} else {
				if val64 > *st.Integer.Rules.Maximum {
					return Errors{{
						Message: fmt.Sprintf("integer value %d is greater than maximum %d", val64, *st.Integer.Rules.Maximum),
					}}, nil
				}
			}
		}

		return nil, nil
	case *schema_j5pb.Field_Key:

		gotKey, ok := gotValue.(string)
		if !ok {
			if gotValue == nil {
				gotKey = "" // default key value
			} else {
				return nil, fmt.Errorf("expected key value for scalar field, got %T", gotValue)
			}
		}

		if st.Key.Format == nil {
			return nil, fmt.Errorf("key field has no format defined")
		}

		switch ft := st.Key.Format.Type.(type) {
		case *schema_j5pb.KeyFormat_Custom_:
			pattern, err := regexp.Compile(ft.Custom.Pattern)
			if err != nil {
				return nil, fmt.Errorf("error compiling custom key pattern %q: %w", ft.Custom.Pattern, err)
			}
			matched := pattern.MatchString(gotKey)
			if !matched {
				return Errors{{
					Message: fmt.Sprintf("key %q does not match custom pattern %q", gotKey, ft.Custom.Pattern),
				}}, nil
			}
		case *schema_j5pb.KeyFormat_Id62:
			_, err := id62.Parse(gotKey)
			if err != nil {
				return Errors{{
					Message: fmt.Sprintf("key %q is not a valid ID62 key: %v", gotKey, err),
				}}, nil
			}
		case *schema_j5pb.KeyFormat_Uuid:
			_, err := id62.ParseUUID(gotKey)
			if err != nil {
				return Errors{{
					Message: fmt.Sprintf("key %q is not a valid UUID key: %v", gotKey, err),
				}}, nil
			}

		case *schema_j5pb.KeyFormat_Informal_:
			// pass
		}

		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported scalar type %T", j5Schema.Type)
	}

	return nil, fmt.Errorf("scalar validation not implemented for type %T", j5Schema.Type)
}
