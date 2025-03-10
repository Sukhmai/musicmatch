-- Clear existing data
TRUNCATE users, artists, user_artists CASCADE;

-- Enable the uuid-ossp extension (should already be enabled from create_tables.sql)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Insert users
INSERT INTO users (user_id, first_name, last_name, email, phone_number) VALUES
(uuid_generate_v4(), 'John', 'Doe', 'john.doe@example.com', '555-123-4567'),
(uuid_generate_v4(), 'Jane', 'Smith', 'jane.smith@example.com', '555-234-5678'),
(uuid_generate_v4(), 'Michael', 'Johnson', 'michael.j@example.com', '555-345-6789'),
(uuid_generate_v4(), 'Emily', 'Williams', 'emily.w@example.com', '555-456-7890'),
(uuid_generate_v4(), 'David', 'Brown', 'david.b@example.com', '555-567-8901'),
(uuid_generate_v4(), 'Sarah', 'Davis', 'sarah.d@example.com', '555-678-9012'),
(uuid_generate_v4(), 'Alex', 'Miller', 'alex.m@example.com', '555-789-0123'),
(uuid_generate_v4(), 'Olivia', 'Wilson', 'olivia.w@example.com', '555-890-1234');

-- Store user IDs in variables for later use in user_artists
DO $$
DECLARE
    user1_id UUID;
    user2_id UUID;
    user3_id UUID;
    user4_id UUID;
    user5_id UUID;
    user6_id UUID;
    user7_id UUID;
    user8_id UUID;
    
    artist1_id INT;
    artist2_id INT;
    artist3_id INT;
    artist4_id INT;
    artist5_id INT;
    artist6_id INT;
    artist7_id INT;
    artist8_id INT;
    artist9_id INT;
    artist10_id INT;
    artist11_id INT;
    artist12_id INT;
    artist13_id INT;
    artist14_id INT;
    artist15_id INT;
