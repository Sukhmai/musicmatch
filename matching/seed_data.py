import psycopg2
import uuid
import csv
import os
import random
from psycopg2.extras import execute_values
from faker import Faker

# Get database parameters from environment variables with defaults
db_name = os.environ.get("DB_NAME", "spotify")
db_user = os.environ.get("DB_USERNAME", "spotifyuser")
db_host = os.environ.get("DB_HOST", "localhost")
db_port = os.environ.get("DB_PORT", "5432")

# Password must be provided via environment variable
db_password = os.environ.get("DB_PASSWORD")
if not db_password:
    raise ValueError("DB_PASSWORD environment variable not set")

DB_PARAMS = f"dbname={db_name} user={db_user} password={db_password} host={db_host} port={db_port}"

# Constants for data generation
NUM_USERS = 100
ARTISTS_PER_USER = 20
CSV_DIR = "seed_data"

# Initialize Faker for generating realistic user data
fake = Faker()
Faker.seed(42)  # For reproducibility

# Expanded list of artists (50 artists to ensure variety)
ARTISTS = [
    {"spotify_artist_id": "4Z8W4fKeB5YxbusRsdQVPb", "artist_name": "Radiohead"},
    {"spotify_artist_id": "3WrFJ7ztbogyGnTHbHJFl2", "artist_name": "The Beatles"},
    {"spotify_artist_id": "0oSGxfWSnnOXhD2fKuz2Gy", "artist_name": "David Bowie"},
    {"spotify_artist_id": "36QJpDe2go2KgaRleHCDTp", "artist_name": "Led Zeppelin"},
    {"spotify_artist_id": "0L8ExT028jH3ddEcZwqJJ5", "artist_name": "Red Hot Chili Peppers"},
    {"spotify_artist_id": "6olE6TJLqED3rqDCT0FyPh", "artist_name": "Nirvana"},
    {"spotify_artist_id": "0YC192cP3KPCRWx8zr8MfZ", "artist_name": "Taylor Swift"},
    {"spotify_artist_id": "3TVXtAsR1Inumwj472S9r4", "artist_name": "Drake"},
    {"spotify_artist_id": "6eUKZXaKkcviH0Ku9w2n3V", "artist_name": "Ed Sheeran"},
    {"spotify_artist_id": "1uNFoZAHBGtllmzznpCI3s", "artist_name": "Justin Bieber"},
    {"spotify_artist_id": "1Xyo4u8uXC1ZmMpatF05PJ", "artist_name": "The Weeknd"},
    {"spotify_artist_id": "3Nrfpe0tUJi4K4DXYWgMUX", "artist_name": "BTS"},
    {"spotify_artist_id": "4gzpq5DPGxSnKTe4SA8HAU", "artist_name": "Coldplay"},
    {"spotify_artist_id": "6M2wZ9GZgrQXHCFfjv46we", "artist_name": "Dua Lipa"},
    {"spotify_artist_id": "66CXWjxzNUsdJxJ2JdwvnR", "artist_name": "Ariana Grande"},
    {"spotify_artist_id": "6vWDO969PvNqNYHIOW5v0m", "artist_name": "Beyoncé"},
    {"spotify_artist_id": "0du5cEVh5yTK9QJze8zA0C", "artist_name": "Bruno Mars"},
    {"spotify_artist_id": "7jVv8c5Fj3E9VhNjxT4snq", "artist_name": "Lil Nas X"},
    {"spotify_artist_id": "4MCBfE4596Uoi2O4DtmEMz", "artist_name": "Juice WRLD"},
    {"spotify_artist_id": "4O15NlyKLIASxsJ0PrXPfz", "artist_name": "Lil Uzi Vert"},
    {"spotify_artist_id": "4q3ewBCX7sLwd24euuV69X", "artist_name": "Bad Bunny"},
    {"spotify_artist_id": "1McMsnEElThX1knmY4oliG", "artist_name": "Olivia Rodrigo"},
    {"spotify_artist_id": "2YZyLoL8N0Wb9xBt1NhZWg", "artist_name": "Kendrick Lamar"},
    {"spotify_artist_id": "4dpARuHxo51G3z768sgnrY", "artist_name": "Adele"},
    {"spotify_artist_id": "5pKCCKE2ajJHZ9KAiaK11H", "artist_name": "Rihanna"},
    {"spotify_artist_id": "246dkjvS1zLTtiykXe5h60", "artist_name": "Post Malone"},
    {"spotify_artist_id": "7dGJo4pcD2V6oG8kP0tJRR", "artist_name": "Eminem"},
    {"spotify_artist_id": "5K4W6rqBFWDnAN6FQUkS6x", "artist_name": "Kanye West"},
    {"spotify_artist_id": "1Cs0zKBU1kc0i8ypK3B9ai", "artist_name": "David Guetta"},
    {"spotify_artist_id": "0C0XlULifJtAgn6ZNCW2eu", "artist_name": "The Killers"},
    {"spotify_artist_id": "53XhwfbYqKCa1cC15pYq2q", "artist_name": "Imagine Dragons"},
    {"spotify_artist_id": "64KEffDW9EtZ1y2vBYgq8T", "artist_name": "Marshmello"},
    {"spotify_artist_id": "6qqNVTkY8uBg9cP3Jd7DAH", "artist_name": "Billie Eilish"},
    {"spotify_artist_id": "0hCNtLu0JehylgoiP8L4Gh", "artist_name": "Nicki Minaj"},
    {"spotify_artist_id": "4AK6F7OLvEQ5QYCBNiQWHq", "artist_name": "One Direction"},
    {"spotify_artist_id": "1ukmGETCwXTbgrTrA2JXP8", "artist_name": "The Rolling Stones"},
    {"spotify_artist_id": "0TnOYISbd1XYRBk9myaseg", "artist_name": "Pitbull"},
    {"spotify_artist_id": "5Pwc4xIPtQLFEnJriah9YJ", "artist_name": "OneRepublic"},
    {"spotify_artist_id": "0LyfQWJT6nXafLPZqxe9Of", "artist_name": "Various Artists"},
    {"spotify_artist_id": "2wY79sveU1sp5g7SokKOiI", "artist_name": "Sam Smith"},
    {"spotify_artist_id": "3fMbdgg4jU18AjLCKBhRSm", "artist_name": "Michael Jackson"},
    {"spotify_artist_id": "2ye2Wgw4gimLv2eAKyk1NB", "artist_name": "Metallica"},
    {"spotify_artist_id": "6XyY86QOPPrYVGvF9ch6wz", "artist_name": "Linkin Park"},
    {"spotify_artist_id": "0Y5tJX1MQlPlqiwlOH1tJY", "artist_name": "Travis Scott"},
    {"spotify_artist_id": "5YGY8feqx7naU7z4HrwZM6", "artist_name": "Miley Cyrus"},
    {"spotify_artist_id": "4NHQUGzhtTLFvgF5SZesLK", "artist_name": "Tame Impala"},
    {"spotify_artist_id": "6KImCVD70vtIoJWnq6nGn3", "artist_name": "Harry Styles"},
    {"spotify_artist_id": "0EmeFodog0BfCgMzAIvKQp", "artist_name": "Shakira"},
    {"spotify_artist_id": "1Xylc3o4UrD53lo9CvFvVg", "artist_name": "Foo Fighters"},
    {"spotify_artist_id": "7CajNmpbOovFoOoasH2HaY", "artist_name": "Calvin Harris"},
]

