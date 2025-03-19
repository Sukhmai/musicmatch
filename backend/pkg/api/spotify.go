package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/url"

	"connectrpc.com/connect"
	spotifyv1 "github.com/sukhmai/spotify-match/gen/spotify/v1"
	"github.com/sukhmai/spotify-match/gen/spotify/v1/spotifyv1connect"
	"github.com/sukhmai/spotify-match/pkg/db"
	"github.com/sukhmai/spotify-match/pkg/spotify"
)

// Maximum number of users allowed for the current round
const MaxUsersPerRound = 500

type SpotifyServer struct {
	spotifyv1connect.UnimplementedSpotifyServiceHandler
	*Server
	*spotify.SpotifyClient
}

func NewSpotifyServer(s *Server) *SpotifyServer {
	spotifyClient, err := spotify.NewSpotifyClient()
	if err != nil {
		log.Fatal(err)
	}
	return &SpotifyServer{
		Server:        s,
		SpotifyClient: spotifyClient,
	}
}

// SaveTopArtists retrieves the user's top artists from Spotify and saves them to the database
func (s *SpotifyServer) SaveTopArtists(ctx context.Context,
	req *connect.Request[spotifyv1.SaveTopArtistsRequest],
) (*connect.Response[spotifyv1.SaveTopArtistsResponse], error) {
	// Extract the access token from the request
	accessToken := req.Msg.AccessToken
	if accessToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("access token is required"))
	}

	// Get the user's profile from Spotify
	profile, err := s.SpotifyClient.GetUserProfile(accessToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user profile: %w", err))
	}

	// Get the user's top artists from Spotify
	artists, err := s.SpotifyClient.GetArtists(accessToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get top artists: %w", err))
	}

	// Get the database client
	dbClient := s.dbClient

	// Check if we've reached the maximum number of users
	userCount, err := dbClient.GetUserCount(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user count: %w", err))
	}

	// If we've reached the limit, return an error
	if userCount >= MaxUsersPerRound {
		return nil, connect.NewError(connect.CodeResourceExhausted,
			errors.New("maximum number of users reached for this round, please wait for the next round"))
	}

	// Create user info struct with Spotify user ID
	userInfo := db.UserInfo{
		FirstName:     req.Msg.FirstName,
		LastName:      req.Msg.LastName,
		Email:         req.Msg.Email,
		PhoneNumber:   req.Msg.Number,
		SpotifyUserID: profile.ID,
	}

	// Convert Spotify artists to database artists with all fields
	dbArtists := make([]db.Artist, len(artists))
	for i, artist := range artists {
		dbArtists[i] = db.Artist{
			ID:         artist.ID,
			Name:       artist.Name,
			Genres:     artist.Genres,
			Popularity: artist.Popularity,
			SpotifyURL: artist.ExternalURLs.Spotify,
		}

		// Convert images
		if len(artist.Images) > 0 {
			dbArtists[i].Images = make([]struct {
				URL    string
				Height int
				Width  int
			}, len(artist.Images))

			for j, img := range artist.Images {
				dbArtists[i].Images[j].URL = img.URL
				dbArtists[i].Images[j].Height = img.Height
				dbArtists[i].Images[j].Width = img.Width
			}
		}
	}

	// Save user and artists to the database
	userID, newArtists, err := dbClient.SaveUserTopArtists(ctx, userInfo, dbArtists)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to save user and artists: %w", err))
	}

	// Convert the new artists to response format with additional information
	uniqueArtists := make([]*spotifyv1.ArtistInfo, len(newArtists))
	for i, artist := range newArtists {
		// Find the original artist object with full details
		var fullArtist *spotify.Artist
		for _, a := range artists {
			if a.ID == artist.ID {
				fullArtist = &a
				break
			}
		}

		artistInfo := &spotifyv1.ArtistInfo{
			Id:   artist.ID,
			Name: artist.Name,
		}

		// Add additional information if available
		if fullArtist != nil {
			// Add images
			artistImages := make([]*spotifyv1.ArtistImage, len(fullArtist.Images))
			for j, img := range fullArtist.Images {
				artistImages[j] = &spotifyv1.ArtistImage{
					Url:    img.URL,
					Height: int32(img.Height),
					Width:  int32(img.Width),
				}
			}
			artistInfo.Images = artistImages

			// Add genres
			artistInfo.Genres = fullArtist.Genres

			// Add popularity
			artistInfo.Popularity = int32(fullArtist.Popularity)

			// Add Spotify URL
			artistInfo.SpotifyUrl = fullArtist.ExternalURLs.Spotify
		}

		uniqueArtists[i] = artistInfo
	}

	// Return the response with user ID and unique artists
	return connect.NewResponse(&spotifyv1.SaveTopArtistsResponse{
		UserId:        userID,
		UniqueArtists: uniqueArtists,
	}), nil
}

