# Spotify Match Email Notification System

This system allows you to match users based on their music preferences and send personalized email notifications to matched pairs.

## Overview

The system consists of two main components:

1. **matching.py**: Matches users based on their music preferences and exports the matches to a CSV file.
2. **send_match_emails.py**: Reads the match data from the CSV file and sends personalized emails to matched users.

## Prerequisites

- Python 3.6+
- PostgreSQL database with user and artist data
- Mailgun account (for sending emails)
- Required Python packages: `psycopg2`, `rustworkx`, `numpy`, `scipy`, `scikit-learn`, `requests`

## Step 1: Generate Match Data

Run the matching script to generate pairs of users based on their music preferences:

```bash
python matching/matching.py
```

This will:
1. Connect to your PostgreSQL database
2. Match users based on their artist preferences
3. Export the matches to a CSV file in the `matching/match_results/` directory
4. Print the path to the generated CSV file

## Step 2: Set Up Mailgun

Before sending emails, you need to set up a Mailgun account:

1. Sign up for a Mailgun account at [mailgun.com](https://www.mailgun.com/)
2. Add and verify your domain in the Mailgun dashboard
3. Get your API key from the dashboard
4. Make sure your sender email address uses the verified domain

## Step 3: Send Match Notification Emails

Use the email script to send personalized notifications to matched users:

```bash
python matching/send_match_emails.py path/to/matches.csv \
  --api-key YOUR_MAILGUN_API_KEY \
  --domain YOUR_MAILGUN_DOMAIN \
  --sender "Spotify Match <matches@yourdomain.com>"
```

### Command-Line Options

- `csv_file`: Path to the matches CSV file (required)
- `--api-key`: Mailgun API key (required)
- `--domain`: Mailgun domain (required)
- `--sender`: Sender email address (required)
- `--template`: Path to a custom email template file (optional)
- `--subject`: Email subject line (default: "Your Spotify Match!")
- `--test`: Run in test mode without sending actual emails (optional)
- `--limit`: Limit the number of matches to process (optional)

### Custom Email Templates

You can create a custom email template file to personalize the email content. The template uses Python's string formatting with the following variables:

- `{first_name}`: Recipient's first name
- `{last_name}`: Recipient's last name
- `{match_first_name}`: Match's first name
- `{match_last_name}`: Match's last name
- `{match_email}`: Match's email address
- `{match_phone}`: Match's phone number (if available)
- `{similarity_score}`: Raw similarity score between the users (0-1 scale)
- `{match_score}`: User-friendly match score (0-100 scale)
- `{common_artists}`: List of common artists

Example template file (email_template.txt):
```
Hello {first_name},

We've found a match for you based on your music taste!

You've been matched with {match_first_name} {match_last_name}.
Your match score is {match_score}/100.

CONTACT INFORMATION:
Email: {match_email}
Phone: {match_phone}

You both enjoy these artists:
{common_artists}

Best regards,
The Spotify Match Team
```

To use a custom template:
```bash
python matching/send_match_emails.py path/to/matches.csv \
  --api-key YOUR_API_KEY \
  --domain YOUR_DOMAIN \
  --sender "Spotify Match <matches@yourdomain.com>" \
  --template matching/email_template.txt
```

## Example Usage Scenarios

### Test Mode

To test the email sending without actually sending emails:

```bash
python matching/send_match_emails.py path/to/matches.csv \
  --api-key YOUR_API_KEY \
  --domain YOUR_DOMAIN \
  --sender "Spotify Match <matches@yourdomain.com>" \
  --test
```

### Limit the Number of Emails

To limit the number of matches (useful for testing or batch processing):

```bash
python matching/send_match_emails.py path/to/matches.csv \
  --api-key YOUR_API_KEY \
  --domain YOUR_DOMAIN \
  --sender "Spotify Match <matches@yourdomain.com>" \
  --limit 5
```

### Custom Subject Line

To use a custom email subject:

```bash
python matching/send_match_emails.py path/to/matches.csv \
  --api-key YOUR_API_KEY \
  --domain YOUR_DOMAIN \
  --sender "Spotify Match <matches@yourdomain.com>" \
  --subject "We Found Your Music Soulmate!"
```

## Troubleshooting

- **CSV Format Issues**: Ensure the CSV file has the correct headers and format
- **Email Sending Failures**: Check your Mailgun API key, domain verification, and sender email format
- **Template Errors**: Verify that your custom template includes all required variables
