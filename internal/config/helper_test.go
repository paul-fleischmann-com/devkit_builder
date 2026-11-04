package config

import (
	"bufio"
	"strings"
)

func newReader(s string) *bufio.Reader { return bufio.NewReader(strings.NewReader(s)) }
