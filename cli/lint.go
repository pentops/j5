package cli

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pentops/jsonapi/source"
	"google.golang.org/protobuf/types/descriptorpb"
)

// {namespace}.{version} is the root package
// {namespace}.{version}.service should have Service suffix
// {namespace}.{version}.topic should have Topic suffix
// {namespace}.{version}.event should have Event suffix
var rePackage = regexp.MustCompile(`(.*)\.v([^.]+)(\.[^.]+)?`)

type SubPackageType struct {
	ServiceSuffix  string
	RequestSuffix  string
	ResponseSuffix string
	ReturnsEmpty   bool
}

var suffixMap = map[string]SubPackageType{
	"": {
		ServiceSuffix: "",
	},
	".service": {
		ServiceSuffix:  "Service",
		ReturnsEmpty:   false,
		RequestSuffix:  "Request",
		ResponseSuffix: "Response",
	},
	".topic": {
		ServiceSuffix: "Topic",
		ReturnsEmpty:  true,
		RequestSuffix: "Message",
	},
	".events": {
		ServiceSuffix: "Events",
		ReturnsEmpty:  true,
		RequestSuffix: "Event",
	},
	".sandbox": {
		ServiceSuffix:  "Sandbox",
		ReturnsEmpty:   false,
		RequestSuffix:  "Request",
		ResponseSuffix: "Response",
	},
}

func runValidate(ctx context.Context, cfg struct {
	Source string `flag:"src" default:"." description:"Source directory containing jsonapi.yaml and buf.lock.yaml"`
}) error {

	img, err := source.ReadImageFromSourceDir(ctx, cfg.Source, false)
	if err != nil {
		return err
	}

	issues := &issueSet{}

	validateFile := func(fileSrc *descriptorpb.FileDescriptorProto) {
		file := newFile(fileSrc, issues)
		packageName := file.GetPackage()
		matches := rePackage.FindStringSubmatch(packageName)
		if len(matches) == 0 {
			file.Issuef("package name %s does not match expected format {namespace}.v{version}<.sub>", packageName)
			return
		}

		//prefix := matches[1]
		//version := "v" + matches[2]
		suffix := matches[3]
		packageType, ok := suffixMap[suffix]
		if !ok {
			file.Issuef("sub-package %s is not valid", suffix)
		}

		serviceLocations := map[string]int32{}
		for _, loc := range file.SourceCodeInfo.Location {
			path := loc.Path
			if len(path) == 0 {
				continue
			}
			parts := make([]string, len(path))
			for i, p := range path {
				parts[i] = fmt.Sprintf("%d", p)
			}
			serviceLocations[strings.Join(parts, ".")] = loc.Span[0]
		}

		for idx, service := range file.Service {
			serviceIssues := file.Location(6, int32(idx)).Name(service.GetName())

			if packageType.ServiceSuffix == "" {
				serviceIssues.Issuef("should be in a sub package")
				continue
			}

			if !strings.HasSuffix(service.GetName(), packageType.ServiceSuffix) {
				serviceIssues.Issuef("should have suffix %s", packageType.ServiceSuffix)
			}

			for idx, method := range service.Method {
				methodIssues := serviceIssues.Location(2, int32(idx)).Name(method.GetName())

				if method.InputType == nil || method.OutputType == nil {
					methodIssues.Issuef("should have input and output types")
					continue
				}
				if !strings.HasSuffix(*method.InputType, packageType.RequestSuffix) {
					methodIssues.Issuef("request '%s' should have suffix %s", *method.InputType, packageType.RequestSuffix)
				}

				if packageType.ReturnsEmpty {
					if *method.OutputType != "google.protobuf.Empty" {
						methodIssues.Issuef("response '%s' should be google.protobuf.Empty", *method.OutputType)
					}
				} else {
					if !strings.HasSuffix(*method.OutputType, packageType.ResponseSuffix) {
						methodIssues.Issuef("response '%s' should have suffix %s", *method.OutputType, packageType.ResponseSuffix)
					}
				}
			}
		}
	}

	for _, file := range img.File {
		validateFile(file)
	}

	for _, issue := range issues.issues {
		fmt.Printf("%s:%d:%d:%s\n", issue.Filenmae, issue.Location[0], issue.Location[1], issue.Problem)
	}

	if len(issues.issues) > 0 {
		return fmt.Errorf("validation failed")
	}

	return nil
}

type issueSet struct {
	issues []issue
}

func (is *issueSet) Add(ii issue) {
	is.issues = append(is.issues, ii)
}

type sourceFile struct {
	*descriptorpb.FileDescriptorProto
	issues *issueSet
}

func newFile(file *descriptorpb.FileDescriptorProto, issues *issueSet) *sourceFile {
	return &sourceFile{FileDescriptorProto: file, issues: issues}
}

func (ss *sourceFile) LocationOf(path []int32) *descriptorpb.SourceCodeInfo_Location {
	for _, loc := range ss.SourceCodeInfo.Location {
		if len(loc.Path) == len(path) {
			match := true
			for i, p := range path {
				if loc.Path[i] != p {
					match = false
					break
				}
			}
			if match {
				return loc
			}
		}
	}
	return nil
}

func (ss *sourceFile) Issuef(format string, args ...interface{}) {
	ii := issue{
		Filenmae: ss.GetName(),
		Problem:  fmt.Sprintf(format, args...),
	}
	ss.issues.Add(ii)
}

type location struct {
	sourceFile
	path   []int32
	issues *issueSet
	name   string
}

func (ss *sourceFile) Location(loc ...int32) *location {
	return &location{sourceFile: *ss, issues: ss.issues, path: loc}
}

func (l *location) Location(loc ...int32) *location {
	return &location{sourceFile: l.sourceFile, issues: l.issues, path: append(l.path, loc...)}
}

func (l *location) Name(name string) *location {
	prefix := ""
	if l.name != "" {
		prefix = l.name + "."
	}
	return &location{sourceFile: l.sourceFile, issues: l.issues, path: l.path, name: prefix + name}
}

func (l *location) Issuef(format string, args ...interface{}) {
	prefix := ""
	if l.name != "" {
		prefix = l.name + ": "
	}
	ii := issue{
		Filenmae: l.GetName(),
		Problem:  prefix + fmt.Sprintf(format, args...),
		Location: l.LocationOf(l.path).Span,
	}
	l.issues.Add(ii)
}

type issue struct {
	Filenmae string
	Location []int32
	Problem  string
}
