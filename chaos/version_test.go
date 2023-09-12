package main

import (
	"testing"
	"time"
)

func TestPrintVersion(t *testing.T) {
	tm, err := time.Parse("2006-01-02T15:04:05MST", "2021-08-18T10:56:57UTC")
	t.Log(tm.Local(), err)
}
