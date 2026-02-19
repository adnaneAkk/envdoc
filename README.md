# envdoc

A fast .env file linter, schema generator, and comparison tool for Go.

[![Go Report Card](https://goreportcard.com/badge/github.com/adnaneAkk/envdoc)](https://goreportcard.com/report/github.com/adnaneAkk/envdoc)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Why?

Ever deployed to production and realized you had a typo in your `.env` file? Or wondered what changed between your dev and prod environments? envdoc catches those errors before they break your app — and makes sure you never accidentally expose secrets in your docs or diffs.

## Installation

```bash
go install github.com/adnaneAkk/envdoc@latest

# if that doesnt work ,please specify by the latest version number
go install github.com/adnaneAkk/envdoc@0.1.1
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

# Generate JSON schema (sensitive values are redacted by default)
envdoc schema -o schema.json

# Generate schema with real values exposed
envdoc schema -o schema.json --unmask

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
- ✅ **Sensitive data redaction** (auto-detect and hide secrets in all outputs)

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

# Show real values instead of [SENSITIVE]
envdoc compare --env1 .env.prod --env2 .env.dev --unmask
```

**Example output:**

```bash
=== Comparison: .env.production vs .env.development ===
  - API_TIMEOUT           = [SENSITIVE] (only in .env.production)
  + DEBUG_MODE            = true (only in .env.development)
  ~ DB_HOST               "prod.db.com" → "localhost"
  ~ DB_PASSWORD           "[SENSITIVE]" → "[SENSITIVE]"
3 difference(s) found
```

### Generate Schema

```bash
envdoc schema                        # JSON to stdout
envdoc schema -o schema.json         # JSON to file
envdoc schema -f yaml                # YAML format
envdoc schema -f text                # Human-readable
envdoc schema --unmask               # Expose sensitive values (prompts for confirmation)
```

**Example schema output (JSON):**

```json
{
  "DB_PASSWORD": {
    "Value": "[SENSITIVE]",
    "Type": "string",
    "Required": false,
    "Sensitive": true
  },
  "PORT": {
    "Value": "8080",
    "Type": "integer",
    "Required": false,
    "Sensitive": false
  }
}
```

### Sensitive Data Detection

envdoc automatically detects and redacts secrets in all outputs. Detection uses two layers:

- **Key-based**: matches known patterns like `PASSWORD`, `SECRET`, `TOKEN`, `JWT`, `PRIVATE_KEY`, `CERT` and more
- **Value-based**: Shannon entropy analysis to catch high-entropy strings (API keys, hashes), plus regex patterns for known formats like Stripe keys, GitHub tokens, PEM blocks, and DSN connection strings

Use `--unmask` on any command to expose real values. When run interactively, you'll be prompted to confirm:

```bash
⚠  This will expose sensitive values in output. Continue? [y/N]:
```

## Roadmap

- [x] Environment file comparison
- [x] Sensitive data redaction
- [ ] Schema validation (validate .env against schema.json)
- [ ] Template generation (.env.example)
- [ ] Multi-file support
- [ ] CI/CD integration examples

## Contributing

This is my first public Go project! Found a bug or have an idea? Open an issue or PR.

