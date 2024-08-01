package pg

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s0vunia/platform_common/pkg/db"

	"github.com/jackc/pgx/v4/pgxpool"
)

type pgClient struct {
	masterDBC db.DB
}

// New creates a new PostgreSQL client.
func New(ctx context.Context, dsn string) (db.Client, error) {
	dbc, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, errors.Errorf("failed to connect to db: %v", err)
	}

	return &pgClient{
		masterDBC: &pg{dbc: dbc},
	}, nil
}

func (c *pgClient) DB() db.DB {
	return c.masterDBC
}

func (c *pgClient) Close() error {
	if c.masterDBC != nil {
		c.masterDBC.Close()
	}

	return nil
}
