# Wound_IQ API - Deployment Guide

This guide covers deploying the Wound_IQ REST API to various environments.

## Table of Contents

1. [Docker Deployment](#docker-deployment)
2. [Docker Compose](#docker-compose)
3. [Cloud Deployment](#cloud-deployment)
4. [Production Checklist](#production-checklist)

---

## Docker Deployment

### Create Dockerfile

Create a `Dockerfile` in the project root:

```dockerfile
# Multi-stage build for smaller image size
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wound_iq_api cmd/api/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

# Copy binary from builder
COPY --from=builder /app/wound_iq_api .

# Change ownership
RUN chown -R appuser:appuser /home/appuser

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./wound_iq_api"]
```

### Build and Run Docker Image

```bash
# Build the image
docker build -t wound_iq_api:latest .

# Run the container
docker run -d \
  --name wound_iq_api \
  -p 8080:8080 \
  -e DB_DSN="postgres://user:pass@host.docker.internal:5432/wound_iq?sslmode=disable" \
  -e PORT=8080 \
  wound_iq_api:latest

# View logs
docker logs -f wound_iq_api

# Stop container
docker stop wound_iq_api

# Remove container
docker rm wound_iq_api
```

---

## Docker Compose

### Create docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    container_name: wound_iq_db
    environment:
      POSTGRES_DB: wound_iq
      POSTGRES_USER: wound_iq_user
      POSTGRES_PASSWORD: secure_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U wound_iq_user -d wound_iq"]
      interval: 10s
      timeout: 5s
      retries: 5

  api:
    build: .
    container_name: wound_iq_api
    ports:
      - "8080:8080"
    environment:
      DB_DSN: postgres://wound_iq_user:secure_password@postgres:5432/wound_iq?sslmode=disable
      PORT: 8080
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

volumes:
  postgres_data:
```

### Create scripts directory

```bash
mkdir -p scripts

# Copy your SQL scripts to scripts directory
cp wound_iq_schema_creation.sql scripts/01_schema.sql
cp wound_iq_functions_corrected.sql scripts/02_functions.sql
cp wound_iq_sample_data_US_corrected.sql scripts/03_sample_data.sql
```

### Run with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f api

# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: deletes data)
docker-compose down -v

# Restart services
docker-compose restart

# Scale API (multiple instances)
docker-compose up -d --scale api=3
```

---

## Cloud Deployment

### AWS (Elastic Beanstalk)

1. **Install AWS CLI and EB CLI:**
```bash
pip install awscli awsebcli
aws configure
```

2. **Initialize EB:**
```bash
eb init -p docker wound-iq-api
```

3. **Create environment:**
```bash
eb create wound-iq-prod
```

4. **Set environment variables:**
```bash
eb setenv DB_DSN="your_rds_connection_string" PORT=8080
```

5. **Deploy:**
```bash
eb deploy
```

### Google Cloud Platform (Cloud Run)

1. **Build and push to Container Registry:**
```bash
gcloud builds submit --tag gcr.io/PROJECT_ID/wound-iq-api

# Or use Cloud Build
gcloud builds submit --config cloudbuild.yaml
```

2. **Deploy to Cloud Run:**
```bash
gcloud run deploy wound-iq-api \
  --image gcr.io/PROJECT_ID/wound-iq-api \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars DB_DSN="your_connection_string"
```

### Heroku

1. **Create Heroku app:**
```bash
heroku create wound-iq-api
```

2. **Add PostgreSQL:**
```bash
heroku addons:create heroku-postgresql:hobby-dev
```

3. **Set environment variables:**
```bash
heroku config:set PORT=8080
```

4. **Deploy:**
```bash
git push heroku main
```

### DigitalOcean App Platform

1. **Create app.yaml:**
```yaml
name: wound-iq-api
services:
- name: api
  github:
    repo: yourusername/wound_iq_api
    branch: main
  build_command: go build -o bin/wound_iq_api cmd/api/main.go
  run_command: ./bin/wound_iq_api
  http_port: 8080
  envs:
  - key: DB_DSN
    value: ${db.DATABASE_URL}
  - key: PORT
    value: "8080"
databases:
- name: db
  engine: PG
  version: "14"
```

2. **Deploy:**
```bash
doctl apps create --spec app.yaml
```

---

## Production Checklist

### Security

- [ ] Use HTTPS/TLS certificates
- [ ] Implement authentication (JWT/OAuth)
- [ ] Add rate limiting
- [ ] Use environment variables for secrets
- [ ] Implement CORS properly
- [ ] Add input sanitization
- [ ] Use prepared statements (already done)
- [ ] Regular security audits
- [ ] Keep dependencies updated

### Performance

- [ ] Enable connection pooling (already configured)
- [ ] Add Redis caching
- [ ] Implement database indexes
- [ ] Use CDN for static assets
- [ ] Enable gzip compression
- [ ] Monitor query performance
- [ ] Set up load balancer
- [ ] Implement horizontal scaling

### Monitoring

- [ ] Set up logging (structured logging)
- [ ] Add metrics (Prometheus/Grafana)
- [ ] Configure alerts
- [ ] Set up error tracking (Sentry)
- [ ] Monitor API response times
- [ ] Database performance monitoring
- [ ] Health check endpoints (already done)
- [ ] Uptime monitoring

### Database

- [ ] Regular backups
- [ ] Database replication
- [ ] Connection pooling
- [ ] Query optimization
- [ ] Index optimization
- [ ] Migration strategy
- [ ] Backup testing

### CI/CD

- [ ] Automated testing
- [ ] Code coverage reports
- [ ] Linting in pipeline
- [ ] Automated deployments
- [ ] Rollback strategy
- [ ] Blue-green deployments
- [ ] Feature flags

### Documentation

- [ ] API documentation (OpenAPI)
- [ ] Deployment guide
- [ ] Architecture diagrams
- [ ] Runbooks for common issues
- [ ] Change log
- [ ] Version control

---

## Environment Variables for Production

```bash
# Production environment variables
export DB_DSN="postgres://user:password@prod-db:5432/wound_iq?sslmode=require"
export PORT=8080
export GIN_MODE=release
export LOG_LEVEL=info
export MAX_DB_CONNECTIONS=25
export ENABLE_CORS=true
export ALLOWED_ORIGINS="https://yourdomain.com"
```

---

## Nginx Reverse Proxy

Create `/etc/nginx/sites-available/wound_iq_api`:

```nginx
upstream wound_iq_api {
    server localhost:8080;
    # Add more servers for load balancing
    # server localhost:8081;
    # server localhost:8082;
}

server {
    listen 80;
    server_name api.yourdomain.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

    # Logging
    access_log /var/log/nginx/wound_iq_access.log;
    error_log /var/log/nginx/wound_iq_error.log;

    location / {
        proxy_pass http://wound_iq_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://wound_iq_api/health;
        access_log off;
    }
}
```

Enable and restart:
```bash
sudo ln -s /etc/nginx/sites-available/wound_iq_api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

## Systemd Service (Linux)

Create `/etc/systemd/system/wound_iq_api.service`:

```ini
[Unit]
Description=Wound_IQ REST API
After=network.target postgresql.service

[Service]
Type=simple
User=apiuser
Group=apiuser
WorkingDirectory=/opt/wound_iq_api
Environment="DB_DSN=postgres://user:pass@localhost:5432/wound_iq?sslmode=disable"
Environment="PORT=8080"
Environment="GIN_MODE=release"
ExecStart=/opt/wound_iq_api/wound_iq_api
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable wound_iq_api
sudo systemctl start wound_iq_api
sudo systemctl status wound_iq_api
```

---

## Kubernetes Deployment

Create `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wound-iq-api
  labels:
    app: wound-iq-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: wound-iq-api
  template:
    metadata:
      labels:
        app: wound-iq-api
    spec:
      containers:
      - name: api
        image: your-registry/wound-iq-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_DSN
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: dsn
        - name: PORT
          value: "8080"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: wound-iq-api-service
spec:
  selector:
    app: wound-iq-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

Deploy:
```bash
kubectl apply -f k8s/deployment.yaml
kubectl get pods
kubectl get services
```

---

## Monitoring with Prometheus

Add to `main.go`:
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// In router setup
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

---

This deployment guide provides multiple options for deploying your API. Choose the method that best fits your infrastructure and requirements.