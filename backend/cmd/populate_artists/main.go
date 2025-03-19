package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/sukhmai/spotify-match/pkg/db"
	"github.com/sukhmai/spotify-match/pkg/spotify"
)

const (
	TargetArtistCount = 25000
)

func main() {
	// Initialize Spotify client
	log.Println("Running...")
	spotifyClient, err := spotify.NewSpotifyClient()
	if err != nil {
		log.Fatalf("Failed to create Spotify client: %v", err)
	}

	// Initialize database client
	dbAddr := os.Getenv("DB_HOST")
	if dbAddr == "" {
		dbAddr = "localhost:5432"
	}

	username := os.Getenv("DB_USERNAME")
	if username == "" {
		username = "spotifyuser"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "spotify"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		log.Fatal("DB_PASSWORD environment variable not set")
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s/%s", username, password, dbAddr, dbName)
	dbClient, err := db.NewClient(connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbClient.Close()

	// Get access token using client credentials
	token, err := spotifyClient.GetClientCredentialsToken()
	if err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}

	log.Println("Starting artist database population...")

	// Get popular artists using enhanced search approach
	artists := getSeedArtists(spotifyClient, token)
	log.Printf("Collected %d artists", len(artists))

	// Save artists to database
	log.Println("Saving artists to database...")
	existingArtists, err := getAllArtistIDs(dbClient)
	if err != nil {
		log.Printf("Warning: Could not load existing artists: %v", err)
	} else {
		log.Printf("Loaded %d existing artists from database", len(existingArtists))
	}

	// Track existing artists to avoid duplicates
	existingMap := make(map[string]bool)
	for _, id := range existingArtists {
		existingMap[id] = true
	}

	// Save artists to database
	var newCount int
	for _, artist := range artists {
		if !existingMap[artist.ID] {
			saveArtistToDatabase(dbClient, artist)
			newCount++

			// Log progress
			if newCount%100 == 0 {
				log.Printf("Added %d new artists to database...", newCount)
			}
		}
	}

	log.Printf("Added %d new artists to database", newCount)
	log.Println("Artist database population completed!")
}

func getSeedArtists(spotifyClient *spotify.SpotifyClient, token string) []spotify.Artist {
	// Track artists by ID to avoid duplicates
	artistMap := make(map[string]spotify.Artist)

	// Letters in order of frequency in English
	// Based on common frequency analysis: E, T, A, O, I, N, S, H, R, D, L, U, C, M, W, F, G, Y, P, B, V, K, J, X, Q, Z
	lettersByFrequency := []string{
		"e", "t", "a", "o", "i", "n", "s", "h", "r", "d",
		"l", "u", "c", "m", "w", "f", "g", "y", "p", "b",
		"v", "k", "j", "x", "q", "z",
	}

	// Calculate how many artists to target per search method to achieve even distribution
	// We aim for TargetArtistCount*2 to have a buffer for sorting by popularity later

	// Allocate 60% to single letters and 40% to letter combinations
	singleLetterTarget := TargetArtistCount * 2 * 60 / 100
	letterCombinationTarget := TargetArtistCount * 2 * 40 / 100

	// Calculate artists per letter for single letter searches
	artistsPerLetter := singleLetterTarget / len(lettersByFrequency)
	pagesPerLetter := (artistsPerLetter + 49) / 50 // Ceiling division to get pages (50 artists per page)

	log.Printf("Targeting approximately %d artists per letter (%d pages per letter)",
		artistsPerLetter, pagesPerLetter)

	// First search by single letters in frequency order
	log.Println("Starting single letter frequency searches...")
	for _, letter := range lettersByFrequency {
		searchTerm := letter + "*"
		log.Printf("Searching for artists with '%s'", searchTerm)

		// Search multiple pages for each letter
		for offset := 0; offset < pagesPerLetter*50; offset += 50 {
			artists, err := spotifyClient.SearchArtists(searchTerm, 50, offset, token, "US")
			if err != nil {
				log.Printf("Error searching for artists with '%s': %v", searchTerm, err)
				continue
			}

			// Add artists to map to avoid duplicates
			for _, artist := range artists {
				if artist.ID != "" {
					artistMap[artist.ID] = artist
				}
			}

			// Log progress
			if offset == 0 {
				log.Printf("Found %d artists with search term '%s'", len(artists), searchTerm)
			}

			// Respect rate limits - increased delay to avoid timeouts
			time.Sleep(300 * time.Millisecond)
		}

		log.Printf("Total unique artists after searching '%s': %d", searchTerm, len(artistMap))
	}

	// Then search by 2-letter combinations using frequency
	log.Println("Starting 2-letter combination searches...")

	// Use the most frequent letters for combinations
	// We'll use the top 8 most frequent letters for combinations
	mostFrequentLetters := lettersByFrequency[:8] // e, t, a, o, i, n, s, h

	// Calculate how many combinations we'll have
	totalCombinations := len(mostFrequentLetters) * len(mostFrequentLetters)
	artistsPerCombination := letterCombinationTarget / totalCombinations

	// Ensure we get at least one page per combination
	pagesPerCombination := 1
	if artistsPerCombination > 50 {
		pagesPerCombination = (artistsPerCombination + 49) / 50 // Ceiling division
	}

	log.Printf("Targeting approximately %d artists per 2-letter combination (%d pages per combination)",
		artistsPerCombination, pagesPerCombination)

	for _, letter1 := range mostFrequentLetters {
		for _, letter2 := range mostFrequentLetters {
			searchTerm := letter1 + letter2 + "*"
			log.Printf("Searching for artists with '%s'", searchTerm)

			// Search multiple pages for each combination
			for offset := 0; offset < pagesPerCombination*50; offset += 50 {
				artists, err := spotifyClient.SearchArtists(searchTerm, 50, offset, token, "US")
				if err != nil {
					log.Printf("Error searching for artists with '%s': %v", searchTerm, err)
					continue
				}

				// Add artists to map to avoid duplicates
				for _, artist := range artists {
					if artist.ID != "" {
						artistMap[artist.ID] = artist
					}
				}

				// Log progress
				if offset == 0 {
					log.Printf("Found %d artists with search term '%s'", len(artists), searchTerm)
				}

				// Respect rate limits - increased delay to avoid timeouts
				time.Sleep(300 * time.Millisecond)
			}

			log.Printf("Total unique artists after searching '%s': %d", searchTerm, len(artistMap))
		}
	}

	// Convert map to slice
	var allArtists []spotify.Artist
	for _, artist := range artistMap {
		allArtists = append(allArtists, artist)
	}

	// Sort by popularity (highest first)
	sort.Slice(allArtists, func(i, j int) bool {
		return allArtists[i].Popularity > allArtists[j].Popularity
	})

	log.Printf("Final artist count after sorting by popularity: %d", len(allArtists))

	return allArtists
}

func getAllArtistIDs(dbClient *db.DBClient) ([]string, error) {
	// Use the method we implemented in the db package
	ctx := context.Background()
	return dbClient.GetAllArtistIDs(ctx)
}

// saveArtistToDatabase saves the artist to the database
func saveArtistToDatabase(dbClient *db.DBClient, artist spotify.Artist) {
	// Use the method we implemented in the db package
	ctx := context.Background()

	// Convert spotify.Artist to db.Artist with all fields
	dbArtist := db.Artist{
		ID:         artist.ID,
		Name:       artist.Name,
		Genres:     artist.Genres,
		Popularity: artist.Popularity,
		SpotifyURL: artist.ExternalURLs.Spotify,
	}

	// Convert images
	if len(artist.Images) > 0 {
		dbArtist.Images = make([]struct {
			URL    string
			Height int
			Width  int
		}, len(artist.Images))

		for i, img := range artist.Images {
			dbArtist.Images[i].URL = img.URL
			dbArtist.Images[i].Height = img.Height
			dbArtist.Images[i].Width = img.Width
		}
	}

	// Insert the artist
	err := dbClient.InsertArtist(ctx, dbArtist)
	if err != nil {
		log.Printf("Error inserting artist %s: %v", artist.Name, err)
	}
}
