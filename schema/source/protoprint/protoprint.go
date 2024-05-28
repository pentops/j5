package protoprint

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pentops/jsonapi/builder/builder"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Options struct {
	PackagePrefixes []string
	OnlyFilenames   []string
}

func PrintProtoFiles(ctx context.Context, out builder.FS, src builder.Source, opts Options) error {

	sourceImg, err := src.SourceImage(ctx)
	if err != nil {
		return err
	}
	if len(opts.OnlyFilenames) == 0 {
		opts.OnlyFilenames = sourceImg.SourceFilenames
	}

	return printProtoFiles(ctx, out, sourceImg.File, opts)
}

func printProtoFiles(ctx context.Context, out builder.FS, files []*descriptorpb.FileDescriptorProto, opts Options) error {
	descriptors, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{
		File: files,
	})
	if err != nil {
		return err
	}

	var walkErr error

	foundExtensions := make([]protoreflect.ExtensionDescriptor, 0)

	descriptors.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		for i := 0; i < file.Extensions().Len(); i++ {
			foundExtensions = append(foundExtensions, file.Extensions().Get(i))
		}
		return true
	})

	descriptors.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		if len(opts.OnlyFilenames) > 0 {
			match := false
			for _, filename := range opts.OnlyFilenames {
				if file.Path() == filename {
					match = true
					break
				}
			}
			if !match {
				return true
			}
		}
		if len(opts.PackagePrefixes) > 0 {
			match := false
			pkg := string(file.Package())
			for _, prefix := range opts.PackagePrefixes {
				if strings.HasPrefix(pkg, prefix) {
					match = true
					break
				}
			}
			if !match {
				return true
			}
		}

		fileData, err := printFile(file, foundExtensions)
		if err != nil {
			walkErr = fmt.Errorf("in file %s: %w", file.Path(), err)
			return false
		}

		if err := out.Put(ctx, file.Path(), bytes.NewReader(fileData)); err != nil {
			walkErr = err
			return false
		}

		return true

	})
	if walkErr != nil {
		return walkErr
	}

	return nil

}

type fileBuffer struct {
	out    *bytes.Buffer
	addGap bool
	exts   map[protoreflect.FullName]map[protoreflect.FieldNumber]protoreflect.ExtensionDescriptor
}

func (fb *fileBuffer) findExtension(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionDescriptor, error) {
	if xt, ok := fb.exts[message][field]; ok {
		return xt, nil
	}
	return nil, fmt.Errorf("extension not found")
}

func (fb *fileBuffer) p(indent int, args ...interface{}) {
	if fb.addGap {
		fb.addGap = false
		fb.out.WriteString("\n")
	}
	fmt.Fprint(fb.out, strings.Repeat(" ", indent*2))
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			fmt.Fprint(fb.out, arg)
		case []string:
			for _, subArg := range arg {
				fmt.Fprint(fb.out, subArg)
			}
		default:
			fmt.Fprintf(fb.out, "%v", arg)
		}
	}
	fb.out.WriteString("\n")
}

type fileBuilder struct {
	out *fileBuffer
	ind int
}

func printFile(ff protoreflect.FileDescriptor, exts []protoreflect.ExtensionDescriptor) ([]byte, error) {

	extMap := map[protoreflect.FullName]map[protoreflect.FieldNumber]protoreflect.ExtensionDescriptor{}

	for _, ext := range exts {
		msgName := ext.ContainingMessage().FullName()
		if _, ok := extMap[msgName]; !ok {
			extMap[msgName] = make(map[protoreflect.FieldNumber]protoreflect.ExtensionDescriptor)
		}
		fieldNum := ext.Number()
		extMap[msgName][fieldNum] = ext
	}

	p := &fileBuilder{
		out: &fileBuffer{
			exts: extMap,
			out:  &bytes.Buffer{},
		},
	}
	return p.printFile(ff)
}

func (fb *fileBuilder) p(args ...interface{}) {
	fb.out.p(fb.ind, args...)
}

func commentLines(comment string) []string {
	if comment == "" {
		return nil
	}
	lines := strings.Split(comment, "\n")
	lines = lines[:len(lines)-1] // comment strings end with a newline
	for i, line := range lines {
		lines[i] = fmt.Sprintf("//%s", line)
	}
	return lines
}