func (s *SpotifyServer) GetAuthURL(ctx context.Context, req *connect.Request[spotifyv1.GetAuthURLRequest]) (*connect.Response[spotifyv1.GetAuthURLResponse], error) {
	// Generate a random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not generate state: %w", err))
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Construct the URL
	baseURL := "https://accounts.spotify.com/authorize"
	params := url.Values{}
	params.Add("client_id", s.SpotifyClient.ClientID)
	params.Add("redirect_uri", s.SpotifyClient.CallbackURL)
	params.Add("response_type", "code")
	params.Add("scope", "user-top-read user-read-email")
	params.Add("state", state)

	authURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	return connect.NewResponse(&spotifyv1.GetAuthURLResponse{
		Url: authURL,
	}), nil
}

// ExchangeToken exchanges the authorization code for access and refresh tokens
// This is called by the frontend after receiving the code from Spotify
func (s *SpotifyServer) ExchangeToken(ctx context.Context,
	req *connect.Request[spotifyv1.ExchangeTokenRequest],
) (*connect.Response[spotifyv1.ExchangeTokenResponse], error) {
	// Validate the request
	if req.Msg.Code == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("code is required"))
	}

	if req.Msg.State == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("state is required"))
	}

	// Exchange the code for access and refresh tokens
	tokenResponse, err := s.SpotifyClient.GetTokens(req.Msg.Code, req.Msg.State)
	if err != nil {
		log.Printf("Error getting tokens: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get tokens: %w", err))
	}

	// Return the tokens in the response
	return connect.NewResponse(&spotifyv1.ExchangeTokenResponse{
		AccessToken:  tokenResponse.AccessToken,
		TokenType:    tokenResponse.TokenType,
		RefreshToken: tokenResponse.RefreshToken,
		ExpiresIn:    int32(tokenResponse.ExpiresIn),
	}), nil
}

func (s *SpotifyServer) GetUserCount(ctx context.Context,
	req *connect.Request[spotifyv1.GetUserCountRequest],
) (*connect.Response[spotifyv1.GetUserCountResponse], error) {
	userCount, err := s.dbClient.GetUserCount(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&spotifyv1.GetUserCountResponse{
		Count:    int32(userCount),
		MaxUsers: MaxUsersPerRound,
	}), nil
}

// SearchArtists searches for artists in the database matching the query
// If database search yields few results, it will search Spotify API directly
func (s *SpotifyServer) SearchArtists(ctx context.Context,
	req *connect.Request[spotifyv1.SearchArtistsRequest],
) (*connect.Response[spotifyv1.SearchArtistsResponse], error) {
	// Extract parameters from the request
	query := req.Msg.Query
	if query == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("query is required"))
	}

	limit := int(req.Msg.Limit)
	if limit <= 0 {
		limit = 10 // Default limit
	}

	offset := int(req.Msg.Offset)
	if offset < 0 {
		offset = 0 // Default offset
	}

	// Search for artists in the database
	dbArtists, total, err := s.dbClient.SearchArtists(ctx, query, limit, offset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to search artists in database: %w", err))
	}

	// Define a minimum threshold for database results
	const minResultsThreshold = 5

	// If database search yields few results, search Spotify API directly
	var spotifyArtists []spotify.Artist
	if total < minResultsThreshold {
		// Get a client credentials token for Spotify API
		token, err := s.SpotifyClient.GetClientCredentialsToken()
		if err != nil {
			log.Printf("Warning: Failed to get Spotify API token: %v", err)
		} else {
			// Search Spotify API
			spotifyArtists, err = s.SpotifyClient.SearchArtists(query, limit, 0, token, "US")
			if err != nil {
				log.Printf("Warning: Failed to search Spotify API: %v", err)
			} else {
				log.Printf("Found %d additional artists from Spotify API for query: %s", len(spotifyArtists), query)

				// Save new artists to database
				for _, artist := range spotifyArtists {
					// Check if artist already exists in our results
					exists := false
					for _, dbArtist := range dbArtists {
						if dbArtist.ID == artist.ID {
							exists = true
							break
						}
					}

					if !exists {
						// Convert to database artist format with all fields
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

							for j, img := range artist.Images {
								dbArtist.Images[j].URL = img.URL
								dbArtist.Images[j].Height = img.Height
								dbArtist.Images[j].Width = img.Width
							}
						}

						// Insert artist into database (non-blocking)
						go func(a db.Artist) {
							err := s.dbClient.InsertArtist(context.Background(), a)
							if err != nil {
								log.Printf("Warning: Failed to insert artist %s: %v", a.Name, err)
							}
						}(dbArtist)
					}
				}
			}
		}
	}

	// Combine database and Spotify results
	// First convert database artists to response format
	artistInfos := make([]*spotifyv1.ArtistInfo, 0, len(dbArtists)+len(spotifyArtists))

	// Add database artists
	for _, artist := range dbArtists {
		artistInfo := &spotifyv1.ArtistInfo{
			Id:   artist.ID,
			Name: artist.Name,
		}

		// Add additional artist information if available
		if len(artist.Images) > 0 {
			artistImages := make([]*spotifyv1.ArtistImage, len(artist.Images))
			for j, img := range artist.Images {
				artistImages[j] = &spotifyv1.ArtistImage{
					Url:    img.URL,
					Height: int32(img.Height),
					Width:  int32(img.Width),
				}
			}
			artistInfo.Images = artistImages
		}

		if len(artist.Genres) > 0 {
			artistInfo.Genres = artist.Genres
		}

		if artist.Popularity > 0 {
			artistInfo.Popularity = int32(artist.Popularity)
		}

		if artist.SpotifyURL != "" {
			artistInfo.SpotifyUrl = artist.SpotifyURL
		}

		artistInfos = append(artistInfos, artistInfo)
	}

	// Add Spotify artists (if any)
	// Track IDs to avoid duplicates
	seenIDs := make(map[string]bool)
	for _, dbArtist := range dbArtists {
		seenIDs[dbArtist.ID] = true
	}

	for _, artist := range spotifyArtists {
		// Skip if already in results
		if seenIDs[artist.ID] {
			continue
		}
		seenIDs[artist.ID] = true

		// Convert to response format
		artistInfo := &spotifyv1.ArtistInfo{
			Id:   artist.ID,
			Name: artist.Name,
		}

		// Add images
		if len(artist.Images) > 0 {
			artistImages := make([]*spotifyv1.ArtistImage, len(artist.Images))
			for j, img := range artist.Images {
				artistImages[j] = &spotifyv1.ArtistImage{
					Url:    img.URL,
					Height: int32(img.Height),
					Width:  int32(img.Width),
				}
			}
			artistInfo.Images = artistImages
		}

		// Add genres
		if len(artist.Genres) > 0 {
			artistInfo.Genres = artist.Genres
		}

		// Add popularity
		if artist.Popularity > 0 {
			artistInfo.Popularity = int32(artist.Popularity)
		}

		// Add Spotify URL
		if artist.ExternalURLs.Spotify != "" {
			artistInfo.SpotifyUrl = artist.ExternalURLs.Spotify
		}

		artistInfos = append(artistInfos, artistInfo)
	}

	// Calculate new total
	newTotal := total + len(spotifyArtists)
	if newTotal > total {
		// Avoid double-counting artists that exist in both results
		newTotal = len(seenIDs)
	}

	// Return the combined response
	return connect.NewResponse(&spotifyv1.SearchArtistsResponse{
		Artists: artistInfos,
		Total:   int32(newTotal),
	}), nil
}

