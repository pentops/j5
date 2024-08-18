package ast

import (
	"fmt"

	"github.com/pentops/j5/internal/j5lang/lexer"
	"github.com/pentops/j5/internal/patherr"
)

// Schema Walker is called for each block in the AST.
// Header first, then each assignment.
// Then for each block, BodySchema is called, which should return a new Schema
// then each block, then Done.
type Schema interface {
	Assign(name Reference, value Value) error
	Block(BlockHeader) (Schema, error)
	Done() error
}

func ApplySchema(file *File, schema Schema) error {
	posErrors, err := convertBody(file.Body, schema)
	if err != nil {
		return err
	}
	if len(posErrors) > 0 {
		return lexer.PositionErrors(posErrors)
	}
	return nil
}

func convertBody(body Body, schema Schema) ([]lexer.PositionError, error) {
	var err error

	for _, decl := range body.Statements {
		switch d := decl.(type) {
		case Assignment:
			err = schema.Assign(d.Key, d.Value)
			if err != nil {
				return nil, patherr.Wrap(err, "assignment", d.Key.String())
			}

		case BlockStatement:
			blockSchema, err := schema.Block(d.BlockHeader)
			if err != nil {
				return nil, patherr.Wrap(err, d.guessName())
			}
			if blockSchema != nil {
				_, err := convertBody(d.Body, blockSchema)
				if err != nil {
					return nil, patherr.Wrap(err, d.guessName())
				}
			}
			if err := blockSchema.Done(); err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("unexpected type at root %T", d)
		}
	}

	return nil, nil
}