BEGIN
    -- Get the user IDs that were just inserted
    SELECT user_id INTO user1_id FROM users WHERE email = 'john.doe@example.com';
    SELECT user_id INTO user2_id FROM users WHERE email = 'jane.smith@example.com';
    SELECT user_id INTO user3_id FROM users WHERE email = 'michael.j@example.com';
    SELECT user_id INTO user4_id FROM users WHERE email = 'emily.w@example.com';
    SELECT user_id INTO user5_id FROM users WHERE email = 'david.b@example.com';
    SELECT user_id INTO user6_id FROM users WHERE email = 'sarah.d@example.com';
    SELECT user_id INTO user7_id FROM users WHERE email = 'alex.m@example.com';
    SELECT user_id INTO user8_id FROM users WHERE email = 'olivia.w@example.com';
    
    -- Insert artists
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('4Z8W4fKeB5YxbusRsdQVPb', 'Radiohead') RETURNING artist_id INTO artist1_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('3WrFJ7ztbogyGnTHbHJFl2', 'The Beatles') RETURNING artist_id INTO artist2_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('0oSGxfWSnnOXhD2fKuz2Gy', 'David Bowie') RETURNING artist_id INTO artist3_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('36QJpDe2go2KgaRleHCDTp', 'Led Zeppelin') RETURNING artist_id INTO artist4_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('0L8ExT028jH3ddEcZwqJJ5', 'Red Hot Chili Peppers') RETURNING artist_id INTO artist5_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('6olE6TJLqED3rqDCT0FyPh', 'Nirvana') RETURNING artist_id INTO artist6_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('0YC192cP3KPCRWx8zr8MfZ', 'Taylor Swift') RETURNING artist_id INTO artist7_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('3TVXtAsR1Inumwj472S9r4', 'Drake') RETURNING artist_id INTO artist8_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('6eUKZXaKkcviH0Ku9w2n3V', 'Ed Sheeran') RETURNING artist_id INTO artist9_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('1uNFoZAHBGtllmzznpCI3s', 'Justin Bieber') RETURNING artist_id INTO artist10_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('06HL4z0CvFAxyc27GXpf02', 'Taylor Swift') RETURNING artist_id INTO artist11_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('1Xyo4u8uXC1ZmMpatF05PJ', 'The Weeknd') RETURNING artist_id INTO artist12_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('3Nrfpe0tUJi4K4DXYWgMUX', 'BTS') RETURNING artist_id INTO artist13_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('4gzpq5DPGxSnKTe4SA8HAU', 'Coldplay') RETURNING artist_id INTO artist14_id;
    INSERT INTO artists (spotify_artist_id, artist_name) VALUES
    ('6M2wZ9GZgrQXHCFfjv46we', 'Dua Lipa') RETURNING artist_id INTO artist15_id;
    
    -- Insert user-artist relationships with ranks
    -- User 1 likes rock and alternative
    INSERT INTO user_artists (user_id, artist_id, rank) VALUES
    (user1_id, artist1_id, 1),  -- Radiohead (top)
    (user1_id, artist2_id, 2),  -- The Beatles
    (user1_id, artist3_id, 3),  -- David Bowie
    (user1_id, artist4_id, 4),  -- Led Zeppelin
    (user1_id, artist5_id, 5),  -- Red Hot Chili Peppers
    (user1_id, artist6_id, 6);  -- Nirvana
    
    -- User 2 likes pop and some rock
    INSERT INTO user_artists (user_id, artist_id, rank) VALUES
    (user2_id, artist7_id, 1),  -- Taylor Swift (top)
    (user2_id, artist9_id, 2),  -- Ed Sheeran
    (user2_id, artist12_id, 3), -- The Weeknd
    (user2_id, artist14_id, 4), -- Coldplay
    (user2_id, artist2_id, 5);  -- The Beatles
    
    -- User 3 likes a mix of genres
    INSERT INTO user_artists (user_id, artist_id, rank) VALUES
    (user3_id, artist5_id, 1),  -- Red Hot Chili Peppers (top)
    (user3_id, artist8_id, 2),  -- Drake
    (user3_id, artist14_id, 3), -- Coldplay
    (user3_id, artist4_id, 4),  -- Led Zeppelin
    (user3_id, artist15_id, 5), -- Dua Lipa
    (user3_id, artist9_id, 6),  -- Ed Sheeran
    (user3_id, artist3_id, 7);  -- David Bowie
    
    -- User 4 likes pop
    INSERT INTO user_artists (user_id, artist_id, rank) VALUES
    (user4_id, artist7_id, 1),  -- Taylor Swift (top)
    (user4_id, artist15_id, 2), -- Dua Lipa
    (user4_id, artist12_id, 3), -- The Weeknd
    (user4_id, artist9_id, 4),  -- Ed Sheeran
    (user4_id, artist10_id, 5), -- Justin Bieber
    (user4_id, artist13_id, 6); -- BTS
    
    -- User 5 likes rock
    INSERT INTO user_artists (user_id, artist_id, rank) VALUES
    (user5_id, artist4_id, 1),  -- Led Zeppelin (top)
    (user5_id, artist1_id, 2),  -- Radiohead
    (user5_id, artist6_id, 3),  -- Nirvana
    (user5_id, artist2_id, 4),  -- The Beatles
    (user5_id, artist3_id, 5),  -- David Bowie
    (user5_id, artist5_id, 6);  -- Red Hot Chili Peppers
    
    -- User 6 likes pop and hip-hop
    INSERT INTO user_artists (user_id, artist_id, rank) VALUES
    (user6_id, artist8_id, 1),  -- Drake (top)
    (user6_id, artist12_id, 2), -- The Weeknd
    (user6_id, artist7_id, 3),  -- Taylor Swift
    (user6_id, artist10_id, 4), -- Justin Bieber
    (user6_id, artist15_id, 5); -- Dua Lipa
    
    -- User 7 likes alternative and some pop
    INSERT INTO user_artists (user_id, artist_id, rank) VALUES
    (user7_id, artist1_id, 1),  -- Radiohead (top)
    (user7_id, artist14_id, 2), -- Coldplay
    (user7_id, artist6_id, 3),  -- Nirvana
    (user7_id, artist3_id, 4),  -- David Bowie
    (user7_id, artist9_id, 5),  -- Ed Sheeran
    (user7_id, artist15_id, 6); -- Dua Lipa
    
    -- User 8 likes a mix of genres
    INSERT INTO user_artists (user_id, artist_id, rank) VALUES
    (user8_id, artist15_id, 1), -- Dua Lipa (top)
    (user8_id, artist14_id, 2), -- Coldplay
    (user8_id, artist2_id, 3),  -- The Beatles
    (user8_id, artist9_id, 4),  -- Ed Sheeran
    (user8_id, artist13_id, 5), -- BTS
    (user8_id, artist7_id, 6);  -- Taylor Swift
END $$;
