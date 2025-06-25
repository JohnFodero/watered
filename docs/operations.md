# Operations Guide

This document provides comprehensive operational procedures for maintaining and monitoring the Watered plant tracking application.

## Table of Contents

- [Monitoring](#monitoring)
- [Logging](#logging)
- [Backup and Recovery](#backup-and-recovery)
- [Performance Management](#performance-management)
- [Security Operations](#security-operations)
- [Incident Response](#incident-response)
- [Maintenance Procedures](#maintenance-procedures)
- [Scaling](#scaling)

## Monitoring

### Health Check Monitoring

#### Basic Health Monitoring

```bash
# Quick health check
curl -f http://localhost:8080/health

# Detailed system health
curl http://localhost:8080/health/detailed | jq '.'

# Monitor health status in production
watch -n 30 'curl -s http://localhost:8080/health | jq ".status"'
```

#### Automated Health Monitoring Script

```bash
#!/bin/bash
# health-monitor.sh

ENDPOINT="http://localhost:8080/health/detailed"
ALERT_EMAIL="admin@yourdomain.com"
LOG_FILE="/var/log/watered/health-monitor.log"

check_health() {
    local response=$(curl -s "$ENDPOINT")
    local status=$(echo "$response" | jq -r '.status')
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    echo "[$timestamp] Health check: $status" >> "$LOG_FILE"
    
    if [ "$status" != "healthy" ]; then
        echo "[$timestamp] ALERT: Application unhealthy - $status" >> "$LOG_FILE"
        
        # Send alert email
        echo "Watered application health alert: $status" | \
            mail -s "Watered Health Alert" "$ALERT_EMAIL"
        
        # Log detailed information
        echo "$response" | jq '.' >> "$LOG_FILE"
    fi
}

# Run health check
check_health
```

### Application Metrics

#### Memory Monitoring

```bash
# Monitor memory usage
curl -s http://localhost:8080/health/detailed | jq '.system.memory'

# Memory usage alert script
memory_check() {
    local usage=$(curl -s http://localhost:8080/health/detailed | \
                  jq -r '.system.memory.memory_usage_percent')
    
    if (( $(echo "$usage > 80" | bc -l) )); then
        echo "High memory usage: ${usage}%"
        # Trigger alert
    fi
}
```

#### Performance Metrics

```bash
# Monitor response times
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/health

# curl-format.txt content:
#     time_namelookup:  %{time_namelookup}\n
#        time_connect:  %{time_connect}\n
#     time_appconnect:  %{time_appconnect}\n
#    time_pretransfer:  %{time_pretransfer}\n
#       time_redirect:  %{time_redirect}\n
#  time_starttransfer:  %{time_starttransfer}\n
#                     ----------\n
#          time_total:  %{time_total}\n
```

### Docker Container Monitoring

```bash
# Monitor container resource usage
docker stats watered

# Container health status
docker inspect watered | jq '.[0].State.Health'

# Monitor container logs
docker logs -f watered --tail 100
```

### System Resource Monitoring

```bash
# CPU and Memory usage
top -p $(pgrep watered)

# Disk usage
df -h
du -sh /opt/watered/data

# Network connections
netstat -tulpn | grep :8080
```

## Logging

### Log Configuration

#### Application Logging

```bash
# View application logs
tail -f /var/log/watered/application.log

# Docker logs
docker compose logs -f watered

# Systemd logs
journalctl -u watered -f
```

#### Log Rotation

Create `/etc/logrotate.d/watered`:

```
/var/log/watered/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 watered watered
    postrotate
        systemctl reload watered
    endscript
}
```

#### Centralized Logging (ELK Stack)

```yaml
# docker-compose.logging.yml
version: '3.8'
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.14.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

  kibana:
    image: docker.elastic.co/kibana/kibana:7.14.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200

  logstash:
    image: docker.elastic.co/logstash/logstash:7.14.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    depends_on:
      - elasticsearch

volumes:
  elasticsearch_data:
```

### Log Analysis

#### Common Log Queries

```bash
# Error analysis
grep "ERROR" /var/log/watered/*.log | tail -20

# Authentication failures
grep "auth.*failed" /var/log/watered/*.log

# High response times
grep "slow" /var/log/watered/*.log

# Database errors
grep "database.*error" /var/log/watered/*.log
```

#### Performance Analysis

```bash
# Request volume analysis
awk '{print $4}' /var/log/nginx/access.log | cut -d: -f2 | sort | uniq -c

# Response time analysis
awk '{print $NF}' /var/log/nginx/access.log | sort -n | tail -20

# Status code analysis
awk '{print $9}' /var/log/nginx/access.log | sort | uniq -c
```

## Backup and Recovery

### Data Backup

#### Database Backup

```bash
#!/bin/bash
# backup-script.sh

BACKUP_DIR="/opt/backups/watered"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="watered_backup_$DATE.tar.gz"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Docker volume backup
docker run --rm \
    -v watered_data:/data \
    -v "$BACKUP_DIR":/backup \
    alpine tar czf "/backup/$BACKUP_FILE" /data

# Verify backup
if [ $? -eq 0 ]; then
    echo "Backup successful: $BACKUP_FILE"
    
    # Remove old backups (keep 30 days)
    find "$BACKUP_DIR" -name "watered_backup_*.tar.gz" -mtime +30 -delete
else
    echo "Backup failed!"
    exit 1
fi
```

#### Configuration Backup

```bash
# Backup configuration files
tar czf config_backup_$(date +%Y%m%d).tar.gz \
    .env \
    docker-compose.yml \
    nginx.conf \
    /etc/systemd/system/watered.service
```

### Automated Backup with Cron

```bash
# Add to crontab
crontab -e

# Daily backup at 2 AM
0 2 * * * /opt/watered/scripts/backup-script.sh

# Weekly configuration backup
0 3 * * 0 /opt/watered/scripts/backup-config.sh
```

### Recovery Procedures

#### Data Recovery

```bash
# Stop application
docker compose down
# or
systemctl stop watered

# Restore from backup
docker run --rm \
    -v watered_data:/data \
    -v /opt/backups/watered:/backup \
    alpine tar xzf /backup/watered_backup_YYYYMMDD_HHMMSS.tar.gz -C /

# Start application
docker compose up -d
# or
systemctl start watered

# Verify recovery
curl http://localhost:8080/health/detailed
```

#### Configuration Recovery

```bash
# Restore configuration
tar xzf config_backup_YYYYMMDD.tar.gz

# Reload systemd if service file changed
systemctl daemon-reload
systemctl restart watered
```

## Performance Management

### Performance Monitoring

#### Response Time Monitoring

```bash
# Monitor API response times
ab -n 100 -c 10 http://localhost:8080/api/plant/status

# Load testing
docker run --rm -i grafana/k6 run - <<EOF
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 10 },
    { duration: '5m', target: 10 },
    { duration: '2m', target: 0 },
  ],
};

export default function() {
  let response = http.get('http://host.docker.internal:8080/health');
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 100ms': (r) => r.timings.duration < 100,
  });
}
EOF
```

#### Resource Usage Optimization

```bash
# Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap

# CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile

# Goroutine analysis
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

### Performance Tuning

#### Application Tuning

```bash
# Optimize Go runtime
export GOGC=100
export GOMAXPROCS=4

# Tune garbage collector
export GODEBUG=gctrace=1
```

#### System Tuning

```bash
# Increase file descriptor limits
echo "watered soft nofile 65536" >> /etc/security/limits.conf
echo "watered hard nofile 65536" >> /etc/security/limits.conf

# Optimize TCP settings
echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65535' >> /etc/sysctl.conf
sysctl -p
```

## Security Operations

### Security Monitoring

#### Access Log Analysis

```bash
# Monitor suspicious activity
tail -f /var/log/nginx/access.log | grep -E "(40[0-9]|50[0-9])"

# Failed authentication attempts
grep "401\|403" /var/log/nginx/access.log | tail -20

# Unusual request patterns
awk '{print $1}' /var/log/nginx/access.log | sort | uniq -c | sort -nr | head -10
```

#### Security Scanning

```bash
# Vulnerability scanning
docker run --rm -v $(pwd):/app aquasec/trivy fs /app

# Port scanning detection
netstat -tulpn | grep LISTEN

# SSL certificate monitoring
openssl x509 -in /etc/ssl/certs/watered.crt -text -noout | grep "Not After"
```

### Security Incident Response

#### Immediate Response

```bash
# Block suspicious IP
iptables -A INPUT -s SUSPICIOUS_IP -j DROP

# Emergency shutdown
systemctl stop watered
docker compose down

# Isolate container
docker network disconnect bridge watered
```

#### Forensic Analysis

```bash
# Preserve logs
cp -r /var/log/watered /forensics/logs_$(date +%Y%m%d_%H%M%S)

# Container analysis
docker exec watered ps aux
docker exec watered netstat -tulpn
docker diff watered
```

## Incident Response

### Incident Classification

- **P0 Critical**: Service completely down
- **P1 High**: Major functionality impaired
- **P2 Medium**: Minor functionality affected
- **P3 Low**: Cosmetic issues

### Response Procedures

#### Service Down (P0)

1. **Immediate Assessment**
   ```bash
   # Check service status
   systemctl status watered
   curl -f http://localhost:8080/health
   ```

2. **Quick Recovery**
   ```bash
   # Restart service
   systemctl restart watered
   # or
   docker compose restart watered
   ```

3. **Escalation**
   - If restart fails, check logs
   - Escalate to development team
   - Consider rollback to previous version

#### Performance Issues (P1)

1. **Resource Check**
   ```bash
   # Check resources
   docker stats watered
   curl http://localhost:8080/health/detailed
   ```

2. **Scaling**
   ```bash
   # Horizontal scaling
   docker compose up --scale watered=2
   ```

3. **Optimization**
   - Review slow queries
   - Check memory usage
   - Optimize database

### Communication Templates

#### Status Page Update

```
🔴 We're experiencing issues with plant status updates. 
Our team is investigating and working on a fix. 
ETA: 15 minutes. 
Last updated: 2024-01-15 14:30 UTC
```

#### Incident Resolution

```
✅ The issue with plant status updates has been resolved. 
Root cause: Database connection timeout
Fix: Increased connection pool size
All services are now operating normally.
```

## Maintenance Procedures

### Regular Maintenance

#### Daily Tasks

```bash
# Check service health
curl -f http://localhost:8080/health/detailed

# Review logs for errors
grep "ERROR\|WARN" /var/log/watered/*.log | tail -10

# Monitor disk space
df -h | grep -E "(8[0-9]|9[0-9])%"
```

#### Weekly Tasks

```bash
# Update system packages
apt update && apt upgrade

# Review performance metrics
# Analyze backup integrity
# Security updates review
```

#### Monthly Tasks

```bash
# Certificate renewal check
certbot certificates

# Performance review
# Capacity planning
# Security audit
```

### Update Procedures

#### Application Updates

```bash
# Backup before update
./scripts/backup-script.sh

# Update with zero downtime
docker compose pull
docker compose up -d --no-deps watered

# Verify update
curl http://localhost:8080/health/detailed
```

#### Security Updates

```bash
# System security updates
apt update && apt list --upgradable
apt upgrade

# Container security updates
docker pull watered:latest
docker compose up -d
```

## Scaling

### Horizontal Scaling

#### Load Balancer Configuration

```nginx
upstream watered_backend {
    server watered_1:8080;
    server watered_2:8080;
    server watered_3:8080;
}

server {
    location / {
        proxy_pass http://watered_backend;
    }
}
```

#### Docker Swarm Scaling

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml watered

# Scale service
docker service scale watered_watered=3
```

### Vertical Scaling

#### Resource Limits

```yaml
# docker-compose.yml
services:
  watered:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

### Auto-scaling

#### Kubernetes HPA

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: watered-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: watered
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

This operations guide provides comprehensive procedures for maintaining, monitoring, and scaling the Watered application in production environments.