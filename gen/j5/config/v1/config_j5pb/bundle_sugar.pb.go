// Code generated by protoc-gen-go-sugar. DO NOT EDIT.

package config_j5pb

import (
	proto "google.golang.org/protobuf/proto"
)

// OutputType is a oneof wrapper
type OutputTypeKey string

const (
	Output_GoProxy OutputTypeKey = "goProxy"
)

func (x *OutputType) TypeKey() (OutputTypeKey, bool) {
	switch x.Type.(type) {
	case *OutputType_GoProxy_:
		return Output_GoProxy, true
	default:
		return "", false
	}
}

type IsOutputTypeWrappedType interface {
	TypeKey() OutputTypeKey
	proto.Message
}

func (x *OutputType) Set(val IsOutputTypeWrappedType) {
	switch v := val.(type) {
	case *OutputType_GoProxy:
		x.Type = &OutputType_GoProxy_{GoProxy: v}
	}
}
func (x *OutputType) Get() IsOutputTypeWrappedType {
	switch v := x.Type.(type) {
	case *OutputType_GoProxy_:
		return v.GoProxy
	default:
		return nil
	}
}
func (x *OutputType_GoProxy) TypeKey() OutputTypeKey {
	return Output_GoProxy
}

type IsOutputType_Type = isOutputType_Type