def generate_users(num_users):
    """Generate a list of unique users with realistic data"""
    users = []
    for _ in range(num_users):
        first_name = fake.first_name()
        last_name = fake.last_name()
        email = f"{first_name.lower()}.{last_name.lower()}@{fake.domain_name()}"
        phone_number = fake.phone_number()
        
        users.append({
            "first_name": first_name,
            "last_name": last_name,
            "email": email,
            "phone_number": phone_number
        })
    return users

def generate_user_artists(user_ids, artist_ids, artists_per_user=ARTISTS_PER_USER):
    """Generate user-artist relationships with each user having exactly the specified number of artists"""
    random.seed(42)  # For reproducibility
    
    user_artists = []
    for user_id in user_ids:
        # Select exactly artists_per_user random artists for this user
        user_artist_ids = random.sample(artist_ids, artists_per_user)
        
        # Assign ranks (1 to artists_per_user)
        for rank, artist_id in enumerate(user_artist_ids, 1):
            user_artists.append({
                "user_id": user_id,
                "artist_id": artist_id,
                "rank": rank
            })
    return user_artists

def create_csv_files(users, artists, user_artists):
    """Create CSV files for users, artists, and user-artist relationships"""
    # Create directory if it doesn't exist
    os.makedirs(CSV_DIR, exist_ok=True)
    
    # Write users to CSV
    with open(f"{CSV_DIR}/users.csv", 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=["first_name", "last_name", "email", "phone_number"])
        writer.writeheader()
        writer.writerows(users)
    
    # Write artists to CSV
    with open(f"{CSV_DIR}/artists.csv", 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=["spotify_artist_id", "artist_name"])
        writer.writeheader()
        writer.writerows(artists)
    
    # Write user-artist relationships to CSV
    with open(f"{CSV_DIR}/user_artists.csv", 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=["user_id", "artist_id", "rank"])
        writer.writeheader()
        writer.writerows(user_artists)
    
    print(f"CSV files created in {CSV_DIR} directory")

