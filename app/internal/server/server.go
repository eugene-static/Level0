package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eugene-static/Level0/app/internal/cache"
	"github.com/eugene-static/Level0/app/internal/config"
	"github.com/eugene-static/Level0/app/internal/nats"
	"github.com/eugene-static/Level0/app/internal/service"
	"github.com/eugene-static/Level0/app/internal/storage/postgres"
	transport "github.com/eugene-static/Level0/app/internal/transport/http"
)

func Run() {
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	//l.Info("")
	cfgPath := os.Getenv("CFG_PATH")
	l.Info("getting config")
	cfg, err := config.Get(cfgPath)
	if err != nil || cfg == nil {
		l.Error("failed to get config", slog.Any("details", err))
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	listener, err := net.Listen(cfg.Server.Network, fmt.Sprintf("%s:%s", cfg.Server.IP, cfg.Server.Port))
	if err != nil {
		l.Error("failed to announce listener", slog.Any("details", err))
		return
	}
	str, err := postgres.New(ctx, &cfg.Postgres)
	if err != nil {
		l.Error("failed to get storage", slog.Any("details", err))
		return
	}
	cch := cache.New(&sync.RWMutex{})
	srv := service.New(str, cch)
	if err = srv.Init(ctx); err != nil {
		l.Error("failed to init cache", slog.Any("details", err))
		return
	}
	handler := transport.New(l, srv)
	router := http.NewServeMux()
	handler.Register(router)
	server := http.Server{
		Handler:      router,
		ReadTimeout:  cfg.Server.Timeout * time.Second,
		WriteTimeout: cfg.Server.Timeout * time.Second,
	}
	stream := nats.New(&cfg.Nats, l, srv)
	sub, err := stream.Connect(ctx)
	if err != nil {
		l.Error("failed to connect to nats-streaming", slog.Any("details", err))
		return
	}
	go func() {
		if err = server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.ErrorContext(ctx, "failed to start server", slog.Any("details", err))
			panic(err)
		}
	}()
	l.Info("listening on", slog.String("IP", cfg.Server.IP), slog.String("port", cfg.Server.Port))
	<-ctx.Done()
	l.Info("server is shutting down...")
	sdCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.Shutdown*time.Second)
	defer cancel()
	if err = server.Shutdown(sdCtx); err != nil {
		l.Error("failed to shutdown server", slog.Any("details", err))
		return
	}
	longSD := make(chan struct{}, 1)

	go func() {
		err = sub.Unsubscribe()
		if err != nil {
			l.Error("unable to unsubscribe", slog.Any("details", err))
			return
		}
		longSD <- struct{}{}
	}()
	select {
	case <-sdCtx.Done():
		l.Error("shutdown error", slog.Any("details", sdCtx.Err()))
		return
	case <-longSD:
		l.Info("shutdown finished")
	}
}
