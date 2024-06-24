package gogen

import (
	"strings"

	"github.com/iancoleman/strcase"
)

func goFieldName(jsonName string) string {
	return strcase.ToCamel(jsonName)
}

func goTypeName(name string) string {
	// Undescores are used to separate nested-scoped types, e.g. a message
	// defined within a message in proto, this function preserves the underscores
	// but fixes up any casing in between - which basically results in capatalizing
	// the first letter.
	parts := strings.Split(name, "_")
	for i, part := range parts {
		parts[i] = strcase.ToCamel(part)
	}
	return strings.Join(parts, "_")
}
