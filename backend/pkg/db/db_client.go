package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBClient struct {
	conn *pgxpool.Pool
}

func NewClient(connString string) (*DBClient, error) {
	conn, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return &DBClient{conn: conn}, nil
}

func (c *DBClient) Close() {
	c.conn.Close()
}
