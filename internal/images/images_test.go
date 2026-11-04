package images

import (
	"bytes"
	"strings"
	"testing"
)

const prefix = "localhost/factory-devkit-"

func TestReport(t *testing.T) {
	listing := strings.Join([]string{
		prefix + "go|abc123def456|sha256:1111|2026-11-04 07:00|500 MB",
		"docker.io/library/debian|12|sha256:2222|2026-01-01|100 MB",
		"garbage line",
		prefix + "py|abc123def456|3333|2026-11-03 07:00|400 MB",
		"",
	}, "\n")
	var out bytes.Buffer
	report(&out, listing, prefix)
	got := out.String()
	for _, want := range []string{
		"kit=go commit=abc123def456 created=2026-11-04 07:00 size=500 MB",
		"FACTORY_TEST_PROFILES entry: go=1111",
		"FACTORY_TEST_PROFILES entry: py=3333",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "debian") || strings.Contains(got, "sha256:") {
		t.Errorf("foreign image or sha256 prefix leaked:\n%s", got)
	}
}

func TestReportNone(t *testing.T) {
	var out bytes.Buffer
	report(&out, "docker.io/x|1|sha256:1|now|1 MB\n", prefix)
	if !strings.Contains(out.String(), "no devkit image yet") {
		t.Errorf("got %q", out.String())
	}
}
