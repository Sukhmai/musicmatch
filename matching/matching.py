import psycopg2
import rustworkx as rx
import numpy as np
from scipy.sparse import csr_matrix
from sklearn.metrics.pairwise import cosine_similarity
import csv
import os
import math
from datetime import datetime

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
CSV_DIR="match_results"

def calculate_match_score(cosine_sim):
    """
    Transform cosine similarity to a user-friendly match score (0-100)
    using a sigmoid function that maps low similarities to good scores.
    
    Args:
        cosine_sim: Raw cosine similarity value (0-1)
        
    Returns:
        An integer score from 0-100
    """
    # Sigmoid function parameters
    steepness = 15  # Controls how steep the curve is
    midpoint = 0.15  # Value that maps to 75
    
    # Apply sigmoid function: 100 / (1 + e^(-steepness * (x - midpoint)))
    score = 100 / (1 + math.exp(-steepness * (cosine_sim - midpoint)))
    
    return round(score)

# Connect to PostgreSQL
conn = psycopg2.connect(DB_PARAMS)
cur = conn.cursor()

# Fetch the maximum artist ID
cur.execute("SELECT MAX(artist_id) FROM artists;")
max_artist_id = cur.fetchone()[0]

# Fetch user details
cur.execute("""
    SELECT user_id, first_name, last_name, email, phone_number
    FROM users
""")
user_details = {user_id: (first_name, last_name, email, phone_number) for user_id, first_name, last_name, email, phone_number in cur.fetchall()}

# Fetch user-artist data
cur.execute("""
    SELECT u.user_id, ua.artist_id, ua.rank
    FROM users u
    JOIN user_artists ua ON u.user_id = ua.user_id
""")
rows = cur.fetchall()

# Process data into user vectors
user_vectors = {}  # {user_id: {artist_id: rank}}
for user_id, artist_id, rank in rows:
    if user_id not in user_vectors:
        user_vectors[user_id] = {}
    user_vectors[user_id][artist_id] = rank

# Convert to a sparse matrix
user_ids = list(user_vectors.keys())
num_users = len(user_ids)
num_artists = max_artist_id + 1

# Prepare data for sparse matrix
data = []
row_ind = []
col_ind = []
for i, user_id in enumerate(user_ids):
    for artist_id, rank in user_vectors[user_id].items():
        row_ind.append(i)
        col_ind.append(artist_id)
        data.append(rank)

# Create sparse matrix
matrix = csr_matrix((data, (row_ind, col_ind)), shape=(num_users, num_artists))

# Compute cosine similarity on the sparse matrix
similarity_matrix = cosine_similarity(matrix)

# Build graph in Rustworkx
graph = rx.PyGraph()
graph.add_nodes_from(user_ids)

# Create a mapping from user_id to node index
node_map = {}
for i, user_id in enumerate(user_ids):
    node_index = graph.add_node(user_id)  # Add node with user_id as data
    node_map[user_id] = node_index

# Then use the node indices when adding edges
for i, user1 in enumerate(user_ids):
    for j, user2 in enumerate(user_ids):
        if i < j:
            similarity = similarity_matrix[i, j]
            graph.add_edge(node_map[user1], node_map[user2], similarity)

# Find matching
matching = rx.max_weight_matching(graph, max_cardinality=True, weight_fn=lambda x: int(x))
print("Matched pairs:", list(matching))

# Fetch all artist names at once (more efficient than querying in a loop)
cur.execute("SELECT artist_id, artist_name FROM artists")
artist_names = {artist_id: name for artist_id, name in cur.fetchall()}

# For each matched pair, print similarity and common artists
for edge in matching:
    user1_id = graph[edge[0]]
    user2_id = graph[edge[1]]
    user1_idx = user_ids.index(user1_id)
    user2_idx = user_ids.index(user2_id)
    
    # Get user details
    user1_first, user1_last, user1_email, user1_phone = user_details.get(user1_id, ("Unknown", "User", "unknown@email.com", None))
    user2_first, user2_last, user2_email, user2_phone = user_details.get(user2_id, ("Unknown", "User", "unknown@email.com", None))
    
    # Get similarity score
    similarity_score = similarity_matrix[user1_idx, user2_idx]
    
    # Find common artists
    user1_artists = set(user_vectors[user1_id].keys())
    user2_artists = set(user_vectors[user2_id].keys())
    common_artists = user1_artists.intersection(user2_artists)
    
    # Calculate match score
    match_score = calculate_match_score(similarity_score)
    
    # Print detailed information
    print(f"\nMatch: {user1_first} {user1_last} ({user1_email}, {user1_phone or 'No phone'}) and {user2_first} {user2_last} ({user2_email}, {user2_phone or 'No phone'})")
    print(f"User IDs: {user1_id} and {user2_id}")
    print(f"Cosine Similarity: {similarity_score:.4f} (Match Score: {match_score}/100)")
    print(f"Common Artists ({len(common_artists)}):")
    for artist_id in common_artists:
        artist_name = artist_names.get(artist_id, "Unknown")
        print(f"  - {artist_name} (ID: {artist_id})")

# Create a directory for match results if it doesn't exist
os.makedirs(CSV_DIR, exist_ok=True)

# Generate a timestamp for the filename
timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
csv_filename = f"{CSV_DIR}/matches_{timestamp}.csv"

# Write match data to CSV
with open(csv_filename, 'w', newline='') as csvfile:
    csvwriter = csv.writer(csvfile)
    
    # Write header
    csvwriter.writerow([
        'user1_id', 'user1_first_name', 'user1_last_name', 'user1_email', 'user1_phone',
        'user2_id', 'user2_first_name', 'user2_last_name', 'user2_email', 'user2_phone',
        'similarity_score', 'match_score', 'common_artists'
    ])
    
    # Write each match
    for edge in matching:
        user1_id = graph[edge[0]]
        user2_id = graph[edge[1]]
        user1_idx = user_ids.index(user1_id)
        user2_idx = user_ids.index(user2_id)
        
        # Get user details
        user1_first, user1_last, user1_email, user1_phone = user_details.get(user1_id, ("Unknown", "User", "unknown@email.com", None))
        user2_first, user2_last, user2_email, user2_phone = user_details.get(user2_id, ("Unknown", "User", "unknown@email.com", None))
        
        # Get similarity score
        similarity_score = similarity_matrix[user1_idx, user2_idx]
        
        # Find common artists
        user1_artists = set(user_vectors[user1_id].keys())
        user2_artists = set(user_vectors[user2_id].keys())
        common_artists = user1_artists.intersection(user2_artists)
        
        # Format common artists as a string
        common_artists_str = "|".join([artist_names.get(artist_id, "Unknown") for artist_id in common_artists])
        
        # Calculate match score
        match_score = calculate_match_score(similarity_score)
        
        # Write to CSV
        csvwriter.writerow([
            user1_id, user1_first, user1_last, user1_email, user1_phone or "",
            user2_id, user2_first, user2_last, user2_email, user2_phone or "",
            similarity_score, match_score, common_artists_str
        ])

print(f"\nMatch data exported to {csv_filename}")

# Close the database connection
cur.close()
conn.close()
