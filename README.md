 # envdoc
A fast .env file linter, schema generator, and comparison tool for Go.

[![Go Report Card](https://goreportcard.com/badge/github.com/adnaneAkk/envdoc)](https://goreportcard.com/report/github.com/adnaneAkk/envdoc)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Why?
Ever deployed to production and realized you had a typo in your `.env` file? Or wondered what changed between your dev and prod environments? envdoc catches those errors before they break your app.

## Installation
```bash
go install github.com/adnaneAkk/envdoc@latest
```

Or build from source:
```bash
git clone https://github.com/adnaneAkk/envdoc.git
cd envdoc
go build
```

## Quick Start
```bash
# Validate .env file
envdoc

# Strict mode (enforce UPPER_CASE)
envdoc -s

# Compare two env files
envdoc compare .env.production .env.development

# Generate JSON schema
envdoc schema -o schema.json

# Generate YAML schema
envdoc schema -f yaml -o schema.yaml
```

## Features
- ✅ Syntax validation (missing `=`, invalid keys)
- ✅ Duplicate key detection
- ✅ Strict mode (enforce uppercase naming)
- ✅ Quote handling with escape sequences
- ✅ Inline comment support
- ✅ Type inference (string, int, float, boolean)
- ✅ Schema generation (JSON/YAML/text)
- ✅ **Environment comparison** (diff production vs development)

## Example
**Input `.env`:**
```env
API_KEY=12345
DB_HOST=localhost
api_key=67890  # Duplicate!
DEBUG=true
MISSING_EQUALS  # Syntax error
```

**Output:**
```bash
$ envdoc
Errors: 1 found
  Line 5 [syntax]: missing '=' (Key: )

Warnings: 1 found
  Line 3 [duplicate]: Duplicate key detected; first occurrence on line 1 (Key: API_KEY)
```

## Usage

### Validate
```bash
envdoc                    # Validate .env
envdoc production.env     # Validate specific file
envdoc -s                 # Strict mode
```

### Compare Files
```bash
# Compare two env files
envdoc compare .env.production .env.development

# Using flags
envdoc compare --env1 .env.prod --env2 .env.dev
# OR
envdoc compare --env1=.env.prod --env2=.env.dev
```

**Example output:**
```bash
=== Comparison: .env.production vs .env.development ===
  - API_TIMEOUT           (only in .env.production)
  + DEBUG_MODE            (only in .env.development)
  ~ DB_HOST               "prod.db.com" → "localhost"

3 difference(s) found
```

### Generate Schema
```bash
envdoc schema                        # JSON to stdout
envdoc schema -o schema.json         # JSON to file
envdoc schema -f yaml                # YAML format
envdoc schema -f text                # Human-readable
```

## Roadmap
- [x] Environment file comparison
- [ ] Schema validation (validate .env against schema.json)
- [ ] Template generation (.env.example)
- [ ] Secret detection
- [ ] Multi-file support
- [ ] CI/CD integration examples
- [ ] Encrypting or hiding sensitive data (API keys, passwords, etc.)

## Contributing
This is my first public Go project! Found a bug or have an idea? Open an issue or PR.
 
