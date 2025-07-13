# Quick Guide: Disable Demo Mode

This is a quick reference for disabling demo mode in the Watered application for production deployments.

## TL;DR - How to Disable Demo Mode

**Demo mode is automatically disabled when you set Google OAuth credentials.**

### Using .env Files (Recommended)

```bash
# Create or edit .env file
cat > .env << 'EOF'
GOOGLE_CLIENT_ID=your-real-google-client-id
GOOGLE_CLIENT_SECRET=your-real-google-client-secret
SESSION_SECRET=your-secure-random-secret
ALLOWED_EMAILS=you@domain.com,partner@domain.com
ADMIN_EMAILS=you@domain.com
EOF

# Start the application (automatically loads .env)
just run
```

### For Local/Docker Production

```bash
# Set these environment variables
export GOOGLE_CLIENT_ID="your-real-google-client-id"
export GOOGLE_CLIENT_SECRET="your-real-google-client-secret"
export SESSION_SECRET="your-secure-random-secret"
export ALLOWED_EMAILS="you@domain.com,partner@domain.com"
export ADMIN_EMAILS="you@domain.com"

# Run the application
just docker-build-gcp
docker run -p 8080:8080 \
  -e GOOGLE_CLIENT_ID="$GOOGLE_CLIENT_ID" \
  -e GOOGLE_CLIENT_SECRET="$GOOGLE_CLIENT_SECRET" \
  -e SESSION_SECRET="$SESSION_SECRET" \
  -e ALLOWED_EMAILS="$ALLOWED_EMAILS" \
  -e ADMIN_EMAILS="$ADMIN_EMAILS" \
  watered:latest
```

### For Google Cloud Run

```bash
gcloud run deploy watered \
  --image us-central1-docker.pkg.dev/your-project/watered-repo/watered:latest \
  --set-env-vars GOOGLE_CLIENT_ID="your-client-id" \
  --set-env-vars GOOGLE_CLIENT_SECRET="your-client-secret" \
  --set-env-vars SESSION_SECRET="your-session-secret" \
  --set-env-vars ALLOWED_EMAILS="you@domain.com" \
  --set-env-vars ADMIN_EMAILS="you@domain.com"
```

## How It Works

### Demo Mode Enabled (Default)
When `GOOGLE_CLIENT_ID` or `GOOGLE_CLIENT_SECRET` are **not set**:
```
❌ Demo login available at /auth/demo-login
❌ Uses demo email allowlist
❌ Shows warnings in logs
❌ Not secure for production
```

### Production Mode Enabled
When `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` **are set**:
```
✅ Demo login returns 404 Not Found
✅ Requires real Google OAuth
✅ Uses your email allowlist  
✅ Ready for production
```

## Quick Test

```bash
# Test if demo mode is disabled
curl http://your-domain.com/auth/demo-login

# Demo mode disabled (production): 404 Not Found
# Demo mode enabled (development): Demo login form
```

## Getting Google OAuth Credentials

1. **Go to [Google Cloud Console](https://console.cloud.google.com/)**
2. **Create/Select Project**
3. **APIs & Services → OAuth consent screen**
   - Configure app details
   - Add your domain
4. **APIs & Services → Credentials**
   - Create OAuth 2.0 Client ID
   - Add redirect URI: `https://yourdomain.com/auth/callback`
5. **Copy Client ID and Secret**

## Environment Variables Summary

| Variable | Demo Mode | Production Mode |
|----------|-----------|-----------------|
| `GOOGLE_CLIENT_ID` | *(not set)* | `your-real-client-id` |
| `GOOGLE_CLIENT_SECRET` | *(not set)* | `your-real-secret` |
| `SESSION_SECRET` | `development-secret` | `secure-random-secret` |
| `ALLOWED_EMAILS` | `demo@example.com,test@example.com` | `you@domain.com,partner@domain.com` |
| `ADMIN_EMAILS` | `admin@example.com` | `you@domain.com` |

## Generate Secure Session Secret

```bash
# Generate a random session secret
openssl rand -base64 32

# Or use the just command
just generate-session-secret
```

## Troubleshooting

### Still seeing demo login?
- ✅ Check `GOOGLE_CLIENT_ID` is set and not empty
- ✅ Check `GOOGLE_CLIENT_SECRET` is set and not empty
- ✅ Restart your application after setting variables
- ✅ Check logs for "Demo mode enabled" warnings

### OAuth errors?
- ✅ Verify redirect URI matches exactly: `https://yourdomain.com/auth/callback`
- ✅ Check OAuth consent screen is configured
- ✅ Ensure your email is in `ALLOWED_EMAILS`

### Users can't log in?
- ✅ Add their emails to `ALLOWED_EMAILS`
- ✅ Check Google OAuth consent screen approval
- ✅ Verify OAuth credentials are correct

For detailed setup instructions, see:
- [Production Setup Guide](./production-setup.md)
- [Google Cloud Setup Guide](./gcp-setup.md)