# Horizon - Security Analytics Platform

Horizon is an open-source alternative to expensive SIEM platforms like Splunk, ELK Stack, and OpenSearch. It provides comprehensive log aggregation, analysis, alerting, and incident management capabilities for SOC operations.

Built with Go for performance and developer productivity, using modern tools like NATS for streaming, PostgreSQL for structured data, and OCSF for log normalization.

## Architecture Overview

Horizon follows a microservices architecture with the following core components:

### Data Ingestion
- **Horizon Server**: High-performance Go service for log ingestion
- **NATS JetStream**: Streaming buffer for reliable log delivery
- **Multiple protocols**: Syslog, HTTP/HTTPS, TCP/UDP, gRPC
- **Secure transmission**: End-to-end encryption and authentication

### Storage Layer
- **PostgreSQL**: Configuration, user data, incidents, and alerts
- **Valkey**: Caching and temporary data storage
- **Quickwit**: Log storage and search (replacing Elasticsearch)
- **ClickHouse**: Analytics and aggregations

### Processing Engine
- **Normalization Service**: OCSF schema transformation
- **Rule Engine**: Sigma rule evaluation and correlation
- **Enrichment Services**: IoC lookup, geolocation, etc.

### Analytics & Visualization
- Advanced statistical analysis
- Custom dashboards and reporting
- Threat hunting capabilities

### Alerting & Incident Management
- Rule-based alerting
- Correlation engine
- Incident lifecycle management
- Notification systems

## Getting Started

### Prerequisites
- Docker
- Docker Compose
- Go 1.21+
- Bun (for client)

### Running Horizon

1. Clone the repository:
```bash
git clone <repository-url>
cd horizon
```

2. Start the services:
```bash
make run-docker
```

3. Initialize the database:
```bash
# Run migrations once services are up
make migrate-up
```

4. Access the UI:
- Horizon Client: [http://localhost:3000](http://localhost:3000)
- Quickwit Console: [http://localhost:7280](http://localhost:7280)
- NATS Monitoring: [http://localhost:8222](http://localhost:8222)

## Development

### Using Make Commands

Horizon includes a comprehensive Makefile for common development tasks:

```bash
make help                    # Show available commands
make setup                  # Setup development environment
make build                  # Build the server binary
make run                    # Run the server locally
make test                   # Run tests
make migrate-up             # Run database migrations
make gen-models             # Generate models from OpenAPI spec
```

### Project Structure
```
horizon/
├── Makefile                # Development commands
├── docker-compose.yml      # Container orchestration
├── README.md
├── client/                 # Frontend UI service (React + Bun)
│   ├── package.json
│   ├── vite.config.js
│   └── ...
├── server/                 # Backend services (Go)
│   ├── go.mod
│   ├── openapi.yaml        # API specification
│   ├── cmd/
│   │   └── api/
│   ├── configs/
│   ├── internal/
│   │   ├── handlers/
│   │   ├── models/
│   │   ├── routes/
│   │   ├── services/
│   │   └── utils/
│   └── Dockerfile
├── nats-config/            # NATS configuration
└── docs/                   # Documentation
    └── ...
```

## Roadmap

### Core Functionality
- [ ] Agent development for log collection
- [ ] OCSF schema normalization
- [ ] Sigma rule engine implementation
- [ ] Advanced analytics features
- [ ] Threat intelligence integration
- [ ] Compliance reporting
- [ ] Performance optimizations

### Advanced Security Features
- [ ] User and Entity Behavior Analytics (UEBA)
- [ ] Threat Intelligence Platform (TIP) with STIX/TAXII support
- [ ] Advanced correlation engine with temporal and cross-source correlation
- [ ] Asset discovery and management
- [ ] Forensic capabilities with timeline analysis
- [ ] Vulnerability correlation

### Enterprise Features
- [ ] Case management and incident response workflows
- [ ] Granular role-based access control (RBAC)
- [ ] Multi-tenancy support
- [ ] Compliance reporting (SOX, PCI-DSS, HIPAA, GDPR)
- [ ] Data retention and archiving policies
- [ ] Privileged access management

### Scalability & Performance
- [ ] Distributed architecture with clustering
- [ ] Horizontal scaling capabilities
- [ ] High availability and disaster recovery
- [ ] Load balancing across multiple nodes
- [ ] Performance optimizations for large deployments

### Analytics & Intelligence
- [ ] Machine learning-powered anomaly detection
- [ ] Predictive analytics
- [ ] Automated threat hunting capabilities
- [ ] Pattern recognition and clustering
- [ ] Risk scoring algorithms

### Integration & Ecosystem
- [ ] Extensive connector framework
- [ ] API-first design with comprehensive documentation
- [ ] Third-party app marketplace
- [ ] Webhook and automation support
- [ ] Integration with popular SOAR platforms

## License

This project is licensed under the MIT License - see the LICENSE file for details.