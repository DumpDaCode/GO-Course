package main

import (
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-course/bookings/internal/config"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig
	mux := routes(&app)
	switch mux.(type) {
	case *chi.Mux:
	default:
		t.Error("Type is not chi.Mux")
	}
}
