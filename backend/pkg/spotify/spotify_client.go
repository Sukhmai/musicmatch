package spotify

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type SpotifyClient struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
}

func NewSpotifyClient() (*SpotifyClient, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	if clientID == "" {
		return nil, errors.New("SPOTIFY_CLIENT_ID environment variable not set")
	}

	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, errors.New("SPOTIFY_CLIENT_SECRET environment variable not set")
	}
	callbackURL := os.Getenv("SPOTIFY_CALLBACK_URL")
	if callbackURL == "" {
		callbackURL = "http://localhost:5173/callback"
	}

	return &SpotifyClient{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CallbackURL:  callbackURL,
	}, nil
}

const (
	spotifyAuthorizeURL = "https://accounts.spotify.com/authorize"
	spotifyTokenURL     = "https://accounts.spotify.com/api/token"
	spotifyAPIURL       = "https://api.spotify.com/v1"
)

// generateRandomState creates a random state string for CSRF protection
func generateRandomState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (c *SpotifyClient) Authorize() (string, string, error) {
	scopes := "user-read-email user-top-read" // Ensure user-read-email scope is included

	// Generate random state for CSRF protection
	state, err := generateRandomState()
	if err != nil {
		return "", "", fmt.Errorf("could not generate random state: %w", err)
	}

	v := url.Values{}
	v.Set("client_id", c.ClientID)
	v.Set("response_type", "code")
	v.Set("redirect_uri", c.CallbackURL)
	v.Set("scope", scopes)
	v.Set("state", state)

	authorizeURL := spotifyAuthorizeURL + "?" + v.Encode()

	return authorizeURL, state, nil
}

// Artist represents a Spotify artist
type Artist struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	URI          string   `json:"uri"`
	Images       []Image  `json:"images"`
	Genres       []string `json:"genres"`
	Popularity   int      `json:"popularity"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
}

// Image represents a Spotify image
type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// UserProfile represents a Spotify user profile
type UserProfile struct {
	ID    string `json:"id"`
	URI   string `json:"uri"`
	Email string `json:"email"`
	// Other fields as needed
}

// TopArtistsResponse represents the response from the Spotify API for top artists
type TopArtistsResponse struct {
	Items []Artist `json:"items"`
}

func (c *SpotifyClient) GetArtists(accessToken string) ([]Artist, error) {
	apiURL := spotifyAPIURL + "/me/top/artists?limit=50"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spotify API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var topArtists TopArtistsResponse
	err = json.Unmarshal(body, &topArtists)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %w", err)
	}

	// Return the full artist objects
	return topArtists.Items, nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// GetUserProfile retrieves the current user's Spotify profile
func (c *SpotifyClient) GetUserProfile(accessToken string) (*UserProfile, error) {
	apiURL := spotifyAPIURL + "/me"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spotify API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var profile UserProfile
	err = json.Unmarshal(body, &profile)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %w", err)
	}

	return &profile, nil
}

func (c *SpotifyClient) GetTokens(code string, state string) (*TokenResponse, error) {
	tokenURL := spotifyTokenURL

	data := url.Values{}
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.CallbackURL)
	data.Set("state", state)

	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = io.NopCloser(io.Reader(strings.NewReader(data.Encode())))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spotify API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResponse TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %w", err)
	}

	return &tokenResponse, nil
}
