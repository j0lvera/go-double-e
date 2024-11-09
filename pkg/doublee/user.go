package doublee

import (
	"context"
	"fmt"
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

func (c *Client) CreateUser(ctx context.Context, email, password string) (db.User, error) {
	user, err := c.queries.CreateUser(ctx, db.CreateUserParams{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return db.User{}, fmt.Errorf("error creating user: %w", err)
	}

	return user, nil
}
