package patherr

import (
	"fmt"
	"strings"
)

type Error struct {
	Path []string
	Err  error
}

func New(err error, path ...string) *Error {
	return &Error{
		Path: path,
		Err:  err,
	}
}

func (v *Error) Error() string {
	return fmt.Sprintf("%s: %v", strings.Join(v.Path, "."), v.Err)
}

func (v *Error) Unwrap() error {
	return v.Err
}

func (v *Error) prefix(prefix ...string) *Error {
	return &Error{
		Path: append(prefix, v.Path...),
		Err:  v.Err,
	}
}

func Wrap(err error, path ...string) error {
	valErr, ok := err.(*Error)
	if !ok {
		return New(err, path...)
	}
	return valErr.prefix(path...)
}
