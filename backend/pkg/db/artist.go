package db

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5" // Import for side effects
)

// Artist represents a Spotify artist with ID and name
type Artist struct {
	ID   string
	Name string
}

// UserInfo represents the user information to be saved
type UserInfo struct {
	FirstName     string
	LastName      string
	Email         string
	PhoneNumber   string
	SpotifyUserID string // Unique identifier from Spotify
}

// SaveUserTopArtists saves a user and their top artists to the database
// Returns the user ID, newly added artists, and any error
func (c *DBClient) SaveUserTopArtists(ctx context.Context, user UserInfo, artists []Artist) (string, []Artist, error) {
	// Begin a transaction
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Ensure the transaction is rolled back if an error occurs
	defer tx.Rollback(ctx)

	// Track newly added artists
	var newArtists []Artist

	// Insert or update the user
	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users (first_name, last_name, email, phone_number, spotify_user_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (spotify_user_id) DO UPDATE
		SET first_name = $1, last_name = $2, email = $3, phone_number = $4
		RETURNING user_id`,
		user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.SpotifyUserID).Scan(&userID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert/update user: %w", err)
	}

	// Delete existing user-artist relationships for this user
	_, err = tx.Exec(ctx, "DELETE FROM user_artists WHERE user_id = $1", userID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to delete existing user-artist relationships: %w", err)
	}

	// For each artist, insert if not exists and link to the user
	for i, artist := range artists {
		// Check if the artist already exists
		var artistID int
		var exists bool
		err = tx.QueryRow(ctx,
			`SELECT artist_id, true FROM artists WHERE spotify_artist_id = $1`,
			artist.ID).Scan(&artistID, &exists)

		if err != nil {
			// Artist doesn't exist, insert it
			err = tx.QueryRow(ctx,
				`INSERT INTO artists (spotify_artist_id, artist_name)
				VALUES ($1, $2)
				RETURNING artist_id`,
				artist.ID, artist.Name).Scan(&artistID)
			if err != nil {
				return "", nil, fmt.Errorf("failed to insert artist %s: %w", artist.Name, err)
			}
			// Add to the list of new artists
			newArtists = append(newArtists, artist)
		} else {
			// Artist exists, update the name if needed
			_, err = tx.Exec(ctx,
				`UPDATE artists SET artist_name = $2 WHERE spotify_artist_id = $1`,
				artist.ID, artist.Name)
			if err != nil {
				return "", nil, fmt.Errorf("failed to update artist %s: %w", artist.Name, err)
			}
		}

		// Link the user to the artist with the appropriate rank
		_, err = tx.Exec(ctx,
			`INSERT INTO user_artists (user_id, artist_id, rank)
			VALUES ($1, $2, $3)`,
			userID, artistID, i+1)
		if err != nil {
			return "", nil, fmt.Errorf("failed to link user to artist %s: %w", artist.Name, err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return "", nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return userID, newArtists, nil
}
