package db

import (
	db "github.com/j0lvera/go-double-e/internal/db/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	// Make sqlc queries public so handlers can use it directly
	Queries *db.Queries
}

func NewClient(pool *pgxpool.Pool) *Client {
	return &Client{
		Queries: db.New(pool),
	}
}
