# Google OAuth2 Setup Guide

This guide will help you set up Google OAuth2 credentials for the Watered plant tracking app.

## Prerequisites

- Google account
- Access to [Google Cloud Console](https://console.cloud.google.com/)

## Step 1: Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click on the project dropdown at the top
3. Click "New Project"
4. Enter project name: `watered-plant-tracker` (or your preferred name)
5. Click "Create"

## Step 2: Enable Google+ API

1. In your project, go to "APIs & Services" > "Library"
2. Search for "Google+ API" 
3. Click on it and click "Enable"
4. Also enable "Google OAuth2 API" if available

## Step 3: Configure OAuth Consent Screen

1. Go to "APIs & Services" > "OAuth consent screen"
2. Choose "External" user type (unless you have Google Workspace)
3. Fill in the required information:
   - **App name**: `Watered Plant Tracker`
   - **User support email**: Your email
   - **App domain**: Leave blank for now
   - **Developer contact information**: Your email
4. Click "Save and Continue"
5. **Scopes**: Click "Add or Remove Scopes"
   - Add: `userinfo.email` 
   - Add: `userinfo.profile`
6. Click "Save and Continue"
7. **Test users**: Add your email and your wife's email
8. Click "Save and Continue"

## Step 4: Create OAuth2 Credentials

1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "OAuth client ID"
3. Choose "Web application"
4. **Name**: `Watered Web Client`
5. **Authorized redirect URIs**: Add these URLs:
   - `http://localhost:8080/auth/callback` (for development)
   - `https://yourdomain.com/auth/callback` (for production)
6. Click "Create"
7. **Important**: Copy the Client ID and Client Secret

## Step 5: Configure Environment Variables

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Update the `.env` file with your credentials:
   ```bash
   # Google OAuth2 Configuration
   GOOGLE_CLIENT_ID=your-actual-client-id-here
   GOOGLE_CLIENT_SECRET=your-actual-client-secret-here
   
   # Session Security
   SESSION_SECRET=your-random-long-secret-key-here
   
   # User Access Control
   ALLOWED_EMAILS=you@gmail.com,yourwife@gmail.com
   ADMIN_EMAILS=you@gmail.com
   ```

3. Generate a secure session secret:
   ```bash
   # On macOS/Linux:
   openssl rand -base64 32
   
   # Or use any random string generator
   ```

## Step 6: Load Environment Variables

### Option A: Using a .env file (recommended)

Install a .env loader for Go:
```bash
go get github.com/joho/godotenv
```

Then update your main.go to load the .env file.

### Option B: Export manually

```bash
export GOOGLE_CLIENT_ID="your-client-id"
export GOOGLE_CLIENT_SECRET="your-client-secret"
export SESSION_SECRET="your-session-secret"
export ALLOWED_EMAILS="you@gmail.com,yourwife@gmail.com"
export ADMIN_EMAILS="you@gmail.com"
```

## Step 7: Test the Setup

1. Start your server:
   ```bash
   go run cmd/server/main.go
   ```

2. You should see:
   ```
   Starting server on port 8080
   ```
   (No more warning about demo mode)

3. Visit `http://localhost:8080/login`
4. Click "Sign in with Google"
5. You should be redirected to Google's real OAuth consent screen
6. After authorizing, you'll be redirected back to your app

## Security Notes

- **Never commit `.env` files** - they're already in `.gitignore`
- **Use strong session secrets** in production
- **Enable HTTPS** in production and update cookie settings
- **Regularly rotate** your OAuth2 credentials
- **Review authorized emails** periodically

## Troubleshooting

### "The OAuth client was not found"
- Double-check your Client ID and Client Secret
- Ensure the redirect URI matches exactly (including http vs https)

### "Access blocked: Authorization Error"
- Make sure your email is in the ALLOWED_EMAILS list
- Check that your email is added as a test user in the OAuth consent screen

### "redirect_uri_mismatch"
- Verify the redirect URI in Google Cloud Console matches your server URL
- For local development, use `http://localhost:8080/auth/callback`

## Production Deployment

When deploying to production:

1. Update redirect URIs in Google Cloud Console
2. Set `Secure: true` in session cookie settings  
3. Use environment variables (not .env files)
4. Enable HTTPS
5. Use a production-grade session secret

## Need Help?

- [Google OAuth2 Documentation](https://developers.google.com/identity/protocols/oauth2)
- [Google Cloud Console](https://console.cloud.google.com/)
- Check the server logs for detailed error messages