# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**acme-console** is a self-hosted ACME certificate management console for real-world operations. It serves as an orchestration and management layer on top of existing ACME clients (primarily acme.sh), providing visibility and workflow management for multi-domain, manual DNS, and wildcard certificate scenarios.

**Status**: Pre-MVP, planning phase. Implementation has not yet begun.

## Proposed Tech Stack

- **Backend**: Go
- **ACME Client**: acme.sh (external CLI integration)
- **API**: REST
- **Storage**: MySQL
- **UI**: Web Console (TBD)

## Architecture

```
Web Console → acme-console API → acme.sh CLI
                    ↓
              Storage (MySQL)
                    ↓
              ACME CA (Let's Encrypt/ZeroSSL)
```

**Certificate Lifecycle**: `Issue → TXT → Verify → Renew → Package → Deploy`

## Core Concepts

- **Certificate**: Entity supporting RSA/ECC, root + wildcard domains, multiple SANs
- **Challenge**: ACME DNS-01 challenge records (manual or API), with TXT summary generation
- **Target**: Deployment destinations (CDN, LoadBalancer, Ingress, File Export)

## Design Philosophy

- **Reality-first**: Acknowledges DNS may not be under direct control
- **Human-friendly**: TXT records designed to be shared with operators/clients
- **Ops-oriented**: Failures are expected; everything must be traceable
- **Composable**: CLI, API, and Web interfaces should all work

## When Implementation Begins

Suggested Go project structure:
```
cmd/           # Application entrypoints
internal/      # Private application code
pkg/           # Public library code
api/           # API definitions (OpenAPI/protobuf)
web/           # Frontend assets
```

Standard Go commands will apply:
```bash
go build ./...
go test ./...
go fmt ./...
```
