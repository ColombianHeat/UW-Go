package main

import (
	"bytes"
	"testing"
)

func TestGreet(t *testing.T) {
	buffer := bytes.Buffer{}
	Greet(&buffer, "Alan")

	got := buffer.String()
	want := "Hello, Alan"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}