package jsonapi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Optional[T any] struct {
	Value T
	Set   bool
}

func (o Optional[T]) ValOk() (interface{}, bool) {
	return o.Value, o.Set
}

func Value[T any](val T) Optional[T] {
	return Optional[T]{
		Value: val,
		Set:   true,
	}
}

type jsonFieldMapper interface {
	jsonFieldMap(map[string]json.RawMessage) error
}

type jsonFieldMapperDirect interface {
	fieldMap() (map[string]json.RawMessage, error)
}

func toJsonFieldMap(object interface{}) (map[string]json.RawMessage, error) {
	m := make(map[string]json.RawMessage)
	err := jsonFieldMap(object, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func jsonFieldMap(object interface{}, m map[string]json.RawMessage) error {

	if jfm, ok := object.(jsonFieldMapper); ok {
		return jfm.jsonFieldMap(m)
	}

	if jfmd, ok := object.(jsonFieldMapperDirect); ok {
		fields, err := jfmd.fieldMap()
		if err != nil {
			return err
		}
		for k, v := range fields {
			m[k] = v
		}
		return nil
	}

	if _, ok := object.(json.Marshaler); ok {
		return fmt.Errorf("%T implements json.Marshaler, it should implement jsonFieldMapper", object)
	}

	return jsonFieldMapFromStructFields(object, m)
}

func jsonFieldMapFromStructFields(object interface{}, m map[string]json.RawMessage) error {

	val := reflect.ValueOf(object)
	if val.Kind() != reflect.Struct {
		existingRaw, ok := object.(map[string]json.RawMessage)
		if ok {
			for k, v := range existingRaw {
				m[k] = v
			}
			return nil
		}
		existingInterface, ok := object.(map[string]interface{})
		if ok {
			for k, v := range existingInterface {
				asJSON, err := json.Marshal(v)
				if err != nil {
					return err
				}
				m[k] = json.RawMessage(asJSON)
			}
			return nil
		}
		return fmt.Errorf("object must be a struct, got %s %T", val.Kind().String(), object)
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if field.Anonymous {
			err := jsonFieldMap(val.Field(i).Interface(), m)
			if err != nil {
				return fmt.Errorf("anon field %s: %w", field.Name, err)
			}
			continue
		}

		tag := field.Tag.Get("json")
		if tag == "" {
			// maybe map lower case?
			continue
		}
		parts := strings.Split(tag, ",")
		name := parts[0]
		if name == "-" {
			continue
		}
		omitempty := false
		for _, part := range parts[1:] {
			if part == "omitempty" {
				omitempty = true
			}
		}

		if omitempty && val.Field(i).IsZero() {
			continue
		}
		iv := val.Field(i).Interface()
		if optional, ok := iv.(interface{ ValOk() (interface{}, bool) }); ok {
			val, isSet := optional.ValOk()
			if !isSet {
				continue
			}
			iv = val
		}

		asJSON, err := json.Marshal(iv)
		if err != nil {
			return err
		}

		m[name] = json.RawMessage(asJSON)
	}

	return nil
}
