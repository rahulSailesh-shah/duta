package slack

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rahulSailesh-shah/duta/internal/web"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/slack", func(r chi.Router) {
		r.Post("/events", h.handleEvents)
	})
}

func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	body, err := web.ReadBody(r)
	if err != nil {
		web.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	isValid := h.svc.VerifySignature(r.Header.Get("X-Slack-Request-Timestamp"),
		r.Header.Get("X-Slack-Signature"),
		body)
	if !isValid {
		web.Error(w, http.StatusUnauthorized, "Invalid signature")
		return
	}

	envelope, err := parseEnvelope(body)
	if err != nil {
		web.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if envelope.Type == "url_verification" {
		web.JSON(w, http.StatusOK, map[string]string{"challenge": envelope.Challenge})
		return
	}

	if shouldAct := h.svc.ShouldAct(&envelope.Event); !shouldAct {
		web.JSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	eventData, err := json.MarshalIndent(envelope.Event, "", "  ")
	fmt.Println("Event:\n", string(eventData))

	web.JSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

func parseEnvelope(body []byte) (*Envelope, error) {
	var envelope Envelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}
