package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eugene-static/Level0/app/lib/config"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(ctx context.Context, cfg *config.Postgres) (*Storage, error) {
	db, err := sql.Open("postgres", dsn(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open database driver: %v", err)
	}
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot ping the database: %v", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Close(ctx context.Context) error {
	err := s.db.Close()
	if err != nil {
		return fmt.Errorf("cannot close database: %v", err)
	}
	return nil
}

func dsn(cfg *config.Postgres) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSL)
}
