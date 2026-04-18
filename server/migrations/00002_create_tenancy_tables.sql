-- +goose Up
-- +goose StatementBegin

CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(100) UNIQUE NOT NULL,
    plan       VARCHAR(50)  NOT NULL DEFAULT 'free',
    settings   JSONB        NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_organizations_slug ON organizations(slug);

-- +goose StatementEnd

-- +goose StatementBegin

CREATE TABLE organization_members (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(50) NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, user_id)
);

CREATE INDEX idx_org_members_org_id  ON organization_members(org_id);
CREATE INDEX idx_org_members_user_id ON organization_members(user_id);

-- +goose StatementEnd

-- +goose StatementBegin

CREATE TABLE api_keys (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID         NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    key_hash     VARCHAR(255) UNIQUE NOT NULL,
    scopes       TEXT[]       NOT NULL DEFAULT '{}',
    last_used_at TIMESTAMP WITH TIME ZONE,
    revoked_at   TIMESTAMP WITH TIME ZONE,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_org_id   ON api_keys(org_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
-- +goose StatementEnd
