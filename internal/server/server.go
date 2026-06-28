package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rahulSailesh-shah/duta/internal/config"
	"github.com/rahulSailesh-shah/duta/internal/database"
	"github.com/rahulSailesh-shah/duta/internal/slack"
)

func New(cfg config.Config, db *database.DB) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// slack service and handler
	slackSvc := slack.NewService(cfg)
	slack.RegisterRoutes(r, slack.NewHandler(slackSvc))

	return r
}
