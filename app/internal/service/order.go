package service

import (
	"context"

	"github.com/eugene-static/Level0/app/internal/models"
)

type Storage interface {
	Unload(context.Context) ([]*models.Order, error)
	Save(context.Context, *models.Order) error
}

type Cache interface {
	Set(string, []byte)
	Get(string) ([]byte, error)
}
type Service struct {
	Storage Storage
	Cache   Cache
}

func New(storage Storage, cache Cache) *Service {
	return &Service{
		Storage: storage,
		Cache:   cache,
	}
}

func (s *Service) Get(uid string) ([]byte, error) {
	return s.Cache.Get(uid)
}

func (s *Service) Init(ctx context.Context) error {
	dump, err := s.Storage.Unload(ctx)
	if err != nil {
		return err
	}
	for _, data := range dump {
		s.Cache.Set(data.UID, data.Data)
	}
	return nil
}

func (s *Service) Set(ctx context.Context, order *models.Order) error {
	err := s.Storage.Save(ctx, order)
	if err != nil {
		return err
	}
	s.Cache.Set(order.UID, order.Data)
	return nil
}
