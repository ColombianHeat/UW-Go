package main

import (
	"testing"
)

func TestSayHello(t *testing.T) {
	got := sayHello()
	want := "Hello World!"

	if got != want {
		t.Errorf("expected '%q', but got '%q'", got, want)
	}
}