def seed_database_from_csv():
    """Seed the database with data from CSV files"""
    conn = None
    try:
        # Connect to the database
        print("Connecting to the PostgreSQL database...")
        conn = psycopg2.connect(DB_PARAMS)
        cur = conn.cursor()
        
        # Clear existing data
        print("Clearing existing data...")
        cur.execute("TRUNCATE users, artists, user_artists CASCADE;")
        
        # Insert users from CSV
        print("Inserting users from CSV...")
        user_ids = []
        with open(f"{CSV_DIR}/users.csv", 'r', newline='') as f:
            reader = csv.DictReader(f)
            for user in reader:
                cur.execute(
                    """
                    INSERT INTO users (first_name, last_name, email, phone_number)
                    VALUES (%s, %s, %s, %s)
                    RETURNING user_id
                    """,
                    (user["first_name"], user["last_name"], user["email"], user["phone_number"])
                )
                user_id = cur.fetchone()[0]
                user_ids.append(user_id)
        
        # Insert artists from CSV
        print("Inserting artists from CSV...")
        artist_ids = []
        with open(f"{CSV_DIR}/artists.csv", 'r', newline='') as f:
            reader = csv.DictReader(f)
            for artist in reader:
                cur.execute(
                    """
                    INSERT INTO artists (spotify_artist_id, artist_name)
                    VALUES (%s, %s)
                    RETURNING artist_id
                    """,
                    (artist["spotify_artist_id"], artist["artist_name"])
                )
                artist_id = cur.fetchone()[0]
                artist_ids.append(artist_id)
        
        # Insert user-artist relationships from CSV
        print("Inserting user-artist relationships from CSV...")
        user_artists_data = []
        with open(f"{CSV_DIR}/user_artists.csv", 'r', newline='') as f:
            reader = csv.DictReader(f)
            for i, ua in enumerate(reader):
                user_artists_data.append((
                    user_ids[int(ua["user_id"]) - 1],  # Convert from 1-indexed in CSV to actual user_id
                    artist_ids[int(ua["artist_id"]) - 1],  # Convert from 1-indexed in CSV to actual artist_id
                    int(ua["rank"])
                ))
        
        print(f"Inserting {len(user_artists_data)} user-artist relationships...")
        execute_values(
            cur,
            """
            INSERT INTO user_artists (user_id, artist_id, rank)
            VALUES %s
            """,
            user_artists_data
        )
        
        # Commit the transaction
        conn.commit()
        print("Database seeded successfully!")
        
    except (Exception, psycopg2.DatabaseError) as error:
        print(f"Error: {error}")
        if conn:
            conn.rollback()
    finally:
        if conn:
            cur.close()
            conn.close()
            print("Database connection closed.")

def seed_database():
    """Generate data and seed the database"""
    # Generate users
    users = generate_users(NUM_USERS)
    
    # Create CSV files
    create_csv_files(users, ARTISTS, [])
    
    # Seed database from CSV files
    conn = None
    try:
        # Connect to the database
        print("Connecting to the PostgreSQL database...")
        conn = psycopg2.connect(DB_PARAMS)
        cur = conn.cursor()
        
        # Clear existing data
        print("Clearing existing data...")
        cur.execute("TRUNCATE users, artists, user_artists CASCADE;")
        
        # Insert users
        print(f"Inserting {NUM_USERS} users...")
        user_ids = []
        for user in users:
            cur.execute(
                """
                INSERT INTO users (first_name, last_name, email, phone_number)
                VALUES (%s, %s, %s, %s)
                RETURNING user_id
                """,
                (user["first_name"], user["last_name"], user["email"], user["phone_number"])
            )
            user_id = cur.fetchone()[0]
            user_ids.append(user_id)
        
        # Insert artists
        print(f"Inserting {len(ARTISTS)} artists...")
        artist_ids = []
        for artist in ARTISTS:
            cur.execute(
                """
                INSERT INTO artists (spotify_artist_id, artist_name)
                VALUES (%s, %s)
                RETURNING artist_id
                """,
                (artist["spotify_artist_id"], artist["artist_name"])
            )
            artist_id = cur.fetchone()[0]
            artist_ids.append(artist_id)
        
        # Generate and insert user-artist relationships
        print("Generating user-artist relationships...")
        user_artists = generate_user_artists(user_ids, artist_ids, ARTISTS_PER_USER)
        
        # Save user-artist relationships to CSV for reference
        with open(f"{CSV_DIR}/user_artists.csv", 'w', newline='') as f:
            writer = csv.DictWriter(f, fieldnames=["user_id", "artist_id", "rank"])
            writer.writeheader()
            writer.writerows(user_artists)
        
        print(f"Inserting {len(user_artists)} user-artist relationships...")
        execute_values(
            cur,
            """
            INSERT INTO user_artists (user_id, artist_id, rank)
            VALUES %s
            """,
            [(ua["user_id"], ua["artist_id"], ua["rank"]) for ua in user_artists]
        )
        
        # Commit the transaction
        conn.commit()
        print("Database seeded successfully!")
        print(f"Each of the {NUM_USERS} users has exactly {ARTISTS_PER_USER} artists")
        
    except (Exception, psycopg2.DatabaseError) as error:
        print(f"Error: {error}")
        if conn:
            conn.rollback()
    finally:
        if conn:
            cur.close()
            conn.close()
            print("Database connection closed.")

if __name__ == "__main__":
    seed_database()
