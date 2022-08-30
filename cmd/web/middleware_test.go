package main

import (
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	var myH myHandler
	h := NoSurf(&myH)
	switch h.(type) {
	case http.Handler:
	default:
		t.Error("Type is not http.Handler")

	}
}

func TestSessionLoad(t *testing.T) {
	var myH myHandler
	h := SessionLoad(&myH)
	switch h.(type) {
	case http.Handler:
	default:
		t.Error("Type is not http.Handler")
	}
}

func TestAuth(t *testing.T) {
	var myH myHandler
	h := Auth(&myH)
	switch h.(type) {
	case http.Handler:
	default:
		t.Error("Type is not http.Handler")
	}
}
