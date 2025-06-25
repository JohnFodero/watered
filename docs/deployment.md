# Deployment Guide

This document provides comprehensive instructions for deploying the Watered plant tracking application in various environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Environment Configuration](#environment-configuration)
- [Development Deployment](#development-deployment)
- [Production Deployment](#production-deployment)
- [Docker Deployment](#docker-deployment)
- [Health Monitoring](#health-monitoring)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **Go**: Version 1.23 or later
- **Docker**: Version 20.10 or later (for containerized deployments)
- **Docker Compose**: Version 2.0 or later
- **Git**: For source code management
- **SSL Certificate**: For production HTTPS (recommended)

### Build Dependencies

```bash
# Install Go dependencies
go mod download
go mod verify

# Verify Go installation
go version
```

## Environment Configuration

### Environment Variables

Create a `.env` file in the project root with the following variables:

```bash
# Server Configuration
PORT=8080
HOST=localhost

# Google OAuth Configuration (Production)
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
OAUTH_REDIRECT_URL=http://localhost:8080/auth/callback

# Database Configuration
DATABASE_PATH=/data/watered.db

# Application Settings
APP_ENV=production
LOG_LEVEL=info
DEMO_MODE=false

# Security Settings
SESSION_SECRET=your_secure_session_secret_key
ALLOWED_ORIGINS=http://localhost:8080,https://yourdomain.com

# Monitoring
HEALTH_CHECK_INTERVAL=30s
MEMORY_LIMIT_MB=512
```

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add authorized redirect URIs:
   - Development: `http://localhost:8080/auth/callback`
   - Production: `https://yourdomain.com/auth/callback`

## Development Deployment

### Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd watered

# Install dependencies
go mod download

# Run in development mode
go run cmd/server/main.go
```

### Using Justfile (Recommended)

```bash
# Install just command runner
# On macOS: brew install just
# On Ubuntu: snap install --edge just

# View available commands
just --list

# Start development server
just dev

# Run tests
just test

# Build application
just build
```

### Development with Docker

```bash
# Start development environment
docker compose up --build

# Run in background
docker compose up -d

# View logs
docker compose logs -f watered

# Stop services
docker compose down
```

## Production Deployment

### Option 1: Direct Binary Deployment

```bash
# Build for production
go build -ldflags="-s -w" -o bin/watered cmd/server/main.go

# Create systemd service
sudo cp scripts/watered.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable watered
sudo systemctl start watered

# Check status
sudo systemctl status watered
```

### Option 2: Docker Production Deployment

```bash
# Build production image
docker build -t watered:latest .

# Run with production compose
docker compose -f docker-compose.prod.yml up -d

# Monitor deployment
docker compose -f docker-compose.prod.yml logs -f
```

### Production Configuration

#### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /path/to/certificate.crt;
    ssl_certificate_key /path/to/private.key;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://localhost:8080;
        access_log off;
    }
}
```

#### Systemd Service

Create `/etc/systemd/system/watered.service`:

```ini
[Unit]
Description=Watered Plant Tracking Application
After=network.target

[Service]
Type=simple
User=watered
Group=watered
WorkingDirectory=/opt/watered
ExecStart=/opt/watered/bin/watered
EnvironmentFile=/opt/watered/.env
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=watered

# Security measures
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/opt/watered/data

[Install]
WantedBy=multi-user.target
```

## Docker Deployment

### Development

```yaml
# docker-compose.yml
version: '3.8'
services:
  watered:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DEMO_MODE=true
    volumes:
      - watered_data:/home/watered/data
    restart: unless-stopped

volumes:
  watered_data:
```

### Production with Nginx

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  watered:
    build: .
    environment:
      - PORT=8080
      - APP_ENV=production
    volumes:
      - watered_data:/home/watered/data
      - ./logs:/home/watered/logs
    restart: always
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/certs
    depends_on:
      - watered
    restart: always

volumes:
  watered_data:
```

### Kubernetes Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: watered
spec:
  replicas: 2
  selector:
    matchLabels:
      app: watered
  template:
    metadata:
      labels:
        app: watered
    spec:
      containers:
      - name: watered
        image: watered:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: APP_ENV
          value: "production"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/detailed
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: watered-service
spec:
  selector:
    app: watered
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Health Monitoring

### Health Check Endpoints

1. **Basic Health Check**: `GET /health`
   - Simple status check
   - Used by load balancers
   - Returns: `{"status":"ok","service":"watered"}`

2. **Detailed Health Check**: `GET /health/detailed`
   - Comprehensive system monitoring
   - Database connectivity
   - Memory usage
   - Application components
   - System metrics

### Monitoring Integration

#### Prometheus Metrics (Future Enhancement)

```yaml
# Add to docker-compose.yml
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

#### Log Monitoring

```bash
# View application logs
docker compose logs -f watered

# Monitor with journalctl (systemd)
sudo journalctl -u watered -f

# Log rotation configuration
sudo logrotate -d /etc/logrotate.d/watered
```

## Security Considerations

### SSL/TLS Configuration

1. **Obtain SSL Certificate**:
   - Let's Encrypt (free): `certbot --nginx -d yourdomain.com`
   - Commercial certificate provider
   - Self-signed for development only

2. **Security Headers**: Ensure nginx includes:
   ```nginx
   add_header X-Frame-Options DENY;
   add_header X-Content-Type-Options nosniff;
   add_header X-XSS-Protection "1; mode=block";
   add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";
   ```

### Firewall Configuration

```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP
sudo ufw allow 443/tcp  # HTTPS
sudo ufw enable

# iptables example
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

### Environment Security

- Use strong session secrets
- Regularly rotate OAuth credentials
- Implement rate limiting
- Monitor access logs
- Use HTTPS in production
- Validate all user inputs

## Troubleshooting

### Common Issues

#### Application Won't Start

```bash
# Check logs
docker compose logs watered
journalctl -u watered -n 50

# Verify configuration
go run cmd/server/main.go --config-check

# Test database connection
curl http://localhost:8080/health/detailed
```

#### OAuth Authentication Issues

1. Verify Google OAuth credentials
2. Check redirect URLs match exactly
3. Ensure domains are whitelisted
4. Test in incognito/private mode

#### Performance Issues

```bash
# Monitor resource usage
docker stats watered

# Check memory usage
curl http://localhost:8080/health/detailed | jq '.system.memory'

# Monitor database performance
# Check plant state query times
```

#### SSL Certificate Issues

```bash
# Test SSL configuration
openssl s_client -connect yourdomain.com:443

# Check certificate expiry
openssl x509 -in certificate.crt -text -noout

# Renew Let's Encrypt
certbot renew --dry-run
```

### Performance Optimization

1. **Resource Limits**:
   - Set appropriate memory limits
   - Monitor CPU usage
   - Configure connection pooling

2. **Caching**:
   - Enable static file caching in nginx
   - Implement application-level caching
   - Use CDN for static assets

3. **Database Optimization**:
   - Regular maintenance
   - Monitor query performance
   - Implement connection pooling

### Backup and Recovery

```bash
# Backup data volume
docker run --rm -v watered_data:/data -v $(pwd):/backup alpine tar czf /backup/watered-backup.tar.gz /data

# Restore from backup
docker run --rm -v watered_data:/data -v $(pwd):/backup alpine tar xzf /backup/watered-backup.tar.gz -C /
```

### Support

For additional support:
- Check application logs: `docker compose logs watered`
- Review health status: `curl http://localhost:8080/health/detailed`
- Verify configuration: Review environment variables
- Monitor resource usage: `docker stats` or system monitoring tools

## Next Steps

1. Set up monitoring and alerting
2. Configure automated backups
3. Implement CI/CD pipeline
4. Set up log aggregation
5. Configure performance monitoring
6. Plan for scaling and high availability