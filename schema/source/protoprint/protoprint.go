package protoprint

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pentops/jsonapi/builder/builder"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// This package surely exists elsewhere, but I can't find it. Hopefully it's a
// quick one.
// Confirmed as copilot basically wrote this itself...

type Options struct {
	PackagePrefixes []string
}

func PrintProtoFiles(ctx context.Context, out builder.FS, source builder.Source, opts Options) error {

	req, err := source.ProtoCodeGeneratorRequest(ctx, "./")
	if err != nil {
		return err
	}

	for _, ff := range req.GetProtoFile() {
		if len(opts.PackagePrefixes) > 0 {
			match := false
			pkg := ff.GetPackage()
			for _, prefix := range opts.PackagePrefixes {
				if strings.HasPrefix(pkg, prefix) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		fileData, err := printFile(ff)
		if err != nil {
			return err
		}

		if err := out.Put(ctx, ff.GetName(), bytes.NewReader(fileData)); err != nil {
			return err
		}

	}

	return nil

}

type fileBuilder struct {
	out *bytes.Buffer
	ind int
}

func printFile(ff *descriptorpb.FileDescriptorProto) ([]byte, error) {
	p := &fileBuilder{out: &bytes.Buffer{}}
	return p.printFile(ff)
}

func (fb *fileBuilder) p(args ...interface{}) {
	fmt.Fprint(fb.out, strings.Repeat(" ", fb.ind*2))
	for _, arg := range args {
		fmt.Fprintf(fb.out, "%v", arg)
	}
	fb.out.WriteString("\n")
}

func (fb fileBuilder) indent() fileBuilder {
	return fileBuilder{out: fb.out, ind: fb.ind + 1}
}

func (fb *fileBuilder) printFile(ff *descriptorpb.FileDescriptorProto) ([]byte, error) {

	if ff.GetSyntax() != "proto3" {
		return nil, errors.New("only proto3 syntax is supported")
	}

	fb.p("syntax = \"proto3\";")
	fb.p()
	fb.p("package ", ff.GetPackage(), ";")
	fb.p()
	for _, dep := range ff.GetDependency() {
		fb.p("import \"", dep, "\";")
	}
	fb.p()
	if ff.Options != nil {
		// This could be manual iteration, but seemed more future-proof and
		// quicker to write.
		refl := ff.Options.ProtoReflect()
		fields := refl.Descriptor().Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			if !refl.Has(field) {
				continue
			}
			switch field.Kind() {
			case protoreflect.BoolKind:
				fb.p("option ", field.Name(), " = ", refl.Get(field).Interface(), ";")
			case protoreflect.StringKind:
				fb.p("option ", field.Name(), " = \"", refl.Get(field).Interface(), "\";")
			}
		}
		fb.p()
	}

	for _, svc := range ff.GetService() {
		if err := fb.printService(ff.GetPackage(), svc); err != nil {
			return nil, err
		}
		fb.p()
	}

	for _, msg := range ff.GetMessageType() {
		if err := fb.printMessage(ff.GetPackage(), msg); err != nil {
			return nil, err
		}
	}

	return fb.out.Bytes(), nil
}

func (fb *fileBuilder) printService(currentPackage string, svc *descriptorpb.ServiceDescriptorProto) error {

	fb.p("service ", svc.GetName(), " {")
	ind := fb.indent()
	for idx, meth := range svc.GetMethod() {
		inputType, err := contextRefName(currentPackage, meth.GetInputType())
		if err != nil {
			return err
		}
		outputType, err := contextRefName(currentPackage, meth.GetOutputType())
		if err != nil {
			return err
		}

		type extensionDef struct {
			desc protoreflect.FieldDescriptor
			val  protoreflect.Value
		}

		extensions := make([]extensionDef, 0)

		meth.Options.ProtoReflect().Range(func(desc protoreflect.FieldDescriptor, val protoreflect.Value) bool {
			if !desc.IsExtension() {
				return true
			}
			extensions = append(extensions, extensionDef{
				desc: desc,
				val:  val,
			})

			return true
		})

		end := " {}"
		if len(extensions) > 0 {
			end = " {"
		}

		ind.p("rpc ", meth.GetName(), "(", inputType, ") returns (", outputType, ")", end)
		extInd := ind.indent()
		if len(extensions) > 0 {
			for _, ext := range extensions {
				valMsg := ext.val.Message()
				pm := valMsg.Interface()
				marshalled, err := prototext.MarshalOptions{
					Multiline: true,
					Indent:    "  ",
				}.Marshal(pm)
				if err != nil {
					return err
				}
				lines := strings.Split(strings.TrimSuffix(string(marshalled), "\n"), "\n")
				for i, line := range lines {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) != 2 {
						return fmt.Errorf("invalid extension line %q", line)
					}
					lines[i] = fmt.Sprintf("%s: %s", parts[0], strings.TrimSpace(parts[1]))
				}

				if len(lines) == 1 {
					extInd.p("option (", ext.desc.FullName(), ") = {", lines[0], "};")
				} else {
					extInd.p("option (", ext.desc.FullName(), ") = {")
					for _, line := range lines {
						extInd.p("  ", line)
					}
					extInd.p("};")
				}

			}
			ind.p("}")
		}

		if idx < len(svc.GetMethod())-1 {
			fb.p()
		}
	}
	fb.p("}")

	return nil
}

