package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/rahulSailesh-shah/duta/internal/config"
)

const (
	mentionedTTL  = 24 * time.Hour
	mentionedSize = 4096

	dedupeTTL  = time.Hour
	dedupeSize = 8192
)

type Service struct {
	cfg config.Config

	mu        sync.Mutex
	mentioned *lru.LRU[string, struct{}]
	seenMsgs  *lru.LRU[string, struct{}]
}

func NewService(cfg config.Config) *Service {
	return &Service{
		cfg:       cfg,
		mentioned: lru.NewLRU[string, struct{}](mentionedSize, nil, mentionedTTL),
		seenMsgs:  lru.NewLRU[string, struct{}](dedupeSize, nil, dedupeTTL),
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

func (s *Service) ShouldAct(event *Event) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	act := false
	switch {
	case event.Type == "app_mention":
		if event.ThreadTS == "" && event.Ts != "" {
			s.mentioned.Add(event.Ts, struct{}{})
		}
		act = true
	case event.Type == "message" && event.ThreadTS != "":
		_, act = s.mentioned.Get(event.ThreadTS)
	}

	if !act {
		return false
	}

	if id := event.ClientMsgID; id != "" {
		if _, seen := s.seenMsgs.Get(id); seen {
			return false
		}
		s.seenMsgs.Add(id, struct{}{})
	}
	return true
}
