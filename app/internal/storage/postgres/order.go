package postgres

import (
	"context"

	"github.com/eugene-static/Level0/app/internal/models"
)

func (s *Storage) Unload(ctx context.Context) ([]*models.Order, error) {
	query := `SELECT uid, data FROM orders`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	dump := make([]*models.Order, 0)
	for rows.Next() {
		var (
			uid  string
			data []byte
		)
		err = rows.Scan(&uid, &data)
		if err != nil {
			return nil, err
		}
		dump = append(dump, &models.Order{
			UID:  uid,
			Data: data,
		})
	}
	return dump, nil
}

func (s *Storage) Save(ctx context.Context, order *models.Order) error {
	query := `INSERT INTO orders VALUES ($1,$2)`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, order.UID, order.Data)
	return err
}
