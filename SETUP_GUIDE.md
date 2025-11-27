# Wound_IQ API - Complete Setup Guide

This guide walks you through setting up the Wound_IQ REST API on your local system and preparing it for GitHub.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local System Setup](#local-system-setup)
3. [Database Setup](#database-setup)
4. [Application Setup](#application-setup)
5. [Testing the API](#testing-the-api)
6. [GitHub Setup](#github-setup)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

1. **Go (version 1.22 or higher)**
   - Download: https://golang.org/dl/
   - Verify installation:
     ```bash
     go version
     # Should output: go version go1.22.x ...
     ```

2. **PostgreSQL (version 12 or higher)**
   - Download: https://www.postgresql.org/download/
   - Verify installation:
     ```bash
     psql --version
     # Should output: psql (PostgreSQL) 12.x or higher
     ```

3. **Git**
   - Download: https://git-scm.com/downloads
   - Verify installation:
     ```bash
     git --version
     ```

4. **Text Editor/IDE** (choose one)
   - VS Code (recommended): https://code.visualstudio.com/
   - GoLand: https://www.jetbrains.com/go/
   - Vim/Nano for terminal editing

---

## Local System Setup

### Step 1: Create Project Directory

```bash
# Create a directory for your project
mkdir -p ~/projects/wound_iq_api
cd ~/projects/wound_iq_api
```

### Step 2: Initialize Go Module

```bash
# Initialize Go module
go mod init wound_iq_api

# This creates go.mod file
```

### Step 3: Create Project Structure

```bash
# Create directory structure
mkdir -p cmd/api
mkdir -p internal/{config,db,models,handlers,router}
mkdir -p docs

# Create initial files
touch cmd/api/main.go
touch internal/config/config.go
touch internal/db/postgres.go
touch internal/models/{common.go,patient.go,clinician.go,assessment.go}
touch internal/handlers/{patient.go,clinician.go,assessment.go,report.go}
touch internal/router/router.go
touch .env.example
touch .gitignore
touch Makefile
touch README.md
```

---

## Database Setup

### Step 1: Start PostgreSQL

**On macOS:**
```bash
# If using Homebrew
brew services start postgresql@14

# Or start manually
pg_ctl -D /usr/local/var/postgres start
```

**On Linux:**
```bash
sudo systemctl start postgresql
# or
sudo service postgresql start
```

**On Windows:**
- Start PostgreSQL service from Services panel
- Or use pgAdmin

### Step 2: Create Database

```bash
# Connect to PostgreSQL as superuser
psql -U postgres

# In psql prompt, create database
CREATE DATABASE wound_iq;

# Create a user (optional but recommended)
CREATE USER wound_iq_user WITH PASSWORD 'your_secure_password';

# Grant privileges
GRANT ALL PRIVILEGES ON DATABASE wound_iq TO wound_iq_user;

# Exit psql
\q
```

### Step 3: Run Database Scripts

You already have these scripts. Run them in order:

```bash
# Navigate to where your SQL scripts are located
cd /path/to/your/sql/scripts

# 1. Create schema
psql -U postgres -d wound_iq -f wound_iq_schema_creation.sql

# 2. Create functions
psql -U postgres -d wound_iq -f wound_iq_functions_corrected.sql

# 3. Load sample data (optional)
psql -U postgres -d wound_iq -f wound_iq_sample_data_US_corrected.sql
```

**Verify Setup:**
```bash
# Connect to database
psql -U postgres -d wound_iq

# Check tables
\dt

# Check functions
\df

# Exit
\q
```

---

## Application Setup

### Step 1: Copy All Source Files

Copy the content from all the artifacts I provided into their respective files:

1. `cmd/api/main.go` - Main application entry point
2. `internal/config/config.go` - Configuration management
3. `internal/db/postgres.go` - Database connection
4. `internal/models/*.go` - All model files
5. `internal/handlers/*.go` - All handler files
6. `internal/router/router.go` - Router configuration
7. `go.mod` - Go module file
8. `.env.example` - Environment template
9. `.gitignore` - Git ignore rules
10. `Makefile` - Build automation
11. `README.md` - Documentation

### Step 2: Configure Environment

```bash
# Copy .env.example to .env
cp .env.example .env

# Edit .env file
nano .env
# or
code .env
```

**Update .env with your database credentials:**
```env
# If using default postgres user
DB_DSN=postgres://postgres:yourpassword@localhost:5432/wound_iq?sslmode=disable

# If using custom user
DB_DSN=postgres://wound_iq_user:your_secure_password@localhost:5432/wound_iq?sslmode=disable

# Server port
PORT=8080
```

### Step 3: Install Dependencies

```bash
# Install all Go dependencies
go mod download
go mod tidy

# This will download:
# - github.com/gin-gonic/gin
# - github.com/jackc/pgx/v5
# - github.com/joho/godotenv
# and their dependencies
```

### Step 4: Verify Setup

```bash
# Check if everything compiles
go build cmd/api/main.go

# If successful, you should see no errors
```

---

## Testing the API

### Step 1: Start the Server

```bash
# Start the API server
go run cmd/api/main.go

# or using Make
make run
```

**You should see:**
```
Successfully connected to PostgreSQL database
Starting server on port 8080
```

### Step 2: Test Health Check

Open a new terminal window:

```bash
# Test health endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","timestamp":"2024-01-15T10:30:00Z"}
```

### Step 3: Test API Endpoints

```bash
# Get all patients
curl http://localhost:8080/v1/patients

# Get specific patient
curl http://localhost:8080/v1/patients/1

# Create a patient
curl -X POST http://localhost:8080/v1/patients \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Test Patient",
    "date_of_birth": "1990-01-01T00:00:00Z",
    "gender": "Male",
    "medical_record_number": "TEST001"
  }'

# Get all assessments
curl http://localhost:8080/v1/assessments

# Get patient wound history
curl http://localhost:8080/v1/patients/1/history
```

### Step 4: Run Tests

```bash
# Run all tests
make test

# or
go test -v ./...
```

---

## GitHub Setup

### Step 1: Initialize Git Repository

```bash
# Initialize git in your project directory
cd ~/projects/wound_iq_api
git init

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: Wound_IQ REST API"
```

### Step 2: Create GitHub Repository

1. Go to https://github.com
2. Click the "+" icon → "New repository"
3. Repository name: `wound_iq_api`
4. Description: "Production-ready REST API for wound assessment management"
5. Choose "Public" or "Private"
6. **DO NOT** initialize with README (we already have one)
7. Click "Create repository"

### Step 3: Connect Local to GitHub

```bash
# Add remote origin (replace YOUR_USERNAME)
git remote add origin https://github.com/YOUR_USERNAME/wound_iq_api.git

# Verify remote
git remote -v

# Push to GitHub
git branch -M main
git push -u origin main
```

### Step 4: Verify on GitHub

1. Go to your repository on GitHub
2. Verify all files are present
3. Check that README.md displays properly

### Step 5: Set Up GitHub Repository Settings

**Add Repository Description:**
- Go to repository settings
- Add description: "Production-ready REST API for wound assessment management built with Go, Gin, and PostgreSQL"
- Add topics: `go`, `rest-api`, `postgresql`, `gin`, `healthcare`, `wound-care`

**Protect Main Branch (Optional):**
- Settings → Branches
- Add branch protection rule for `main`
- Enable "Require pull request reviews before merging"

---

## Project Files Checklist

Ensure you have all these files:

```
✓ cmd/api/main.go
✓ internal/config/config.go
✓ internal/db/postgres.go
✓ internal/models/common.go
✓ internal/models/patient.go
✓ internal/models/clinician.go
✓ internal/models/assessment.go
✓ internal/handlers/patient.go
✓ internal/handlers/clinician.go
✓ internal/handlers/assessment.go
✓ internal/handlers/report.go
✓ internal/router/router.go
✓ docs/api.yaml
✓ .env.example
✓ .gitignore
✓ go.mod
✓ Makefile
✓ README.md
✓ SETUP_GUIDE.md (this file)
```

---

## Troubleshooting

### Problem: "Failed to connect to database"

**Solutions:**
1. Check PostgreSQL is running:
   ```bash
   # macOS/Linux
   ps aux | grep postgres
   
   # Windows - check Services
   ```

2. Verify database exists:
   ```bash
   psql -U postgres -l | grep wound_iq
   ```

3. Check connection string in `.env`:
   - Verify username/password
   - Verify database name
   - Verify host and port

4. Test connection manually:
   ```bash
   psql -U postgres -d wound_iq
   ```

### Problem: "go: cannot find module"

**Solution:**
```bash
go mod download
go mod tidy
```

### Problem: "Port 8080 already in use"

**Solutions:**
1. Find process using port 8080:
   ```bash
   # macOS/Linux
   lsof -i :8080
   
   # Windows
   netstat -ano | findstr :8080
   ```

2. Kill the process or change PORT in `.env`

### Problem: "Function add_patient does not exist"

**Solution:**
```bash
# Re-run the functions script
psql -U postgres -d wound_iq -f wound_iq_functions_corrected.sql
```

### Problem: "Invalid date format"

**Solution:**
Dates must be in ISO-8601 format:
```
Correct: "2024-01-15T00:00:00Z"
Wrong: "2024-01-15" or "01/15/2024"
```

### Problem: Building on Windows

**Solution:**
If you encounter path issues:
```bash
# Use PowerShell or Git Bash
# Or set environment variables:
set CGO_ENABLED=0
set GOOS=windows
go build cmd/api/main.go
```

---

## Next Steps

After successful setup:

1. **Explore the API** - Try all endpoints using curl or Postman
2. **Read the code** - Understand the structure and patterns
3. **Add features** - Extend functionality as needed
4. **Write tests** - Add more comprehensive test coverage
5. **Deploy** - Consider Docker or cloud deployment

---

## Quick Reference Commands

```bash
# Start server
make run

# Run tests
make test

# Build binary
make build

# Run linter
make lint

# Clean build artifacts
make clean

# View all targets
make help
```

---

## Support

If you encounter issues:

1. Check this troubleshooting section
2. Review the README.md
3. Check PostgreSQL logs
4. Review application logs
5. Open an issue on GitHub

---

## Success Checklist

- [ ] PostgreSQL is installed and running
- [ ] Database `wound_iq` exists with schema and functions
- [ ] Go 1.22+ is installed
- [ ] All source files are created
- [ ] `.env` file is configured with correct credentials
- [ ] Dependencies are installed (`go mod download`)
- [ ] Application compiles without errors
- [ ] Server starts successfully
- [ ] Health endpoint returns 200 OK
- [ ] Can retrieve patients from database
- [ ] Code is committed to Git
- [ ] Repository is pushed to GitHub

Congratulations! Your Wound_IQ API is now set up and running! 🎉