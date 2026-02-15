# Tarh Workflow Orchestrator

A production-grade Go CLI that executes the full **SP → FP → Funding → Opportunity** lifecycle against the Tarh API. Every run generates fully randomized data so each execution is unique.

## Architecture

```
tarh-script/
├── main.go                 # Entry point & orchestration pipeline
├── config/
│   └── config.go           # Environment variable loading
├── client/
│   └── http_client.go      # Reusable HTTP client (JSON, multipart, retry)
├── models/
│   └── models.go           # Shared data structures
├── services/
│   ├── sp.go               # SP: register, login, OTP, create company
│   ├── fp.go               # FP: login, OTP, approve company
│   ├── funding.go          # Create funding request
│   └── opportunity.go      # Convert funding to opportunity
├── utils/
│   ├── random.go           # Random data generators
│   ├── date.go             # Date helpers
│   └── files.go            # Test asset generation
├── assets/                 # Auto-generated test images & PDFs
├── .env.example            # Environment variable template
└── README.md
```

## Prerequisites

- **Go 1.21+** installed
- Network access to the Tarh API

## Quick Start

```bash
# 1. Clone / navigate to the project
cd tarh-script

# 2. Copy the example env file
cp .env.example .env

# 3. (Optional) Edit .env to change the base URL or credentials
#    By default it targets https://api-dev2.tarh.com.sa/api

# 4. Install dependencies
go mod tidy

# 5. Run the orchestrator
go run .
```

## Configuration

All configuration is via environment variables (or `.env` file):

| Variable       | Default                                  | Description                  |
|----------------|------------------------------------------|------------------------------|
| `BASE_URL`     | `https://api-dev2.tarh.com.sa/api`       | API base URL (no trailing /) |
| `FP_EMAIL`     | `fp_user_0@sa.com`                       | Fund Provider login email    |
| `FP_PASSWORD`  | `secret@123`                             | Fund Provider login password |
| `SP_PASSWORD`  | `Secret@1234`                            | SP registration password     |
| `OTP`          | `123456`                                 | Fixed OTP code               |

## Workflow Steps

| # | Step                 | Endpoint                               | Auth    |
|---|----------------------|----------------------------------------|---------|
| 1 | SP Register          | `POST /auth/sp/register`               | None    |
| 2 | SP Login             | `POST /auth/login`                     | None    |
| 3 | SP OTP               | `POST /auth/otp`                       | None    |
| 4 | Create Company       | `POST /auth/sp/company`                | SP      |
| 5 | FP Login             | `POST /auth/Fplogin`                   | None    |
| 6 | FP OTP               | `POST /auth/otp`                       | None    |
| 7 | Approve Company      | `POST /company/companies/{id}/approve` | FP      |
| 8 | Create Funding       | `POST /v1/funding/sp/funding-request`  | SP      |
| 9 | Create Opportunity   | `POST /v1/funding/{id}/to-opportunity` | FP      |

## Features

- **Randomized data** on every run (names, emails, IDs, company details, etc.)
- **Retry with exponential back-off** for transient failures
- **Structured logging** via `slog`
- **Multipart/form-data** support for file uploads (images + PDFs)
- **Auto-generated test assets** (PNG image & PDF document)
- **Fail-fast** — stops immediately if any step fails
- **Context-aware** — global 5-minute timeout, cancellation propagation
- **No hardcoded IDs** — all identifiers extracted from API responses

## Building a Binary

```bash
go build -o tarh-orchestrator .
./tarh-orchestrator
```

## License

Internal tool — not for public distribution.
