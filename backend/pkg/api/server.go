package api

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/sukhmai/spotify-match/pkg/db"
	"go.uber.org/zap"
)

type Server struct {
	dbClient *db.DBClient
	logger   *zap.SugaredLogger
}

const defaultDbUsername = "spotifyuser"
const defaultDbName = "spotify"
const defaultServerAddr = "localhost:5432"

func NewDefaultServer() (*Server, error) {
	// Get database address from environment variable or use default
	dbAddr := os.Getenv("DB_HOST")
	if dbAddr == "" {
		dbAddr = defaultServerAddr
	}
	return NewServer(dbAddr)
}

func NewServer(dbAddr string) (*Server, error) {
	prod, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	// Get database credentials from environment variables with defaults
	username := os.Getenv("DB_USERNAME")
	if username == "" {
		username = defaultDbUsername
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = defaultDbName
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		return nil, errors.New("DB_PASSWORD environment variable not set")
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s/%s", username, password, dbAddr, dbName)
	dbClient, err := db.NewClient(connString)
	if err != nil {
		return nil, err
	}
	logger := prod.Sugar()
	if err != nil {
		return nil, err
	}
	return &Server{
		dbClient: dbClient,
		logger:   logger,
	}, nil
}

func (s *Server) Close(ctx context.Context) error {
	s.dbClient.Close()
	return nil
}
