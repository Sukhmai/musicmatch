package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sukhmai/spotify-match/gen/spotify/v1/spotifyv1connect"
	"github.com/sukhmai/spotify-match/pkg/api"
)

func main() {
	server, err := api.NewDefaultServer()
	if err != nil {
		log.Fatal(err)
	}
	spotifyServer := api.NewSpotifyServer(server)

	defer server.Close(context.Background())

	r := chi.NewRouter()
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	spotifyPath, spotifyHandler := spotifyv1connect.NewSpotifyServiceHandler(spotifyServer)
	r.Mount(spotifyPath, spotifyHandler)

	fmt.Println("Server starting on port 8080")
	http.ListenAndServe(
		"0.0.0.0:8080",
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(r, &http2.Server{}),
	)
}
