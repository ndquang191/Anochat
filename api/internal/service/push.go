package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/model"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/ndquang191/Anochat/api/pkg/metrics"
)

const pushJobBuffer = 128

type PushSubscriptionInput struct {
	Endpoint string
	P256DH   string
	Auth     string
	Locale   string
}

type pushJob struct {
	UserID         uuid.UUID
	SubscriptionID uuid.UUID
}

type PushService struct {
	repo repository.PushSubscriptionRepository
	cfg  config.WebPushConfig
	jobs chan pushJob
}

func NewPushService(repo repository.PushSubscriptionRepository, cfg config.WebPushConfig) *PushService {
	return &PushService{repo: repo, cfg: cfg, jobs: make(chan pushJob, pushJobBuffer)}
}

func (s *PushService) Enabled() bool     { return s.cfg.Enabled }
func (s *PushService) PublicKey() string { return s.cfg.PublicKey }

func (s *PushService) Upsert(ctx context.Context, userID uuid.UUID, input PushSubscriptionInput) (*model.PushSubscription, error) {
	if !s.cfg.Enabled {
		return nil, fmt.Errorf("web push is disabled")
	}
	input = normalizePushInput(input.Endpoint, input.P256DH, input.Auth, input.Locale)
	parsed, err := url.ParseRequestURI(input.Endpoint)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || len(input.Endpoint) > 2048 ||
		input.P256DH == "" || len(input.P256DH) > 512 || input.Auth == "" || len(input.Auth) > 512 ||
		(input.Locale != "vi" && input.Locale != "en") {
		return nil, fmt.Errorf("invalid push subscription")
	}
	item := &model.PushSubscription{ID: uuid.New(), UserID: userID, Endpoint: input.Endpoint, P256DH: input.P256DH, Auth: input.Auth, Locale: input.Locale}
	if err := s.repo.Upsert(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *PushService) IsOwned(ctx context.Context, id, userID uuid.UUID) bool {
	if !s.cfg.Enabled {
		return false
	}
	_, err := s.repo.FindOwned(ctx, id, userID)
	return err == nil
}

func (s *PushService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.DeleteOwned(ctx, id, userID)
}

func (s *PushService) EnqueueMatch(userID, subscriptionID uuid.UUID) {
	if !s.cfg.Enabled || subscriptionID == uuid.Nil {
		return
	}
	select {
	case s.jobs <- pushJob{UserID: userID, SubscriptionID: subscriptionID}:
		metrics.PushEnqueued.Inc()
	default:
		metrics.PushDropped.Inc()
		slog.Warn("Push queue full; dropping match notification", "user_id", userID, "subscription_id", subscriptionID)
	}
}

func (s *PushService) Run(ctx context.Context) {
	if !s.cfg.Enabled {
		<-ctx.Done()
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-s.jobs:
			s.sendMatch(ctx, job)
		}
	}
}

func (s *PushService) sendMatch(parent context.Context, job pushJob) {
	ctx, cancel := context.WithTimeout(parent, 8*time.Second)
	defer cancel()
	item, err := s.repo.FindOwned(ctx, job.SubscriptionID, job.UserID)
	if err != nil {
		return
	}
	title, body := "Chat partner found", "Open AnoChat to start chatting."
	if item.Locale == "vi" {
		title, body = "Đã tìm thấy người trò chuyện", "Mở AnoChat để bắt đầu trò chuyện."
	}
	payload, _ := json.Marshal(map[string]string{"type": "match_found", "title": title, "body": body, "url": "/", "tag": "match-found"})
	response, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: item.Endpoint,
		Keys:     webpush.Keys{P256dh: item.P256DH, Auth: item.Auth},
	}, &webpush.Options{
		HTTPClient: &http.Client{Timeout: 8 * time.Second}, Subscriber: s.cfg.Subject,
		VAPIDPublicKey: s.cfg.PublicKey, VAPIDPrivateKey: s.cfg.PrivateKey,
		TTL: 300, Topic: "match-found", Urgency: webpush.UrgencyHigh,
	})
	if err != nil {
		metrics.PushSendTotal.WithLabelValues("transport_error").Inc()
		slog.Warn("Failed to send Web Push", "subscription_id", item.ID, "error", err)
		return
	}
	defer response.Body.Close()
	status := response.StatusCode
	switch {
	case status >= 200 && status < 300:
		metrics.PushSendTotal.WithLabelValues("success").Inc()
		_ = s.repo.MarkSuccess(ctx, item.ID, time.Now().UTC())
	case status == http.StatusNotFound || status == http.StatusGone:
		metrics.PushSendTotal.WithLabelValues("expired").Inc()
		_ = s.repo.DeleteByID(ctx, item.ID)
	default:
		metrics.PushSendTotal.WithLabelValues(fmt.Sprintf("http_%d", status)).Inc()
		slog.Warn("Web Push rejected", "subscription_id", item.ID, "status", status)
	}
}

func normalizePushInput(endpoint, p256dh, auth, locale string) PushSubscriptionInput {
	return PushSubscriptionInput{Endpoint: strings.TrimSpace(endpoint), P256DH: strings.TrimSpace(p256dh), Auth: strings.TrimSpace(auth), Locale: strings.TrimSpace(locale)}
}