func trailingComment(loc protoreflect.SourceLocation) []string {
	lines := strings.Split(loc.TrailingComments, "\n")
	lines = lines[:len(lines)-1] // comment strings end with a newline
	for i, line := range lines {
		lines[i] = fmt.Sprintf(" //%s", line)
	}
	return lines
}

func (fb *fileBuilder) leadingComments(loc protoreflect.SourceLocation) {
	for _, comment := range loc.LeadingDetachedComments {
		parts := commentLines(comment)
		for _, part := range parts {
			fb.p(part)
		}
		fb.addGap()
	}

	if loc.LeadingComments != "" {
		parts := commentLines(loc.LeadingComments)
		for _, part := range parts {
			fb.p(part)
		}
	}
}

func (fb *fileBuilder) addGap() {
	fb.out.addGap = true
}

func (fb *fileBuilder) endElem(end string) {
	// gaps should only occur between elements, not after the last one
	fb.out.addGap = false
	fb.p(end)
}

func (fb fileBuilder) indent() fileBuilder {
	return fileBuilder{out: fb.out, ind: fb.ind + 1}
}

func (fb *fileBuilder) printFile(ff protoreflect.FileDescriptor) ([]byte, error) {

	if ff.Syntax() != protoreflect.Proto3 {

		return nil, errors.New("only proto3 syntax is supported")
	}

	fb.p("syntax = \"proto3\";")
	fb.p()
	fb.p("package ", ff.Package(), ";")
	fb.p()
	imports := ff.Imports()
	for idx := 0; idx < imports.Len(); idx++ {
		dep := imports.Get(idx)
		// TODO: Sort
		fb.p("import \"", dep.Path(), "\";")
	}
	fb.p()
	// This could be manual iteration, but seemed more future-proof and
	// quicker to write.
	refl := ff.Options().ProtoReflect()
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
	fb.addGap()

	var elements = make(sourceElements, 0)

	messages := ff.Messages()
	for idx := 0; idx < messages.Len(); idx++ {
		elements.add(messages.Get(idx))
	}

	services := ff.Services()
	for idx := 0; idx < services.Len(); idx++ {
		elements.add(services.Get(idx))
	}

	enums := ff.Enums()
	for idx := 0; idx < enums.Len(); idx++ {
		elements.add(enums.Get(idx))
	}

	exts := ff.Extensions()
	for idx := 0; idx < exts.Len(); idx++ {
		elements.add(exts.Get(idx))
	}

	if err := fb.printElements(elements); err != nil {
		return nil, err
	}

	return fb.out.out.Bytes(), nil
}

func fieldTypeName(field protoreflect.FieldDescriptor) (string, error) {
	fieldType := field.Kind()

	var refElement protoreflect.Descriptor

	switch fieldType {
	case protoreflect.EnumKind:
		refElement = field.Enum()
	case protoreflect.MessageKind:
		refElement = field.Message()
	default:
		return fieldType.String(), nil
	}

	fieldMsg := field.Parent()

	return contextRefName(fieldMsg, refElement)
}

func contextRefName(contextOfCall protoreflect.Descriptor, refElement protoreflect.Descriptor) (string, error) {

	if contextOfCall.ParentFile().Package() != refElement.ParentFile().Package() {
		// if the thing the field references is in a different package, then the
		// full reference is used
		return string(refElement.FullName()), nil
	}

	refPath := pathToPackage(refElement)
	contextPath := pathToPackage(contextOfCall)

	for i := 0; i < len(contextPath); i++ {
		if len(refPath) == 0 || refPath[0] != contextPath[i] {
			break
		}
		refPath = refPath[1:]
	}

	return strings.Join(refPath, "."), nil
}

func pathToPackage(refElement protoreflect.Descriptor) []string {

	refPath := make([]string, 0)
	parentFileName := refElement.ParentFile().FullName()
	parent := refElement
	for parent.FullName() != parentFileName {
		refPath = append(refPath, string(parent.Name()))
		parent = parent.Parent()
	}

	stringsOut := make([]string, len(refPath))
	for i, part := range refPath {
		stringsOut[len(refPath)-i-1] = part
	}

	return stringsOut
}
