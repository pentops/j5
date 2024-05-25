package docker

import (
	"strings"
	"testing"

	"github.com/pentops/jsonapi/gen/j5/builder/v1/builder_j5pb"
)

func TestMapEnvVars(t *testing.T) {
	sampleIn := []string{"PROTOC_GEN_GO_MESSAGING_EXTRA_HEADERS=api-version:$GIT_HASH"}

	cInfo := builder_j5pb.CommitInfo{
		Hash: "abcdef",
	}

	r, err := mapEnvVars(sampleIn, &cInfo)
	if err != nil {
		t.Errorf("Received error in mapenvvars: %v", err.Error())
	}
	if len(r) != 1 {
		t.Error("Got len other than expected")
	}
	if strings.Count(r[0], "=") > 1 {
		t.Error("Too many equal signs in env var")
	}
	if r[0] != "PROTOC_GEN_GO_MESSAGING_EXTRA_HEADERS=api-version:abcdef" {
		t.Error("Output not correct for git hash substitution")
	}
}
