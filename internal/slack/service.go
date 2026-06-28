package slack

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rahulSailesh-shah/duta/internal/config"
	"github.com/rahulSailesh-shah/duta/internal/workspace"
)

type Action int

const (
	ActionIgnore Action = iota
	ActionCreateWorkspace
	ActionAppendMessage
)

type Service struct {
	cfg  config.Config
	repo *workspace.Repo
}

func NewService(cfg config.Config, repo *workspace.Repo) *Service {
	return &Service{
		cfg:  cfg,
		repo: repo,
	}
}

func (s *Service) VerifySignature(timestamp, signature string, body []byte) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if d := time.Since(time.Unix(ts, 0)); d > 5*time.Minute || d < -5*time.Minute {
		return false
	}

	mac := hmac.New(sha256.New, []byte(s.cfg.SlackSigningSecret))
	mac.Write([]byte("v0:" + timestamp + ":"))
	mac.Write(body)
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (s *Service) ShouldAct(ctx context.Context, event *Event) (Action, error) {
	switch {
	case event.Type == "app_mention" && event.ThreadTS == "" && event.Ts != "":
		return ActionCreateWorkspace, nil
	case event.ThreadTS != "":
		ws, err := s.repo.GetWorkspace(ctx, event.Channel, event.ThreadTS)
		if err != nil {
			return ActionIgnore, err
		}
		if ws != nil {
			return ActionAppendMessage, nil
		}
	}
	return ActionIgnore, nil
}

func (s *Service) Handle(ctx context.Context, event *Event) (string, error) {
	action, err := s.ShouldAct(ctx, event)
	if err != nil {
		return "", err
	}

	switch action {
	case ActionCreateWorkspace:
		now := time.Now().UTC()
		created, err := s.repo.CreateWorkspace(ctx, workspace.Workspace{
			ID:        uuid.NewString(),
			Channel:   event.Channel,
			ThreadTS:  event.Ts,
			Status:    workspace.StatusQueued,
			RootUser:  event.User,
			RootText:  event.Text,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return "", err
		}
		if !created {
			return "ignored", nil
		}
		return "created", nil

	case ActionAppendMessage:
		created, err := s.repo.AppendMessage(ctx, event.Channel, event.ThreadTS, workspace.Message{
			Role:        workspace.RoleUser,
			Author:      event.User,
			Text:        event.Text,
			SlackTS:     event.Ts,
			ClientMsgID: event.ClientMsgID,
			CreatedAt:   time.Now().UTC(),
		})
		if err != nil {
			return "", err
		}
		if !created {
			return "ignored", nil
		}
		return "appended", nil

	default:
		return "ignored", nil
	}
}
