package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5" // Import for side effects
)

// Artist represents a Spotify artist with all available information
type Artist struct {
	ID     string
	Name   string
	Genres []string
	Images []struct {
		URL    string
		Height int
		Width  int
	}
	Popularity int
	SpotifyURL string
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
			// Artist doesn't exist, insert it with all fields
			// Convert genres to JSONB
			var genresJSON []byte
			if len(artist.Genres) > 0 {
				genresJSON, err = json.Marshal(artist.Genres)
				if err != nil {
					return "", nil, fmt.Errorf("failed to marshal genres: %w", err)
				}
			}

			// Convert images to JSONB
			var imagesJSON []byte
			if len(artist.Images) > 0 {
				imagesJSON, err = json.Marshal(artist.Images)
				if err != nil {
					return "", nil, fmt.Errorf("failed to marshal images: %w", err)
				}
			}

			err = tx.QueryRow(ctx,
				`INSERT INTO artists 
				(spotify_artist_id, artist_name, genres, images, popularity, spotify_url)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING artist_id`,
				artist.ID, artist.Name, genresJSON, imagesJSON, artist.Popularity, artist.SpotifyURL).Scan(&artistID)
			if err != nil {
				return "", nil, fmt.Errorf("failed to insert artist %s: %w", artist.Name, err)
			}
			// Add to the list of new artists
			newArtists = append(newArtists, artist)
		} else {
			// Artist exists, update all fields
			// Convert genres to JSONB
			var genresJSON []byte
			if len(artist.Genres) > 0 {
				genresJSON, err = json.Marshal(artist.Genres)
				if err != nil {
					return "", nil, fmt.Errorf("failed to marshal genres: %w", err)
				}
			}

			// Convert images to JSONB
			var imagesJSON []byte
			if len(artist.Images) > 0 {
				imagesJSON, err = json.Marshal(artist.Images)
				if err != nil {
					return "", nil, fmt.Errorf("failed to marshal images: %w", err)
				}
			}

			_, err = tx.Exec(ctx,
				`UPDATE artists 
				SET artist_name = $2, 
				    genres = $3, 
				    images = $4, 
				    popularity = $5, 
				    spotify_url = $6 
				WHERE spotify_artist_id = $1`,
				artist.ID, artist.Name, genresJSON, imagesJSON, artist.Popularity, artist.SpotifyURL)
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

// SaveUserSelectedArtists saves a user and their manually selected artists to the database
// This is similar to SaveUserTopArtists but doesn't require a Spotify user ID
func (c *DBClient) SaveUserSelectedArtists(ctx context.Context, user UserInfo, artistIDs []string) (string, []Artist, error) {
	// Begin a transaction
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Ensure the transaction is rolled back if an error occurs
	defer tx.Rollback(ctx)

	// Track artists to return
	var returnArtists []Artist

	// Insert the user (without Spotify ID)
	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users (first_name, last_name, email, phone_number)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id`,
		user.FirstName, user.LastName, user.Email, user.PhoneNumber).Scan(&userID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert user: %w", err)
	}

	// For each artist ID, check if it exists and link to the user
	for i, artistID := range artistIDs {
		// Get artist details with all fields
		var dbArtistID int
		var artistName string
		var genresJSON, imagesJSON []byte
		var popularity sql.NullInt32
		var spotifyURL sql.NullString

		err = tx.QueryRow(ctx,
			`SELECT artist_id, artist_name, genres, images, popularity, spotify_url 
			FROM artists WHERE spotify_artist_id = $1`,
			artistID).Scan(&dbArtistID, &artistName, &genresJSON, &imagesJSON, &popularity, &spotifyURL)

		if err != nil {
			return "", nil, fmt.Errorf("artist with ID %s not found: %w", artistID, err)
		}

		// Link the user to the artist with the appropriate rank
		_, err = tx.Exec(ctx,
			`INSERT INTO user_artists (user_id, artist_id, rank)
			VALUES ($1, $2, $3)`,
			userID, dbArtistID, i+1)
		if err != nil {
			return "", nil, fmt.Errorf("failed to link user to artist ID %s: %w", artistID, err)
		}

		// Create artist with all available information
		artist := Artist{
			ID:   artistID,
			Name: artistName,
		}

		// Unmarshal genres if present
		if len(genresJSON) > 0 {
			if err := json.Unmarshal(genresJSON, &artist.Genres); err != nil {
				return "", nil, fmt.Errorf("failed to unmarshal genres: %w", err)
			}
		}

		// Unmarshal images if present
		if len(imagesJSON) > 0 {
			if err := json.Unmarshal(imagesJSON, &artist.Images); err != nil {
				return "", nil, fmt.Errorf("failed to unmarshal images: %w", err)
			}
		}

		// Set popularity if present
		if popularity.Valid {
			artist.Popularity = int(popularity.Int32)
		}

		// Set Spotify URL if present
		if spotifyURL.Valid {
			artist.SpotifyURL = spotifyURL.String
		}

		// Add to the list of artists to return
		returnArtists = append(returnArtists, artist)
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return "", nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return userID, returnArtists, nil
}

// SearchArtists searches for artists in the database by name
func (c *DBClient) SearchArtists(ctx context.Context, query string, limit, offset int) ([]Artist, int, error) {
	// Default values
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Prepare the search query (case insensitive partial match)
	searchQuery := "%" + strings.ToLower(query) + "%"

	// First, get the total count
	var total int
	err := c.conn.QueryRow(ctx,
		`SELECT COUNT(*) FROM artists WHERE LOWER(artist_name) LIKE $1`,
		searchQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count matching artists: %w", err)
	}

	// Then get the paginated results with all fields
	rows, err := c.conn.Query(ctx,
		`SELECT spotify_artist_id, artist_name, genres, images, popularity, spotify_url 
		FROM artists 
		WHERE LOWER(artist_name) LIKE $1
		ORDER BY artist_name 
		LIMIT $2 OFFSET $3`,
		searchQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search artists: %w", err)
	}
	defer rows.Close()

	// Process results
	var artists []Artist
	for rows.Next() {
		var artist Artist
		var genresJSON, imagesJSON []byte
		var popularity sql.NullInt32
		var spotifyURL sql.NullString

		if err := rows.Scan(&artist.ID, &artist.Name, &genresJSON, &imagesJSON, &popularity, &spotifyURL); err != nil {
			return nil, 0, fmt.Errorf("failed to scan artist row: %w", err)
		}

		// Unmarshal genres if present
		if len(genresJSON) > 0 {
			if err := json.Unmarshal(genresJSON, &artist.Genres); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal genres: %w", err)
			}
		}

		// Unmarshal images if present
		if len(imagesJSON) > 0 {
			if err := json.Unmarshal(imagesJSON, &artist.Images); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal images: %w", err)
			}
		}

		// Set popularity if present
		if popularity.Valid {
			artist.Popularity = int(popularity.Int32)
		}

		// Set Spotify URL if present
		if spotifyURL.Valid {
			artist.SpotifyURL = spotifyURL.String
		}

		artists = append(artists, artist)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating artists rows: %w", err)
	}

	return artists, total, nil
}

// GetAllArtistIDs returns all artist IDs from the database
func (c *DBClient) GetAllArtistIDs(ctx context.Context) ([]string, error) {
	rows, err := c.conn.Query(ctx, "SELECT spotify_artist_id FROM artists")
	if err != nil {
		return nil, fmt.Errorf("failed to query artists: %w", err)
	}
	defer rows.Close()

	var artistIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan artist ID: %w", err)
		}
		artistIDs = append(artistIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating artist rows: %w", err)
	}

	return artistIDs, nil
}

// InsertArtist inserts a new artist into the database with all available information
func (c *DBClient) InsertArtist(ctx context.Context, artist Artist) error {
	// Convert genres to JSONB
	var genresJSON []byte
	if len(artist.Genres) > 0 {
		var err error
		genresJSON, err = json.Marshal(artist.Genres)
		if err != nil {
			return fmt.Errorf("failed to marshal genres: %w", err)
		}
	}

	// Convert images to JSONB
	var imagesJSON []byte
	if len(artist.Images) > 0 {
		var err error
		imagesJSON, err = json.Marshal(artist.Images)
		if err != nil {
			return fmt.Errorf("failed to marshal images: %w", err)
		}
	}

	// Insert or update the artist with all fields
	_, err := c.conn.Exec(ctx,
		`INSERT INTO artists 
		(spotify_artist_id, artist_name, genres, images, popularity, spotify_url) 
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (spotify_artist_id) DO UPDATE 
		SET artist_name = $2, 
		    genres = $3, 
		    images = $4, 
		    popularity = $5, 
		    spotify_url = $6`,
		artist.ID, artist.Name, genresJSON, imagesJSON, artist.Popularity, artist.SpotifyURL)

	if err != nil {
		return fmt.Errorf("failed to insert/update artist: %w", err)
	}

	return nil
}
