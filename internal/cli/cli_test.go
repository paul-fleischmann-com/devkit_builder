package cli

import (
	"io"
	"reflect"
	"testing"
)

func TestParseInterspersed(t *testing.T) {
	fs := NewFlagSet("t", io.Discard)
	commit := fs.String("commit", "", "")
	pos, err := Parse(fs, []string{"kit", "--commit", "abc", "other"})
	if err != nil {
		t.Fatal(err)
	}
	if *commit != "abc" || !reflect.DeepEqual(pos, []string{"kit", "other"}) {
		t.Errorf("commit=%q pos=%v", *commit, pos)
	}
}

func TestSrcFlagAlias(t *testing.T) {
	fs := NewFlagSet("t", io.Discard)
	src := SrcFlag(fs)
	if _, err := Parse(fs, []string{"--root", "/x"}); err != nil || *src != "/x" {
		t.Errorf("src=%q err=%v", *src, err)
	}
}
