#!/usr/bin/env python3
import csv
import os
import argparse
import requests
from typing import List, Dict, Any, Tuple

def read_matches_csv(csv_path: str) -> List[Dict[str, Any]]:
    """
    Read the matches CSV file and return a list of match dictionaries.
    """
    matches = []
    with open(csv_path, 'r', newline='') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            matches.append(row)
    return matches

def prepare_email_pairs(matches: List[Dict[str, Any]]) -> List[Tuple[Dict[str, Any], Dict[str, Any]]]:
    """
    Prepare email data for each user in the match pairs.
    Returns a list of tuples, each containing email data for both users in a match.
    """
    email_pairs = []
    
    for match in matches:
        # Extract data for first user
        user1_data = {
            'email': match['user1_email'],
            'first_name': match['user1_first_name'],
            'last_name': match['user1_last_name'],
            'match_first_name': match['user2_first_name'],
            'match_last_name': match['user2_last_name'],
            'match_email': match['user2_email'],
            'match_phone': match['user2_phone'],
            'similarity_score': float(match['similarity_score']),
            'match_score': int(match['match_score']),
            'common_artists': match['common_artists'].split('|')
        }
        
        # Extract data for second user
        user2_data = {
            'email': match['user2_email'],
            'first_name': match['user2_first_name'],
            'last_name': match['user2_last_name'],
            'match_first_name': match['user1_first_name'],
            'match_last_name': match['user1_last_name'],
            'match_email': match['user1_email'],
            'match_phone': match['user1_phone'],
            'similarity_score': float(match['similarity_score']),
            'match_score': int(match['match_score']),
            'common_artists': match['common_artists'].split('|')
        }
        
        email_pairs.append((user1_data, user2_data))
    
    return email_pairs

def format_email_content(user_data: Dict[str, Any], template_path: str = None) -> str:
    """
    Format the email content using the user data.
    If a template path is provided, it will use that template.
    Otherwise, it will use a default template.
    """
    # Format common artists as a list with bullet points
    common_artists_list = "\n".join([f"- {artist}" for artist in user_data['common_artists']])
    
    # Create a copy of user_data with formatted common_artists
    template_data = user_data.copy()
    template_data['common_artists'] = common_artists_list
    
    # Handle empty phone numbers for template
    if not template_data.get('match_phone') or not template_data['match_phone'].strip():
        template_data['match_phone'] = "Not provided"
    
    if template_path and os.path.exists(template_path):
        with open(template_path, 'r') as f:
            template = f.read()
            # Replace placeholders in the template
            return template.format(**template_data)
    return "No email"

def send_email_mailgun(
    recipient: str, 
    subject: str, 
    body: str, 
    api_key: str, 
    domain: str, 
    sender: str,
    test_mode: bool = False
) -> Dict[str, Any]:
    """
    Send an email using the Mailgun API.
    
    Args:
        recipient: Email address of the recipient
        subject: Email subject
        body: Email body content
        api_key: Mailgun API key
        domain: Verified Mailgun domain
        sender: Sender email address (must be from the verified domain)
        test_mode: If True, don't actually send the email
        
    Returns:
        Response from the Mailgun API
    """
    if test_mode:
        print(f"TEST MODE: Would send email to {recipient}")
        print(f"Subject: {subject}")
        print(f"Body: {body}")
        return {"id": "test", "message": "Test mode, no email sent"}
    
    return requests.post(
        f"https://api.mailgun.net/v3/{domain}/messages",
        auth=("api", api_key),
        data={
            "from": sender,
            "to": recipient,
            "subject": subject,
            "text": body
        }
    ).json()

def main():
    parser = argparse.ArgumentParser(description='Send match notification emails to users')
    parser.add_argument('csv_file', help='Path to the matches CSV file')
    parser.add_argument('--api-key', required=True, help='Mailgun API key')
    parser.add_argument('--domain', required=True, help='Mailgun domain')
    parser.add_argument('--sender', required=True, help='Sender email address')
    parser.add_argument('--template', required=True, help='Path to email template file')
    parser.add_argument('--subject', default='Your Musical Match!', help='Email subject')
    parser.add_argument('--test', action='store_true', help='Test mode - do not send actual emails')
    parser.add_argument('--limit', type=int, help='Limit the number of emails to send')
    
    args = parser.parse_args()
    
    # Read matches from CSV
    matches = read_matches_csv(args.csv_file)
    print(f"Read {len(matches)} matches from {args.csv_file}")
    
    # Prepare email data
    email_pairs = prepare_email_pairs(matches)
    
    # Apply limit if specified
    if args.limit and args.limit > 0:
        email_pairs = email_pairs[:args.limit]
        print(f"Limiting to {args.limit} matches")
    
    # Send emails
    emails_sent = 0
    for user1_data, user2_data in email_pairs:
        # Send email to first user
        user1_body = format_email_content(user1_data, args.template)
        user1_response = send_email_mailgun(
            user1_data['email'],
            args.subject,
            user1_body,
            args.api_key,
            args.domain,
            args.sender,
            args.test
        )
        print(f"Email to {user1_data['email']}: {user1_response.get('message', 'Sent')}")
        emails_sent += 1
        
        # Send email to second user
        user2_body = format_email_content(user2_data, args.template)
        user2_response = send_email_mailgun(
            user2_data['email'],
            args.subject,
            user2_body,
            args.api_key,
            args.domain,
            args.sender,
            args.test
        )
        print(f"Email to {user2_data['email']}: {user2_response.get('message', 'Sent')}")
        emails_sent += 1
    
    print(f"\nTotal emails sent: {emails_sent}")

if __name__ == "__main__":
    main()
