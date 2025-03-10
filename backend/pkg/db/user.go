package db

import (
	"context"
	"fmt"
)

// GetUserCount returns the total number of users in the database
func (c *DBClient) GetUserCount(ctx context.Context) (int, error) {
	var count int

	err := c.conn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return count, nil
}
