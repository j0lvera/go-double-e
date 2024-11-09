package doublee

import (
	db "github.com/j0lvera/go-double-e/internal/db/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	queries *db.Queries
}

func NewClient(pool *pgxpool.Pool) *Client {
	return &Client{
		queries: db.New(pool),
	}
}
