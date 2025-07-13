# Production Setup Guide

This guide explains how to configure the Watered application for production deployment, specifically focusing on disabling demo mode and setting up proper authentication.

## Table of Contents

- [Demo Mode vs Production Mode](#demo-mode-vs-production-mode)
- [Google OAuth Setup for Production](#google-oauth-setup-for-production)
- [Environment Variables for Production](#environment-variables-for-production)
- [Docker Production Configuration](#docker-production-configuration)
- [Google Cloud Deployment](#google-cloud-deployment)
- [Security Considerations](#security-considerations)

## Demo Mode vs Production Mode

### Demo Mode (Development)
When these environment variables are **not set**:
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`

The application automatically:
- ✅ Enables demo login at `/auth/demo-login`
- ✅ Uses demo email allowlist (`demo@example.com`, `test@example.com`, etc.)
- ✅ Shows warning messages in logs
- ⚠️ **Not suitable for production use**

### Production Mode
When Google OAuth credentials **are set**:
- ✅ Disables demo login (returns 404)
- ✅ Requires real Google OAuth authentication
- ✅ Uses your configured email allowlist
- ✅ Secure for production use

## Google OAuth Setup for Production

### 1. Create Google Cloud Project

```bash
# Create new project
gcloud projects create your-watered-project-id --name="Watered Production"

# Set as default project
gcloud config set project your-watered-project-id

# Enable required APIs
gcloud services enable iam.googleapis.com
```

### 2. Configure OAuth Consent Screen

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **APIs & Services → OAuth consent screen**
3. Choose **External** user type
4. Fill in application details:
   - **App name**: Watered Plant Tracker
   - **User support email**: your-email@domain.com
   - **Developer contact**: your-email@domain.com
5. Add scopes:
   - `https://www.googleapis.com/auth/userinfo.email`
   - `https://www.googleapis.com/auth/userinfo.profile`
6. Add test users (your email addresses)

### 3. Create OAuth 2.0 Credentials

1. Navigate to **APIs & Services → Credentials**
2. Click **Create Credentials → OAuth 2.0 Client IDs**
3. Choose **Web application**
4. Configure:
   - **Name**: Watered Web Client
   - **Authorized JavaScript origins**:
     - `https://yourdomain.com`
     - `http://localhost:8080` (for local testing)
   - **Authorized redirect URIs**:
     - `https://yourdomain.com/auth/callback`
     - `http://localhost:8080/auth/callback` (for local testing)

5. Save the **Client ID** and **Client Secret**

## Environment Variables for Production

### Required Production Variables

Create a production environment file or set these in your deployment:

```bash
# Google OAuth (REQUIRED to disable demo mode)
GOOGLE_CLIENT_ID=your-google-client-id-from-step-3
GOOGLE_CLIENT_SECRET=your-google-client-secret-from-step-3

# Session Security (REQUIRED for production)
SESSION_SECRET=your-secure-random-32-character-secret

# User Access Control
ALLOWED_EMAILS=you@yourdomain.com,partner@yourdomain.com,friend@gmail.com
ADMIN_EMAILS=you@yourdomain.com

# Production Settings
ENVIRONMENT=production
PORT=8080

# Database (optional, defaults to ./data/watered.db)
DATABASE_PATH=/app/data/watered.db
```

### Generate Secure Session Secret

```bash
# Generate a secure session secret
openssl rand -base64 32

# Or use the just command
just generate-session-secret
```

## Docker Production Configuration

### 1. Create Production Environment File

```bash
# Create .env.production
cat > .env.production << 'EOF'
GOOGLE_CLIENT_ID=your-actual-google-client-id
GOOGLE_CLIENT_SECRET=your-actual-google-client-secret
SESSION_SECRET=your-secure-session-secret
ALLOWED_EMAILS=you@yourdomain.com,partner@yourdomain.com
ADMIN_EMAILS=you@yourdomain.com
ENVIRONMENT=production
PORT=8080
DATABASE_PATH=/app/data/watered.db
EOF
```

### 2. Build Production Docker Image

```bash
# Build for production (AMD64 for cloud compatibility)
just docker-build-gcp

# Or manually
docker buildx build --platform linux/amd64 -t watered:production .
```

### 3. Run with Production Configuration

```bash
# Run with production environment
docker run -d \
  --name watered-prod \
  -p 8080:8080 \
  --env-file .env.production \
  -v watered-prod-data:/app/data \
  watered:production

# Verify demo mode is disabled
curl http://localhost:8080/auth/demo-login
# Should return: 404 Not Found
```

## Google Cloud Deployment

### 1. Build and Push to Artifact Registry

```bash
# Set environment variables
export GCP_PROJECT_ID="your-watered-project-id"
export GCP_REGION="us-central1"

# Build and push
just docker-deploy-gcp
```

### 2. Deploy to Cloud Run

```bash
# Deploy to Cloud Run with production environment
gcloud run deploy watered \
  --image $GCP_REGION-docker.pkg.dev/$GCP_PROJECT_ID/watered-repo/watered:latest \
  --platform managed \
  --region $GCP_REGION \
  --allow-unauthenticated \
  --port 8080 \
  --memory 512Mi \
  --cpu 1 \
  --set-env-vars ENVIRONMENT=production \
  --set-env-vars PORT=8080 \
  --set-env-vars DATABASE_PATH=/app/data/watered.db \
  --set-env-vars GOOGLE_CLIENT_ID="your-google-client-id" \
  --set-env-vars GOOGLE_CLIENT_SECRET="your-google-client-secret" \
  --set-env-vars SESSION_SECRET="your-session-secret" \
  --set-env-vars ALLOWED_EMAILS="you@domain.com,partner@domain.com" \
  --set-env-vars ADMIN_EMAILS="you@domain.com"
```

### 3. Update OAuth Redirect URLs

After deployment, update your Google OAuth configuration:

1. Get your Cloud Run URL: `gcloud run services describe watered --region=$GCP_REGION --format="value(status.url)"`
2. Add to **Authorized redirect URIs**: `https://your-cloud-run-url/auth/callback`
3. Add to **Authorized JavaScript origins**: `https://your-cloud-run-url`

### 4. Verify Production Deployment

```bash
# Get your Cloud Run URL
CLOUD_RUN_URL=$(gcloud run services describe watered --region=$GCP_REGION --format="value(status.url)")

# Test that demo mode is disabled
curl $CLOUD_RUN_URL/auth/demo-login
# Should return: 404 Not Found

# Test application health
curl $CLOUD_RUN_URL/health
# Should return: {"status":"ok","service":"watered"}

# Test authentication (should redirect to Google)
curl -I $CLOUD_RUN_URL/auth/login
# Should return: 307 Temporary Redirect to accounts.google.com
```

## Security Considerations

### 1. Environment Variables Security

**✅ DO:**
- Use Google Cloud Secret Manager for sensitive values
- Set environment variables through your deployment platform
- Use strong, randomly generated session secrets
- Regularly rotate OAuth credentials

**❌ DON'T:**
- Commit `.env.production` to git
- Use demo credentials in production
- Share session secrets
- Use weak or predictable secrets

### 2. OAuth Security

```bash
# Use production OAuth settings
GOOGLE_CLIENT_ID=your-real-client-id  # NOT "demo-client-id"
GOOGLE_CLIENT_SECRET=your-real-secret  # NOT "demo-client-secret"

# Restrict allowed emails
ALLOWED_EMAILS=specific@domain.com,trusted@domain.com  # NOT demo emails

# Use HTTPS redirect URLs
# ✅ https://yourdomain.com/auth/callback
# ❌ http://localhost:8080/auth/callback (only for local dev)
```

### 3. Session Security

```bash
# Strong session secret (32+ characters)
SESSION_SECRET=$(openssl rand -base64 32)

# Enable secure cookies for HTTPS
# This is automatically handled when ENVIRONMENT=production
```

## Troubleshooting

### Demo Mode Still Appearing

**Problem**: Demo login is still accessible in production

**Solution**: Verify OAuth credentials are set
```bash
# Check environment variables in your container
docker exec watered-container env | grep GOOGLE_CLIENT_ID
# Should show your real client ID, not "demo-client-id"

# Check application logs
docker logs watered-container
# Should NOT show "Warning: Google OAuth2 credentials not set"
```

### OAuth Redirect Mismatch

**Problem**: "redirect_uri_mismatch" error during login

**Solution**: Update OAuth configuration
1. Check your current URL in browser
2. Add exact URL to Google OAuth redirect URIs
3. Include both `http://localhost:8080/auth/callback` (dev) and `https://yourdomain.com/auth/callback` (prod)

### Authentication Failures

**Problem**: Users can't log in after OAuth setup

**Solution**: Check email allowlist
```bash
# Verify allowed emails are set correctly
ALLOWED_EMAILS=actual@emails.com,that@users.have  # NOT demo emails

# Check logs for "user not in allowlist" errors
docker logs watered-container | grep "allowlist"
```

### Session Issues

**Problem**: Users get logged out frequently

**Solution**: Check session secret consistency
```bash
# Ensure SESSION_SECRET is consistent across deployments
# Don't regenerate it unless necessary (will log out all users)

# Check cookie settings for HTTPS
# Secure cookies require HTTPS in production
```

## Complete Production Checklist

- [ ] Google Cloud project created
- [ ] OAuth consent screen configured
- [ ] OAuth 2.0 credentials created
- [ ] Production environment variables set
- [ ] Session secret generated and set
- [ ] Email allowlist configured
- [ ] Admin emails configured
- [ ] Docker image built for AMD64
- [ ] Image pushed to Artifact Registry
- [ ] Application deployed to Cloud Run
- [ ] OAuth redirect URLs updated
- [ ] Demo mode verified as disabled
- [ ] Authentication flow tested
- [ ] Admin panel access verified
- [ ] HTTPS enabled (recommended)
- [ ] Monitoring and logging set up

## Next Steps

After production setup:

1. **Set up monitoring**: Use Google Cloud Monitoring for uptime and performance
2. **Configure backups**: Set up regular data backups
3. **Enable HTTPS**: Use Cloud Load Balancer or custom domain with SSL
4. **Set up alerts**: Monitor for authentication failures and errors
5. **Documentation**: Document your specific OAuth and deployment configuration
6. **Testing**: Create a staging environment for testing updates