-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS log_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    source_type VARCHAR(100) NOT NULL,
    mapping_config JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_log_mappings_source_type ON log_mappings(source_type);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_log_mappings_created_at ON log_mappings(created_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS raw_log_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source VARCHAR(255) NOT NULL,
    source_type VARCHAR(100),
    blob_storage_path VARCHAR(500) NOT NULL,
    size_bytes BIGINT NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMP WITH TIME ZONE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_raw_log_references_source_type ON raw_log_references(source_type);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_raw_log_references_processed ON raw_log_references(processed);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_raw_log_references_received_at ON raw_log_references(received_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS normalized_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ocsf_event JSONB NOT NULL,
    source_mapping_id UUID REFERENCES log_mappings(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    indexed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_normalized_logs_source_mapping_id ON normalized_logs(source_mapping_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_normalized_logs_created_at ON normalized_logs(created_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sigma_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sigma_yaml TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_sigma_rules_enabled ON sigma_rules(enabled);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID REFERENCES sigma_rules(id),
    title VARCHAR(500) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'NEW',
    description TEXT,
    context JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_alerts_rule_id ON alerts(rule_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_alerts_status ON alerts(status);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_alerts_created_at ON alerts(created_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'OPEN',
    priority VARCHAR(50) DEFAULT 'MEDIUM',
    assigned_to UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_incidents_status ON incidents(status);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_incidents_priority ON incidents(priority);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS sigma_rules;
DROP TABLE IF EXISTS normalized_logs;
DROP TABLE IF EXISTS raw_logs;
DROP TABLE IF EXISTS log_mappings;