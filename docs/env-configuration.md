# Environment Configuration Guide

The Watered application now supports loading environment variables from `.env` files automatically. This makes it easy to configure the application for different environments without modifying code.

## Table of Contents

- [How It Works](#how-it-works)
- [File Loading Order](#file-loading-order)
- [Setting Up Your Environment](#setting-up-your-environment)
- [Configuration Examples](#configuration-examples)
- [Environment-Specific Files](#environment-specific-files)
- [Best Practices](#best-practices)

## How It Works

When the application starts, it automatically loads environment variables from `.env` files in a specific order. This allows you to:

- ✅ Set default values in `.env.example`
- ✅ Override with your personal settings in `.env`
- ✅ Use local overrides in `.env.local`
- ✅ Have environment-specific configurations

The application will log which files were loaded and show your current configuration status.

## File Loading Order

Files are loaded in this order (later files override earlier ones):

1. **`.env.example`** - Template with defaults (lowest priority)
2. **`.env`** - Main environment file
3. **`.env.local`** - Local overrides (highest priority)
4. **`.env.{ENVIRONMENT}`** - Environment-specific file (if `ENVIRONMENT` is set)

### Example Loading Sequence

```bash
# If ENVIRONMENT=production is set, loads in this order:
1. .env.example     (defaults)
2. .env             (main config)
3. .env.local       (local overrides)
4. .env.production  (production-specific)
```

## Setting Up Your Environment

### 1. Quick Start for Development

```bash
# Copy the example file
cp .env.example .env

# Edit with your values
nano .env
```

### 2. Production Configuration

```bash
# Create production environment file
cat > .env << 'EOF'
# Production Configuration
ENVIRONMENT=production

# Google OAuth (required to disable demo mode)
GOOGLE_CLIENT_ID=your-actual-google-client-id
GOOGLE_CLIENT_SECRET=your-actual-google-client-secret

# Security
SESSION_SECRET=your-secure-32-character-session-secret

# User Management
ALLOWED_EMAILS=you@yourdomain.com,partner@yourdomain.com
ADMIN_EMAILS=you@yourdomain.com

# Server Configuration
PORT=8080
EOF
```

### 3. Local Development Overrides

```bash
# Create local overrides (git-ignored)
cat > .env.local << 'EOF'
# Local Development Overrides
PORT=3000
GOOGLE_CLIENT_ID=your-dev-client-id
GOOGLE_CLIENT_SECRET=your-dev-client-secret
EOF
```

## Configuration Examples

### Development Mode (Demo Enabled)

```bash
# .env
ENVIRONMENT=development
# Leave GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET empty for demo mode
ALLOWED_EMAILS=demo@example.com,test@example.com
ADMIN_EMAILS=admin@example.com
```

**Result**: Demo login available at `/auth/demo-login`

### Production Mode (Demo Disabled)

```bash
# .env
ENVIRONMENT=production
GOOGLE_CLIENT_ID=123456789-abc123def456.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-your-client-secret
SESSION_SECRET=your-secure-random-session-secret
ALLOWED_EMAILS=you@company.com,partner@company.com
ADMIN_EMAILS=you@company.com
```

**Result**: Demo login returns 404, requires Google OAuth

### Google Cloud Configuration

```bash
# .env
ENVIRONMENT=production
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
SESSION_SECRET=your-session-secret
ALLOWED_EMAILS=you@domain.com
ADMIN_EMAILS=you@domain.com

# GCP Configuration
GCP_PROJECT_ID=your-gcp-project
GCP_REGION=us-central1
```

## Environment-Specific Files

You can create environment-specific configuration files:

```bash
# .env.development
GOOGLE_CLIENT_ID=dev-client-id
GOOGLE_CLIENT_SECRET=dev-client-secret
ALLOWED_EMAILS=dev@example.com,test@example.com

# .env.staging
GOOGLE_CLIENT_ID=staging-client-id
GOOGLE_CLIENT_SECRET=staging-client-secret
ALLOWED_EMAILS=staging@company.com

# .env.production
GOOGLE_CLIENT_ID=prod-client-id
GOOGLE_CLIENT_SECRET=prod-client-secret
ALLOWED_EMAILS=prod@company.com,admin@company.com
```

Then set the environment:

```bash
# Load staging configuration
export ENVIRONMENT=staging
./watered

# Or set in your deployment
docker run -e ENVIRONMENT=production watered:latest
```

## Application Startup Logs

When you start the application, you'll see helpful logs:

```
2024/01/15 10:30:00 Loaded environment variables from: [.env .env.local]
2024/01/15 10:30:00 Configuration Status:
2024/01/15 10:30:00   Environment: production
2024/01/15 10:30:00   OAuth Mode: Production (Google OAuth enabled)
2024/01/15 10:30:00   Demo Login: Disabled
2024/01/15 10:30:00   Session Secret: Configured
2024/01/15 10:30:00   Allowed Emails: Configured
2024/01/15 10:30:00   Admin Emails: Configured
```

### Demo Mode Logs

```
2024/01/15 10:30:00 Loaded environment variables from: [.env]
2024/01/15 10:30:00 Configuration Status:
2024/01/15 10:30:00   Environment: development
2024/01/15 10:30:00   OAuth Mode: Demo (Google OAuth not configured)
2024/01/15 10:30:00   Demo Login: Available at /auth/demo-login
2024/01/15 10:30:00   Session Secret: Using development default
2024/01/15 10:30:00   Allowed Emails: Using demo defaults
2024/01/15 10:30:00   Admin Emails: Using demo defaults
```

## Best Practices

### 1. File Management

```bash
# ✅ DO commit to git
.env.example    # Template for other developers

# ✅ DO commit to git (if no secrets)
.env.development
.env.staging

# ❌ DON'T commit to git (contains secrets)
.env
.env.local
.env.production
```

### 2. Security

```bash
# ✅ Use strong session secrets
SESSION_SECRET=$(openssl rand -base64 32)

# ✅ Use real OAuth credentials in production
GOOGLE_CLIENT_ID=your-real-client-id  # Not demo values

# ✅ Restrict email access
ALLOWED_EMAILS=specific@domain.com,trusted@domain.com

# ❌ Don't use demo values in production
GOOGLE_CLIENT_ID=demo-client-id  # Wrong!
```

### 3. Development Workflow

```bash
# 1. Copy example file
cp .env.example .env

# 2. Set up for development (demo mode)
# Leave GOOGLE_CLIENT_ID empty in .env

# 3. Override locally if needed
echo "PORT=3000" > .env.local

# 4. Test production mode locally
cat > .env.local << 'EOF'
GOOGLE_CLIENT_ID=your-dev-oauth-id
GOOGLE_CLIENT_SECRET=your-dev-oauth-secret
EOF
```

### 4. Docker Integration

```bash
# Build with environment
docker build -t watered:latest .

# Run with .env file
docker run --env-file .env watered:latest

# Run with specific environment
docker run --env-file .env.production watered:latest

# Override individual variables
docker run \
  --env-file .env \
  -e ENVIRONMENT=production \
  -e PORT=8080 \
  watered:latest
```

## Troubleshooting

### Environment Variables Not Loading

**Problem**: Changes to `.env` file don't take effect

**Solution**: Restart the application
```bash
# Kill running server
just stop

# Start again (will reload .env files)
just run
```

### Configuration Status Check

**Problem**: Not sure what configuration is loaded

**Solution**: Check startup logs
```bash
# Look for these log lines:
# "Loaded environment variables from: [.env .env.local]"
# "Configuration Status:"
# "OAuth Mode: Production/Demo"
```

### Demo Mode Still Enabled

**Problem**: Demo login still works despite setting OAuth credentials

**Solution**: Verify environment variables are loaded
```bash
# Check your .env file
cat .env | grep GOOGLE_CLIENT_ID

# Check if file was loaded (look at startup logs)
# Restart application to reload files
```

### File Not Found Errors

**Problem**: Application can't find `.env` file

**Solution**: Files are optional - missing files are ignored
```bash
# Check current directory
pwd

# Verify .env file exists
ls -la .env*

# Check permissions
ls -la .env
```

## Advanced Usage

### Dynamic Environment Loading

```bash
# Load different environments
ENVIRONMENT=staging ./watered
ENVIRONMENT=production ./watered
```

### Multiple Configuration Sources

```bash
# Combine multiple sources
# 1. System environment variables (highest priority)
export GOOGLE_CLIENT_ID=system-override

# 2. .env files (loaded by application)
echo "GOOGLE_CLIENT_ID=file-value" > .env

# 3. Application defaults (lowest priority)
# Result: Uses "system-override" from system environment
```

### Just Commands with Environment

```bash
# Build with specific environment
ENVIRONMENT=production just docker-build-gcp

# Run with environment file
just run  # Automatically loads .env files
```

The new environment file support makes it much easier to configure the Watered application for different deployment scenarios while keeping sensitive credentials secure!