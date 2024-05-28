package protoprint

import (
	"fmt"
	"math"
	"math/bits"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type parsedOption struct {
	extName          string
	subField         []string
	valueLines       []string
	oneLine          bool
	lineInSrc        int32
	inlineWithParent bool
}

func (po parsedOption) fullType() string {
	if len(po.subField) == 0 {
		return fmt.Sprintf("(%s)", po.extName)
	}
	return fmt.Sprintf("(%s).%s", po.extName, strings.Join(po.subField, "."))
}

var maxExtDepth = map[protoreflect.FullName]int{
	"google.api.http": 0,
	//"buf.validate.field": 1,
}

func parseOption(ext extensionDef) *parsedOption {
	singleLine := false
	srcLoc := ext.locs
	var startLine int32
	if len(srcLoc) == 1 {
		startLine = srcLoc[0].Span[0]
		var endLine int32
		if len(srcLoc[0].Span) == 3 {
			endLine = startLine
		} else {
			endLine = srcLoc[0].Span[2]
		}
		if startLine == endLine {
			singleLine = true
		}
	}

	parentLocation := ext.parent.ParentFile().SourceLocations().ByDescriptor(ext.parent)
	inlineWithPareht := singleLine && parentLocation.StartLine == int(startLine)

	encoderDesc := ext.desc
	encoderVal := ext.val
	var namePrefix []string

	simplify := true //singleLine

	// convention seems to dictate these are always specified as
	// (google.api.http) = { ... }
	// even if it's just a get.
	if ext.desc.FullName() == "google.api.http" {
		simplify = false
	}

	printAsScalar := false
	maxDepth, ok := maxExtDepth[ext.desc.FullName()]
	if !ok {
		maxDepth = 5
	}
	for depth := 0; depth < maxDepth; depth++ {
		if !simplify {
			break
		}
		if encoderDesc.Kind() != protoreflect.MessageKind {
			printAsScalar = true
			break
		}
		encoderMessageDesc := encoderDesc.Message()
		encoderMessageVal := encoderVal.Message()
		descFields := encoderMessageDesc.Fields()
		definedFields := make([]protoreflect.FieldDescriptor, 0, descFields.Len())
		for i := 0; i < descFields.Len(); i++ {
			field := descFields.Get(i)
			if encoderMessageVal.Has(field) {
				definedFields = append(definedFields, field)
			}
		}

		if len(definedFields) != 1 {
			break
		}

		field := definedFields[0]

		if field.Kind() != protoreflect.MessageKind {
			// encoderVal/desc describe a message with a signle scalar field.
			printAsScalar = true
			break

		}

		namePrefix = append(namePrefix, string(field.Name()))

		encoderDesc = field
		encoderVal = encoderMessageVal.Get(field)

	}

	encMsg := encoderVal.Message()
	fields := mapOptionFields(encMsg)

	extName := string(ext.desc.FullName())
	if ext.parent.ParentFile().Package() == ext.desc.ParentFile().Package() {
		extName = string(ext.desc.Name())
	}

	if len(fields) == 0 {
		return &parsedOption{
			extName:          extName,
			subField:         namePrefix,
			valueLines:       []string{"{}"},
			oneLine:          true,
			lineInSrc:        startLine,
			inlineWithParent: inlineWithPareht,
		}
	}

	if len(fields) == 1 {
		field := fields[0]
		keyStr := string(field.field.Name())

		valStr, ok := marshalSingular(field.field, field.val)
		if ok {
			if printAsScalar {
				return &parsedOption{
					extName:          extName,
					subField:         append(namePrefix, keyStr),
					valueLines:       []string{valStr},
					oneLine:          true,
					lineInSrc:        startLine,
					inlineWithParent: inlineWithPareht,
				}

			}

			if singleLine {
				return &parsedOption{
					extName:          extName,
					subField:         namePrefix,
					valueLines:       []string{fmt.Sprintf("{%s: %s}", keyStr, valStr)},
					oneLine:          true,
					lineInSrc:        startLine,
					inlineWithParent: inlineWithPareht,
				}
			}
		}
	}

	linesOut := extensionJsonToText(encoderVal.Message())

	return &parsedOption{
		extName:    extName,
		subField:   namePrefix,
		valueLines: linesOut,
		lineInSrc:  startLine,
	}
}

type extFieldPair struct {
	key   string
	field protoreflect.FieldDescriptor
	val   protoreflect.Value
}

func mapOptionFields(msgVal protoreflect.Message) []extFieldPair {
	fields := msgVal.Descriptor().Fields()
	pairs := make([]extFieldPair, 0, fields.Len())
	for idx := 0; idx < fields.Len(); idx++ {
		fieldRefl := fields.Get(idx)
		if !msgVal.Has(fieldRefl) {
			continue
		}
		val := msgVal.Get(fieldRefl)

		pair := extFieldPair{
			key:   string(fieldRefl.Name()),
			val:   val,
			field: fieldRefl,
		}
		pairs = append(pairs, pair)
	}
	return pairs
}

func extensionJsonToText(msgVal protoreflect.Message) []string {

	fields := mapOptionFields(msgVal)

	linesOut := make([]string, 0, len(fields))

	for _, pair := range fields {

		if pair.field.IsList() {
			list := pair.val.List()
			listLen := list.Len()
			if listLen == 1 {
				scalar, ok := marshalSingular(pair.field, list.Get(0))
				if ok {
					nLine := fmt.Sprintf("%s: [%s]", pair.key, scalar)
					linesOut = append(linesOut, nLine)
					continue
				}
			}
			nLine := fmt.Sprintf("%s: [", pair.key)
			linesOut = append(linesOut, nLine)
			for i := 0; i < listLen; i++ {
				val := list.Get(i)
				subComma := ","
				if i == listLen-1 {
					subComma = ""
				}

				extField := extensionField(pair.field, val)
				for idx, line := range extField {
					if idx == len(extField)-1 {
						line = fmt.Sprintf("%s%s", line, subComma)
					}
					linesOut = append(linesOut, fmt.Sprintf("  %s", line))
				}
			}

			linesOut = append(linesOut, "]")
			continue
		}

		extField := extensionField(pair.field, pair.val)
		for idx, line := range extField {
			if idx == 0 {
				line = fmt.Sprintf("%s: %s", pair.key, line)
			}
			linesOut = append(linesOut, line)
		}

	}

	return linesOut

}

func extensionField(field protoreflect.FieldDescriptor, val protoreflect.Value) []string {
	linesOut := make([]string, 0, 1)
	scalar, ok := marshalSingular(field, val)
	if ok {
		nLine := scalar
		linesOut = append(linesOut, nLine)
		return linesOut
	}

	switch field.Kind() {
	case protoreflect.MessageKind:

		nLine := "{"
		linesOut = append(linesOut, nLine)
		msgLines := extensionJsonToText(val.Message())
		for _, line := range msgLines {
			linesOut = append(linesOut, fmt.Sprintf("  %s", line))
		}
		linesOut = append(linesOut, "}")

	default:
		panic(fmt.Sprintf("unexpected type %s", field.Kind().GoString()))
	}
	return linesOut
}

type extensionDef struct {
	desc   protoreflect.FieldDescriptor
	val    protoreflect.Value
	locs   []*descriptorpb.SourceCodeInfo_Location
	parent protoreflect.Descriptor
}

func subLocations(locs []*descriptorpb.SourceCodeInfo_Location, pathRoot []int32) []*descriptorpb.SourceCodeInfo_Location {
	var filtered []*descriptorpb.SourceCodeInfo_Location
	for _, loc := range locs {
		if !isPrefix(pathRoot, loc.Path) {
			continue
		}
		subPath := loc.Path[len(pathRoot):]
		filtered = append(filtered, &descriptorpb.SourceCodeInfo_Location{
			Path:                    subPath,
			Span:                    loc.Span,
			LeadingComments:         loc.LeadingComments,
			TrailingComments:        loc.TrailingComments,
			LeadingDetachedComments: loc.LeadingDetachedComments,
		})
	}
	return filtered
}

func isPrefix(prefix, path []int32) bool {
	if len(prefix) > len(path) {
		return false
	}
	for i, p := range prefix {
		if p != path[i] {
			return false
		}
	}
	return true
}

func (fb *fileBuilder) collectExtensions(parent protoreflect.Descriptor) ([]*parsedOption, error) {

	srcReflect := parent.Options().ProtoReflect()
	options := make([]*parsedOption, 0)

	// The reflection PB doesn't seem to give a way to get the source location
	// of the place the option was defined. This filters down all of the
	// locations in the parent object to the 'option' ones (7). Each field of
	// the option is then its own location.
	parentFile := parent.ParentFile()
	sourceLoc := protodesc.ToFileDescriptorProto(parentFile).SourceCodeInfo

	var optionsLocs []*descriptorpb.SourceCodeInfo_Location
	if sourceLoc != nil {
		parentRoot := parentFile.SourceLocations().ByDescriptor(parent)
		// options are at different indexes depending on the wrapper type.

		var optionFieldNumberInParent int32

		// field 7 of the Message type is the options field
		// see google/protobuf/descriptor.proto
		switch parent.(type) {
		case protoreflect.MessageDescriptor:
			optionFieldNumberInParent = 7
		case protoreflect.FieldDescriptor:
			optionFieldNumberInParent = 8
		case protoreflect.MethodDescriptor:
			optionFieldNumberInParent = 4
		case protoreflect.ServiceDescriptor:
			optionFieldNumberInParent = 3
		case protoreflect.EnumDescriptor:
			optionFieldNumberInParent = 3
		case protoreflect.EnumValueDescriptor:
			optionFieldNumberInParent = 3
		case protoreflect.OneofDescriptor:
			optionFieldNumberInParent = 2
		default:
			return nil, fmt.Errorf("unsupported parent type %T", parent)
		}

		parentPath := append(parentRoot.Path, optionFieldNumberInParent)
		optionsLocs = subLocations(sourceLoc.Location, parentPath)
	}

	var rangeErr error
	srcReflect.Range(func(desc protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		if !desc.IsExtension() {
			return true
		}

		num := desc.Number()
		optionLocs := subLocations(optionsLocs, []int32{int32(num)})

		parsed := parseOption(extensionDef{
			desc:   desc,
			val:    val,
			locs:   optionLocs,
			parent: parent,
		})
		options = append(options, parsed)

		return true
	})

	unknown := srcReflect.GetUnknown()
	if unknown != nil {
		b := unknown
		for len(b) > 0 {
			fNumber, fType, n := protowire.ConsumeTag(b)
			b = b[n:]

			optionLocs := subLocations(optionsLocs, []int32{int32(fNumber)})

			if fType != protowire.BytesType {
				return nil, fmt.Errorf("unknown field type %d", fType)
			}

			raw, n := protowire.ConsumeBytes(b)
			b = b[n:]
			parentName := srcReflect.Descriptor().FullName()

			serviceExt, err := fb.out.findExtension(parentName, fNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to find extension: %w", err)
			}

			// TODO: This assumes all extensions are messages
			extMsg := serviceExt.Message()

			dynamicExt := dynamicpb.NewMessage(extMsg)
			if err := proto.Unmarshal(raw, dynamicExt); err != nil {
				return nil, fmt.Errorf("failed to unmarshal extension: %w", err)
			}

			parsed := parseOption(extensionDef{
				desc:   serviceExt,
				val:    protoreflect.ValueOfMessage(dynamicExt),
				locs:   optionLocs,
				parent: parent,
			})

			options = append(options, parsed)

		}
	}

	return options, rangeErr
}

func (extInd *fileBuilder) printOption(parsed *parsedOption) {
	if parsed.oneLine {
		extInd.p("option ", parsed.fullType(), " = ", parsed.valueLines[0], ";")
		return
	}

	extInd.p("option ", parsed.fullType(), " = {")
	ind := extInd.indent()
	for _, line := range parsed.valueLines {
		ind.p(line)
	}
	extInd.endElem("};")
}

type optionsByLine []*parsedOption

func (o optionsByLine) Len() int {
	return len(o)
}

func (o optionsByLine) Less(i, j int) bool {
	return o[i].lineInSrc < o[j].lineInSrc
}

func (o optionsByLine) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (fb *fileBuilder) printFieldOptions(options []*parsedOption) {
	sort.Sort(optionsByLine(options))
	for idx, parsed := range options {
		hasMore := idx < len(options)-1
		suffix := ""
		if hasMore {
			suffix = ","
		}

		if len(parsed.valueLines) == 1 {
			fb.p(parsed.fullType(), " = ", parsed.valueLines[0], suffix)
			continue
		}

		fb.p(parsed.fullType(), " = {")
		ind := fb.indent()
		for _, line := range parsed.valueLines {
			ind.p(line)
		}
		fb.endElem("}" + suffix)
	}
}

// adapted from prototext
func marshalSingular(fd protoreflect.FieldDescriptor, val protoreflect.Value) (string, bool) {
	kind := fd.Kind()
	switch kind {
	case protoreflect.BoolKind:
		if val.Bool() {
			return "true", true
		} else {
			return "false", true
		}

	case protoreflect.StringKind:
		return prototextString(val.String()), true

	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		return strconv.FormatInt(val.Int(), 10), true

	case protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		return strconv.FormatUint(val.Uint(), 10), true

	case protoreflect.FloatKind:
		return fFloat(val.Float(), 32), true

	case protoreflect.DoubleKind:
		return fFloat(val.Float(), 64), true

	case protoreflect.BytesKind:
		return prototextString(string(val.Bytes())), true

	case protoreflect.EnumKind:
		num := val.Enum()
		if desc := fd.Enum().Values().ByNumber(num); desc != nil {
			return prototextString(string(desc.Name())), true
		} else {
			// Use numeric value if there is no enum description.
			return strconv.FormatInt(int64(num), 10), true
		}

	default:
		return "", false
	}
}

func fFloat(n float64, bitSize int) string {
	switch {
	case math.IsNaN(n):
		return "nan"
	case math.IsInf(n, +1):
		return "inf"
	case math.IsInf(n, -1):
		return "-inf"
	default:
		return strconv.FormatFloat(n, 'g', -1, bitSize)
	}
}

func prototextString(in string) string {
	outputASCII := true
	out := make([]byte, 0, len(in)+2)
	out = append(out, '"')
	i := indexNeedEscapeInString(in)
	in, out = in[i:], append(out, in[:i]...)
	for len(in) > 0 {
		switch r, n := utf8.DecodeRuneInString(in); {
		case r == utf8.RuneError && n == 1:
			// We do not report invalid UTF-8 because strings in the text format
			// are used to represent both the proto string and bytes type.
			r = rune(in[0])
			fallthrough
		case r < ' ' || r == '"' || r == '\\' || r == 0x7f:
			out = append(out, '\\')
			switch r {
			case '"', '\\':
				out = append(out, byte(r))
			case '\n':
				out = append(out, 'n')
			case '\r':
				out = append(out, 'r')
			case '\t':
				out = append(out, 't')
			default:
				out = append(out, 'x')
				out = append(out, "00"[1+(bits.Len32(uint32(r))-1)/4:]...)
				out = strconv.AppendUint(out, uint64(r), 16)
			}
			in = in[n:]
		case r >= utf8.RuneSelf && (outputASCII || r <= 0x009f):
			out = append(out, '\\')
			if r <= math.MaxUint16 {
				out = append(out, 'u')
				out = append(out, "0000"[1+(bits.Len32(uint32(r))-1)/4:]...)
				out = strconv.AppendUint(out, uint64(r), 16)
			} else {
				out = append(out, 'U')
				out = append(out, "00000000"[1+(bits.Len32(uint32(r))-1)/4:]...)
				out = strconv.AppendUint(out, uint64(r), 16)
			}
			in = in[n:]
		default:
			i := indexNeedEscapeInString(in[n:])
			in, out = in[n+i:], append(out, in[:n+i]...)
		}
	}
	out = append(out, '"')
	return string(out)
}

// indexNeedEscapeInString returns the index of the character that needs
// escaping. If no characters need escaping, this returns the input length.
func indexNeedEscapeInString(s string) int {
	for i := 0; i < len(s); i++ {
		if c := s[i]; c < ' ' || c == '"' || c == '\'' || c == '\\' || c >= 0x7f {
			return i
		}
	}
	return len(s)
}
