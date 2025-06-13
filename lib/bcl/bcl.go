package bcl

import (
	"github.com/pentops/j5/gen/j5/bcl/v1/bcl_j5pb"
	"github.com/pentops/j5/internal/bcl"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Parser struct {
	impl *bcl.Parser
}

func NewParser(schemaSpec *bcl_j5pb.Schema) (*Parser, error) {
	p, err := bcl.NewParser(schemaSpec)
	if err != nil {
		return nil, err
	}
	return &Parser{
		impl: p,
	}, nil
}

func (p *Parser) ParseFile(filename string, data string, msg protoreflect.Message) (*bcl_j5pb.SourceLocation, error) {
	loc, err := p.impl.ParseFile(filename, data, msg)
	if err != nil {
		return nil, err
	}
	return loc, nil
}
