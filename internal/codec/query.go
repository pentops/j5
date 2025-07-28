package codec

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pentops/j5/lib/j5reflect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func propertyAtPath(root j5reflect.Root, path string) (j5reflect.Property, error) {
	parts := strings.Split(path, ".")
	pathParts, tail := parts[:len(parts)-1], parts[len(parts)-1]
	tail = strcase.ToLowerCamel(tail)
	for _, part := range pathParts {
		part = strcase.ToLowerCamel(part)
		prop, err := root.GetProperty(part)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unknown property %q", part))
		}
		var field j5reflect.Field
		if prop.IsSet() {
			field, err = prop.Field()
			if err != nil {
				return nil, err
			}
		} else {
			field, err = prop.Field()
			if err != nil {
				return nil, err
			}
		}

		if propSet, ok := field.AsContainer(); ok {
			root = propSet
			continue
		}
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("property %q is not a container", part))
	}
	return root.GetProperty(tail)
}

func (c *Codec) decodeQuery(queryString url.Values, msg protoreflect.Message) error {
	root, err := c.refl.NewRoot(msg)
	if err != nil {
		return err
	}

	for key, values := range queryString {
		prop, err := propertyAtPath(root, key)
		if err != nil {
			return err
		}

		field, err := prop.Field()
		if err != nil {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid query parameter %q", key))
		}

		if scalar, ok := field.AsScalar(); ok {
			if len(values) > 1 {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("multiple values provided for non-repeated field %q", key))
			}
			err = scalar.SetGoValue(values[0])
			if err != nil {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid value %q for field %q", values[0], key))
			}
			continue
		}

		if array, ok := field.AsArrayOfScalar(); ok {
			if len(values) == 1 && strings.HasPrefix(values[0], "[") {
				// json array
				val := values[0]
				items := make([]any, 0)
				err = json.Unmarshal([]byte(val), &items)
				if err != nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid value %q for array field %q", val, key))
				}
				for _, value := range items {
					_, err = array.AppendGoValue(value)
					if err != nil {
						return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid value %q for array field %q", value, key))
					}
				}
				continue
			}
			for _, value := range values {
				_, err = array.AppendGoValue(value)
				if err != nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid value %q for array field %q", value, key))
				}
			}
			continue
		}

		if array, ok := field.AsArrayOfContainer(); ok {
			if len(values) == 1 && strings.HasPrefix(values[0], "[") {
				// json array
				val := values[0]
				items := make([]json.RawMessage, 0)
				err = json.Unmarshal([]byte(val), &items)
				if err != nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid value %q for array field %q", val, key))
				}
				for _, value := range items {
					elem, _ := array.NewContainerElement()

					err = c.decodeRoot([]byte(value), elem)
					if err != nil {
						return err
					}
				}
				continue
			}

			for _, value := range values {
				elem, _ := array.NewContainerElement()

				err = c.decodeRoot([]byte(value), elem)
				if err != nil {
					return err
				}
			}
			continue
		}

		if container, ok := field.AsContainer(); ok {
			if len(values) > 1 {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("multiple values provided for non-repeated field %q", key))
			}
			val := strings.TrimSpace(values[0])
			if !strings.HasPrefix(val, "{") {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid value %q for field %q", val, key))
			}

			err = c.decodeRoot([]byte(val), container)
			if err != nil {
				return err
			}

			continue

		}

		return status.Error(codes.InvalidArgument, fmt.Sprintf("field %q is not supported for query %s", field.FullTypeName(), field.TypeName()))

	}

	return nil
}
