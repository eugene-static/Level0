package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/eugene-static/Level0/app/internal/models"
	"github.com/eugene-static/Level0/app/lib/config"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

type Service interface {
	Set(ctx context.Context, order *models.Order) error
}

type Streaming struct {
	cfg     *config.Nats
	l       *slog.Logger
	service Service
}

func New(cfg *config.Nats, l *slog.Logger, service Service) *Streaming {
	return &Streaming{
		cfg:     cfg,
		l:       l,
		service: service,
	}
}

func (s *Streaming) Connect(ctx context.Context) (stan.Subscription, error) {
	sc, err := stan.Connect(s.cfg.ClusterID, s.cfg.ClientID, stan.NatsURL(s.cfg.URL))
	if err != nil {
		s.l.Error("failed to connect to stream",
			slog.String("cluster ID", s.cfg.ClusterID),
			slog.String("client ID", s.cfg.ClientID),
			slog.String("URL", s.cfg.URL),
			slog.Any("details", err))
		return nil, err
	}
	nctx := nats.Context(ctx)
	sub, err := sc.Subscribe(s.cfg.Subject,
		func(m *stan.Msg) {
			if err = s.handler(nctx, m); err != nil {
				s.l.Error("failed while receiving msg", slog.Any("details", err))
				return
			}
			s.l.Info("data received successfully")
		})
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *Streaming) handler(ctx context.Context, m *stan.Msg) error {
	key, err := validate(m.Data)
	if err != nil {
		return err
	}
	order := &models.Order{
		UID:  key,
		Data: m.Data,
	}
	if err = s.service.Set(ctx, order); err != nil {
		return err
	}
	return nil
}

func validate(data []byte) (string, error) {
	if isValid := json.Valid(data); !isValid {
		return "", errors.New("incoming data is not valid json")
	}
	order := &Order{}
	if err := json.Unmarshal(data, &order); err != nil {
		return "", fmt.Errorf("failed to unmarhal data: %v", err)
	}
	return order.OrderUID, nil
}