var typeNames = map[descriptorpb.FieldDescriptorProto_Type]string{
	descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:   "double",
	descriptorpb.FieldDescriptorProto_TYPE_FLOAT:    "float",
	descriptorpb.FieldDescriptorProto_TYPE_INT64:    "int64",
	descriptorpb.FieldDescriptorProto_TYPE_UINT64:   "uint64",
	descriptorpb.FieldDescriptorProto_TYPE_INT32:    "int32",
	descriptorpb.FieldDescriptorProto_TYPE_FIXED64:  "fixed64",
	descriptorpb.FieldDescriptorProto_TYPE_FIXED32:  "fixed32",
	descriptorpb.FieldDescriptorProto_TYPE_BOOL:     "bool",
	descriptorpb.FieldDescriptorProto_TYPE_STRING:   "string",
	descriptorpb.FieldDescriptorProto_TYPE_GROUP:    "group",
	descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:  "message",
	descriptorpb.FieldDescriptorProto_TYPE_BYTES:    "bytes",
	descriptorpb.FieldDescriptorProto_TYPE_UINT32:   "uint32",
	descriptorpb.FieldDescriptorProto_TYPE_ENUM:     "enum",
	descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: "sfixed32",
	descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: "sfixed64",
	descriptorpb.FieldDescriptorProto_TYPE_SINT32:   "sint32",
	descriptorpb.FieldDescriptorProto_TYPE_SINT64:   "sint64",
}

func (fb *fileBuilder) printMessage(currentPackage string, msg *descriptorpb.DescriptorProto) error {
	var err error
	fullName := fmt.Sprintf("%s.%s", currentPackage, msg.GetName())
	fb.p("message ", msg.GetName(), " {")
	ind := fb.indent()

	remainingNested := make([]*descriptorpb.DescriptorProto, 0, len(msg.GetNestedType()))
	mapNested := make(map[string]*descriptorpb.DescriptorProto, len(msg.GetNestedType()))

	for _, nested := range msg.GetNestedType() {
		if nested.GetOptions().GetMapEntry() {
			nextPrefix := fmt.Sprintf(".%s.%s", fullName, nested.GetName())
			mapNested[nextPrefix] = nested
		} else {
			remainingNested = append(remainingNested, nested)
		}
	}

	for _, field := range msg.GetField() {

		var typeName string
		var label string
		mapMessage, ok := mapNested[field.GetTypeName()]
		if ok {
			delete(mapNested, field.GetTypeName())
			keyTypeName, err := fieldTypeName(currentPackage, mapMessage.GetField()[0])
			if err != nil {
				return err
			}
			valueTypeName, err := fieldTypeName(currentPackage, mapMessage.GetField()[1])
			if err != nil {
				return err
			}
			typeName = fmt.Sprintf("map<%s, %s>", keyTypeName, valueTypeName)

		} else {
			typeName, err = fieldTypeName(currentPackage, field)
			if err != nil {
				return err
			}

			if field.Label != nil && *field.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
				label = "repeated "
			} else if field.Proto3Optional != nil && *field.Proto3Optional {
				label = "optional "
			}
		}
		ind.p(label, typeName, " ", field.GetName(), " = ", field.GetNumber(), ";")

	}
	for _, nested := range remainingNested {
		nextPrefix := fmt.Sprintf("%s.%s.", fullName, nested.GetName())
		if err := ind.printMessage(nextPrefix, nested); err != nil {

			return err
		}
	}
	// Return the first entry
	for prefix := range mapNested {
		return fmt.Errorf("map entry '%s' not used", prefix)
	}

	fb.p("}")
	return nil
}

func contextRefName(currentPackage string, ref string) (string, error) {
	contextPrefix := fmt.Sprintf(".%s.", currentPackage)
	if strings.HasPrefix(ref, contextPrefix) {
		return strings.TrimPrefix(ref, contextPrefix), nil
	}
	return strings.TrimPrefix(ref, "."), nil
}

func fieldTypeName(currentPackage string, field *descriptorpb.FieldDescriptorProto) (string, error) {
	fieldType := field.GetType()

	if fieldType == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
		if strings.HasPrefix(field.GetTypeName(), "."+currentPackage) {
			return strings.TrimPrefix(field.GetTypeName(), "."+currentPackage), nil
		}
		return strings.TrimPrefix(field.GetTypeName(), "."), nil
	}

	typeName, ok := typeNames[fieldType]
	if !ok {
		return "", fmt.Errorf("unknown field type %v", field.GetType())
	}
	return typeName, nil

}
