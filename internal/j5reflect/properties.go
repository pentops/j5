package j5reflect

import "github.com/pentops/j5/internal/j5schema"

type Property interface {
	JSONName() string
	Field
}

type fieldBase struct {
	schema *j5schema.ObjectProperty
}

func (f fieldBase) JSONName() string {
	return f.schema.JSONName
}

type objectProperty struct {
	*objectField
	fieldBase
}

type oneofProperty struct {
	*oneofField
	fieldBase
}

type enumProperty struct {
	*enumField
	fieldBase
}

type mutableArrayProperty struct {
	*mutableArrayField
	fieldBase
}

type leafArrayProperty struct {
	*leafArrayField
	fieldBase
}

type mutableMapProperty struct {
	*mutableMapField
	fieldBase
}

type leafMapProperty struct {
	*leafMapField
	fieldBase
}

type scalarProperty struct {
	*scalarField
	fieldBase
}
