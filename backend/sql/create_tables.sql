drop table if exists user_artists;
drop table if exists artists;
drop table if exists users;

-- Enable the uuid-ossp extension for generating UUIDs (if needed)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone_number TEXT,
    spotify_user_id TEXT UNIQUE,  -- Unique identifier from Spotify
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE artists (
    artist_id SERIAL PRIMARY KEY,
    spotify_artist_id TEXT UNIQUE NOT NULL,
    artist_name TEXT NOT NULL
);

CREATE TABLE user_artists (
    user_id INT REFERENCES users(user_id),
    artist_id INT REFERENCES artists(artist_id),
    rank INT,  -- Optional: rank of the artist for the user (e.g., 1 for top artist)
    PRIMARY KEY (user_id, artist_id)
);

CREATE INDEX idx_user_artists_user_id ON user_artists(user_id);
CREATE INDEX idx_user_artists_artist_id ON user_artists(artist_id);
