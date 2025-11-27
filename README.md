# Wound_IQ REST API

A production-ready REST API for managing wound assessments, built with Go, Gin framework, and PostgreSQL.

## 🚀 Features

- **Complete CRUD Operations** for patients, clinicians, and assessments
- **Advanced Filtering** for assessments by patient, clinician, and date range
- **Pagination Support** for all list endpoints
- **PostgreSQL Functions Integration** for optimized database operations
- **Structured Logging** with request/response tracking
- **Graceful Shutdown** for production reliability
- **Comprehensive Error Handling** with proper HTTP status codes
- **Input Validation** using Gin binding
- **CORS Support** for cross-origin requests
- **Health Check Endpoint** for monitoring

## 📋 Prerequisites

- **Go 1.22+** ([Download](https://golang.org/dl/))
- **PostgreSQL 12+** ([Download](https://www.postgresql.org/download/))
- **Git** ([Download](https://git-scm.com/downloads))
- **Make** (optional, for Makefile commands)

## 🛠️ Installation & Setup

### 1. Clone the Repository

```bash
# Clone the repository
git clone https://github.com/yourusername/wound_iq_api.git
cd wound_iq_api
```

### 2. Setup PostgreSQL Database

Ensure PostgreSQL is running, then create the database:

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE wound_iq;

# Exit psql
\q
```

Run the SQL scripts in order:

```bash
# Navigate to your SQL scripts directory
cd /path/to/sql/scripts

# Create schema
psql -U postgres -d wound_iq -f wound_iq_schema_creation.sql

# Create functions
psql -U postgres -d wound_iq -f wound_iq_functions_corrected.sql

# Load sample data (optional)
psql -U postgres -d wound_iq -f wound_iq_sample_data_US_corrected.sql
```

### 3. Configure Environment Variables

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your database credentials
# Example:
# DB_DSN=postgres://postgres:yourpassword@localhost:5432/wound_iq?sslmode=disable
# PORT=8080
```

### 4. Install Dependencies

```bash
# Install Go dependencies
go mod download
# or
make install-deps
```

### 5. Run the Application

```bash
# Run directly
go run cmd/api/main.go

# or use Make
make run

# or build and run binary
make build
./bin/wound_iq_api
```

The API will start on `http://localhost:8080`

## 📚 API Documentation

### Base URL
```
http://localhost:8080/v1
```

### Health Check
```bash
GET /health
```

### Patients

#### Get All Patients
```bash
GET /v1/patients?page=1&page_size=10

curl http://localhost:8080/v1/patients
```

#### Get Patient by ID
```bash
GET /v1/patients/:id

curl http://localhost:8080/v1/patients/1
```

#### Create Patient
```bash
POST /v1/patients
Content-Type: application/json

{
  "full_name": "John Doe",
  "date_of_birth": "1985-06-15T00:00:00Z",
  "gender": "Male",
  "medical_record_number": "MRN123456"
}

# Example curl
curl -X POST http://localhost:8080/v1/patients \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "John Doe",
    "date_of_birth": "1985-06-15T00:00:00Z",
    "gender": "Male",
    "medical_record_number": "MRN123456"
  }'
```

#### Update Patient
```bash
PUT /v1/patients/:id
Content-Type: application/json

{
  "full_name": "Jane Doe",
  "gender": "Female"
}

# Example curl
curl -X PUT http://localhost:8080/v1/patients/1 \
  -H "Content-Type: application/json" \
  -d '{"full_name": "Jane Doe"}'
```

#### Delete Patient
```bash
DELETE /v1/patients/:id

curl -X DELETE http://localhost:8080/v1/patients/1
```

#### Get Patient Wound History
```bash
GET /v1/patients/:id/history

curl http://localhost:8080/v1/patients/1/history
```

### Clinicians

#### Get All Clinicians
```bash
GET /v1/clinicians?page=1&page_size=10

curl http://localhost:8080/v1/clinicians
```

#### Get Clinician by ID
```bash
GET /v1/clinicians/:id

curl http://localhost:8080/v1/clinicians/1
```

#### Create Clinician
```bash
POST /v1/clinicians
Content-Type: application/json

{
  "full_name": "Dr. Sarah Johnson",
  "role": "Physician",
  "department": "Wound Care",
  "contact_info": "555-1234",
  "license_number": "LIC12345"
}

# Example curl
curl -X POST http://localhost:8080/v1/clinicians \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Dr. Sarah Johnson",
    "role": "Physician",
    "department": "Wound Care",
    "contact_info": "555-1234",
    "license_number": "LIC12345"
  }'
```

#### Update Clinician
```bash
PUT /v1/clinicians/:id
Content-Type: application/json

{
  "contact_info": "555-5678"
}

# Example curl
curl -X PUT http://localhost:8080/v1/clinicians/1 \
  -H "Content-Type: application/json" \
  -d '{"contact_info": "555-5678"}'
```

#### Delete Clinician
```bash
DELETE /v1/clinicians/:id

curl -X DELETE http://localhost:8080/v1/clinicians/1
```

### Assessments

#### Get All Assessments (with filters)
```bash
GET /v1/assessments?page=1&page_size=10&patient_id=1&clinician_id=2&start_date=2024-01-01T00:00:00Z&end_date=2024-12-31T23:59:59Z

curl "http://localhost:8080/v1/assessments?patient_id=1&page=1&page_size=10"
```

#### Get Assessment by ID
```bash
GET /v1/assessments/:id

curl http://localhost:8080/v1/assessments/1
```

#### Create Simple Assessment
```bash
POST /v1/assessments
Content-Type: application/json

{
  "clinician_id": 1,
  "patient_id": 1,
  "location": "Left Heel",
  "etiology": "Pressure Injury",
  "depth_of_injury": "Partial Thickness",
  "stage": "Stage II",
  "chronicity": "Acute",
  "healing_status": "Improving",
  "return_to_clinic": true
}

# Example curl
curl -X POST http://localhost:8080/v1/assessments \
  -H "Content-Type: application/json" \
  -d '{
    "clinician_id": 1,
    "patient_id": 1,
    "location": "Left Heel",
    "etiology": "Pressure Injury",
    "depth_of_injury": "Partial Thickness",
    "stage": "Stage II",
    "chronicity": "Acute",
    "healing_status": "Improving",
    "return_to_clinic": true
  }'
```

#### Create Full Assessment
```bash
POST /v1/assessments/full
Content-Type: application/json

{
  "clinician_id": 1,
  "patient_id": 1,
  "location": "Right Foot",
  "etiology": "Diabetic Ulcer",
  "depth_of_injury": "Full Thickness",
  "stage": "Stage III",
  "chronicity": "Chronic",
  "healing_status": "Slow Healing",
  "return_to_clinic": true,
  "infection_pain": {
    "localized_symptoms": "Redness",
    "systemic_symptoms": "None",
    "pain_present": "Yes",
    "pain_score": "5",
    "culture_results": "Negative",
    "antibiotic": "None"
  },
  "tissue_status": {
    "granulation_percent": 50,
    "epithelial_percent": 20,
    "slough_percent": 20,
    "eschar_percent": 5,
    "necrotic_percent": 5,
    "debridement": "Sharp"
  },
  "vitals": {
    "blood_pressure": "130/85",
    "temperature": 37.2,
    "pulse": 75,
    "respiration_rate": 16,
    "oxygen_saturation": 97
  },
  "wound_condition": {
    "length": 3.5,
    "width": 2.8,
    "depth": 0.8,
    "tunneling": false,
    "undermining": true,
    "edges": "Attached",
    "skin_condition": "Dry",
    "edema": "Mild",
    "blister": "No"
  },
  "exudate": {
    "exudate_type": "Serous",
    "exudate_amount": "Moderate",
    "odor": "None"
  },
  "treatment": {
    "primary_dressing": "Foam",
    "secondary_dressing": "Gauze",
    "tertiary_dressing": "Wrap",
    "frequency": "Daily",
    "supplies": "Saline, dressing pack",
    "orders": "Monitor glucose"
  }
}

# See docs/postman_collection.json for full example
```

#### Update Assessment
```bash
PUT /v1/assessments/:id
Content-Type: application/json

{
  "healing_status": "Healed",
  "return_to_clinic": false
}

# Example curl
curl -X PUT http://localhost:8080/v1/assessments/1 \
  -H "Content-Type: application/json" \
  -d '{"healing_status": "Healed"}'
```

#### Delete Assessment
```bash
DELETE /v1/assessments/:id

curl -X DELETE http://localhost:8080/v1/assessments/1
```

#### Get Full Assessment Details
```bash
GET /v1/assessments/:id/full

curl http://localhost:8080/v1/assessments/1/full
```

## 🧪 Testing

```bash
# Run all tests
make test

# or
go test -v ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🔍 Linting

```bash
# Install golangci-lint (if not already installed)
# See: https://golangci-lint.run/usage/install/

# Run linter
make lint

# or
golangci-lint run
```

## 🏗️ Building

```bash
# Build binary
make build

# Run the binary
./bin/wound_iq_api
```

## 📁 Project Structure

```
wound_iq_api/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── db/
│   │   └── postgres.go          # Database connection
│   ├── models/
│   │   ├── common.go            # Common types
│   │   ├── patient.go           # Patient models
│   │   ├── clinician.go         # Clinician models
│   │   └── assessment.go        # Assessment models
│   ├── handlers/
│   │   ├── patient.go           # Patient handlers
│   │   ├── clinician.go         # Clinician handlers
│   │   ├── assessment.go        # Assessment handlers
│   │   └── report.go            # Report handlers
│   └── router/
│       └── router.go            # Route definitions
├── docs/
│   ├── api.yaml                 # OpenAPI specification
│   └── postman_collection.json # Postman collection
├── .env.example                 # Environment variables template
├── .gitignore                   # Git ignore rules
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── Makefile                     # Build automation
└── README.md                    # This file
```

## 🐳 Docker Support (Optional)

Create a `Dockerfile`:

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o wound_iq_api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/wound_iq_api .
EXPOSE 8080
CMD ["./wound_iq_api"]
```

Build and run:
```bash
docker build -t wound_iq_api .
docker run -p 8080:8080 --env-file .env wound_iq_api
```

## 📝 Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_DSN` | PostgreSQL connection string | - | Yes |
| `PORT` | Server port | 8080 | No |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style Guidelines

- Follow standard Go formatting (`gofmt`)
- Write meaningful commit messages
- Add tests for new features
- Update documentation as needed
- Run linter before committing

## 🔒 Security

- Never commit `.env` files with credentials
- Use environment variables for sensitive data
- Implement proper authentication/authorization (future enhancement)
- Keep dependencies updated
- Follow security best practices

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Authors

- Your Name - Initial work

## 🙏 Acknowledgments

- Gin framework for excellent HTTP routing
- PostgreSQL for reliable data storage
- The Go community for amazing tools and libraries

## 📞 Support

For issues and questions:
- Open an issue on GitHub
- Contact: your.email@example.com

## 🗺️ Roadmap

- [ ] Add authentication (JWT)
- [ ] Add role-based authorization
- [ ] Implement caching (Redis)
- [ ] Add rate limiting
- [ ] Create admin dashboard
- [ ] Add file upload for wound images
- [ ] Implement WebSocket for real-time updates
- [ ] Add comprehensive audit logging
- [ ] Create Docker Compose setup
- [ ] Add Kubernetes manifests