// SaveUserSelectedArtists saves a user and their manually selected artists
func (s *SpotifyServer) SaveUserSelectedArtists(ctx context.Context,
	req *connect.Request[spotifyv1.SaveUserSelectedArtistsRequest],
) (*connect.Response[spotifyv1.SaveUserSelectedArtistsResponse], error) {
	// Validate the request
	if req.Msg.FirstName == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("first_name is required"))
	}
	if req.Msg.LastName == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("last_name is required"))
	}
	if req.Msg.Email == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email is required"))
	}
	if len(req.Msg.ArtistIds) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("artist_ids is required"))
	}

	// Check if we've reached the maximum number of users
	userCount, err := s.dbClient.GetUserCount(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user count: %w", err))
	}

	// If we've reached the limit, return an error
	if userCount >= MaxUsersPerRound {
		return nil, connect.NewError(connect.CodeResourceExhausted,
			errors.New("maximum number of users reached for this round, please wait for the next round"))
	}

	// Create user info struct (without Spotify user ID since we don't have it)
	userInfo := db.UserInfo{
		FirstName:   req.Msg.FirstName,
		LastName:    req.Msg.LastName,
		Email:       req.Msg.Email,
		PhoneNumber: req.Msg.Number,
	}

	// Save the user and their selected artists
	userID, artists, err := s.dbClient.SaveUserSelectedArtists(ctx, userInfo, req.Msg.ArtistIds)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to save user and artists: %w", err))
	}

	// Convert to response format
	uniqueArtists := make([]*spotifyv1.ArtistInfo, len(artists))
	for i, artist := range artists {
		artistInfo := &spotifyv1.ArtistInfo{
			Id:   artist.ID,
			Name: artist.Name,
		}

		// Add additional artist information if available
		if len(artist.Images) > 0 {
			artistImages := make([]*spotifyv1.ArtistImage, len(artist.Images))
			for j, img := range artist.Images {
				artistImages[j] = &spotifyv1.ArtistImage{
					Url:    img.URL,
					Height: int32(img.Height),
					Width:  int32(img.Width),
				}
			}
			artistInfo.Images = artistImages
		}

		if len(artist.Genres) > 0 {
			artistInfo.Genres = artist.Genres
		}

		if artist.Popularity > 0 {
			artistInfo.Popularity = int32(artist.Popularity)
		}

		if artist.SpotifyURL != "" {
			artistInfo.SpotifyUrl = artist.SpotifyURL
		}

		uniqueArtists[i] = artistInfo
	}

	// Return the response
	return connect.NewResponse(&spotifyv1.SaveUserSelectedArtistsResponse{
		UserId:        userID,
		UniqueArtists: uniqueArtists,
	}), nil
}
