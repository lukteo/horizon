# Horizon - Security Analytics Platform - Progress Tracking

## Project Overview
Horizon is an open-source alternative to expensive SIEM platforms like Splunk, ELK Stack, and OpenSearch. It provides comprehensive log aggregation, analysis, alerting, and incident management capabilities for SOC operations.

## Current Status
**Date:** February 9, 2026  
**Phase:** Foundation & Architecture

## Completed Features

### ✅ Core Infrastructure
- [x] Go-based backend with Chi router
- [x] PostgreSQL for metadata storage
- [x] NATS for message queuing
- [x] MinIO/S3-compatible blob storage for raw logs
- [x] Docker Compose orchestration
- [x] Database migrations with Goose

### ✅ Data Processing Pipeline
- [x] Raw log ingestion via NATS
- [x] Blob storage integration (MinIO) for cost-effective log storage
- [x] PostgreSQL metadata management for log references
- [x] Asynchronous log processing pipeline
- [x] Error handling and retry mechanisms

### ✅ OCSF Normalization Engine
- [x] OCSF schema implementation
- [x] Dynamic log mapping configuration
- [x] Field transformation engine
- [x] Severity mapping and categorization
- [x] Nested field extraction (dot notation)

### ✅ Architecture & Best Practices
- [x] Domain-driven design structure
- [x] Interface-based architecture ("accept interfaces, return structs")
- [x] Proper separation of concerns
- [x] Testable components
- [x] Go module management
- [x] Configuration management with Viper

### ✅ Development Infrastructure
- [x] Makefile with common commands
- [x] Docker Compose setup
- [x] Database migration system
- [x] Testing framework
- [x] Structured logging with zerolog

## Planned Features

### 🔄 In Progress
- [ ] Sigma rule engine implementation
- [ ] Advanced correlation engine
- [ ] User interface with React and Bun
- [ ] Authentication and authorization system

### 📋 Roadmap
- [ ] Agent development for log collection
- [ ] Threat intelligence integration
- [ ] Advanced analytics features
- [ ] Compliance reporting (SOX, PCI-DSS, HIPAA, GDPR)
- [ ] Performance optimizations
- [ ] User and Entity Behavior Analytics (UEBA)
- [ ] Case management and incident response workflows
- [ ] Horizontal scaling capabilities
- [ ] API-first design with comprehensive documentation

## Architecture Highlights

### Data Flow
```
Logs → NATS (Buffer) → Blob Storage (Raw Logs) → PostgreSQL (Metadata) → Quickwit (Search) → ClickHouse (Analytics)
```

### Domain Structure
```
server/
├── cmd/
│   └── api/
├── internal/
│   ├── config/           # Configuration management
│   ├── logprocessor/     # Log processing domain
│   ├── ocsf/             # OCSF normalization domain
│   ├── storage/          # Storage operations (DB + Blob)
│   ├── messaging/        # NATS operations
│   ├── route/            # Routing logic
│   ├── middleware/       # Shared middleware
│   └── modelz/           # Shared data models
```

### Technology Stack
- **Backend**: Go 1.25 with Chi router
- **Database**: PostgreSQL for metadata
- **Messaging**: NATS JetStream
- **Storage**: MinIO for raw logs, Quickwit for search, ClickHouse for analytics
- **Frontend**: React + Bun (planned)
- **Infrastructure**: Docker Compose

## Next Steps
1. Implement Sigma rule engine
2. Develop React UI with RTK Query integration
3. Add authentication system
4. Implement advanced correlation features
5. Create log ingestion agents

---
*Last Updated: February 9, 2026*