# Database Schema

## Overview

PostgreSQL 18。Atlas でスキーマ管理、sqlc でクエリ生成。

---

## Layer 1: Prompt Management Platform

### organizations

```sql
CREATE TABLE organizations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100)  NOT NULL,
    slug        VARCHAR(50)   NOT NULL UNIQUE,
    plan        VARCHAR(20)   NOT NULL DEFAULT 'free',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_organizations_plan CHECK (plan IN ('free', 'pro', 'team', 'enterprise')),
    CONSTRAINT chk_organizations_slug CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$')
);
```

### users

```sql
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255)  NOT NULL UNIQUE,
    name        VARCHAR(100)  NOT NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
```

### organization_members

```sql
CREATE TABLE organization_members (
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (organization_id, user_id),
    CONSTRAINT chk_org_members_role CHECK (role IN ('owner', 'admin', 'member', 'viewer'))
);

CREATE INDEX idx_org_members_user_id ON organization_members (user_id);
```

### api_keys

```sql
CREATE TABLE api_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    key_hash        VARCHAR(255) NOT NULL,
    key_prefix      VARCHAR(20)  NOT NULL,
    last_used_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_api_keys_prefix CHECK (key_prefix ~ '^pl_(live|test)_')
);

CREATE INDEX idx_api_keys_org_id ON api_keys (organization_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys (key_hash);
```

### projects

```sql
CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    slug            VARCHAR(50)  NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_projects_org_slug UNIQUE (organization_id, slug),
    CONSTRAINT chk_projects_slug CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$')
);

CREATE INDEX idx_projects_org_id ON projects (organization_id);
```

### prompts

```sql
CREATE TABLE prompts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name                VARCHAR(200) NOT NULL,
    slug                VARCHAR(80)  NOT NULL,
    prompt_type         VARCHAR(20)  NOT NULL,
    description         TEXT,
    latest_version      INTEGER      NOT NULL DEFAULT 0,
    production_version  INTEGER,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_prompts_project_slug UNIQUE (project_id, slug),
    CONSTRAINT chk_prompts_type CHECK (prompt_type IN ('system', 'user', 'combined')),
    CONSTRAINT chk_prompts_slug CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$')
);

CREATE INDEX idx_prompts_project_id ON prompts (project_id);
```

### prompt_versions

```sql
CREATE TABLE prompt_versions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prompt_id           UUID         NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    version_number      INTEGER      NOT NULL,
    status              VARCHAR(20)  NOT NULL DEFAULT 'draft',
    content             JSONB        NOT NULL,
    variables           JSONB,
    change_description  VARCHAR(500),
    semantic_diff       JSONB,
    lint_result         JSONB,
    author_id           UUID         NOT NULL REFERENCES users(id),
    published_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_prompt_versions_prompt_number UNIQUE (prompt_id, version_number),
    CONSTRAINT chk_prompt_versions_number CHECK (version_number >= 1),
    CONSTRAINT chk_prompt_versions_status CHECK (status IN ('draft', 'review', 'production', 'archived'))
);

CREATE INDEX idx_prompt_versions_prompt_id ON prompt_versions (prompt_id);
CREATE INDEX idx_prompt_versions_status ON prompt_versions (status);
CREATE INDEX idx_prompt_versions_author_id ON prompt_versions (author_id);
CREATE INDEX idx_prompt_versions_created_at ON prompt_versions (created_at DESC);
```

### execution_logs

```sql
CREATE TABLE execution_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prompt_version_id   UUID         NOT NULL REFERENCES prompt_versions(id),
    organization_id     UUID         NOT NULL REFERENCES organizations(id),
    request_body        JSONB        NOT NULL,
    response_body       JSONB        NOT NULL,
    provider_name       VARCHAR(50)  NOT NULL,
    model_name          VARCHAR(100) NOT NULL,
    token_usage         JSONB,
    latency_ms          INTEGER,
    metadata            JSONB,
    executed_at         TIMESTAMPTZ  NOT NULL,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exec_logs_prompt_version ON execution_logs (prompt_version_id);
CREATE INDEX idx_exec_logs_org_id ON execution_logs (organization_id);
CREATE INDEX idx_exec_logs_executed_at ON execution_logs (executed_at DESC);
CREATE INDEX idx_exec_logs_model ON execution_logs (model_name);
CREATE INDEX idx_exec_logs_org_executed ON execution_logs (organization_id, executed_at DESC);
```

### evaluations

```sql
CREATE TABLE evaluations (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_log_id  UUID          NOT NULL REFERENCES execution_logs(id) ON DELETE CASCADE,
    evaluator_id      UUID          REFERENCES users(id),
    evaluation_type   VARCHAR(20)   NOT NULL,
    overall_score     DECIMAL(5,2)  NOT NULL,
    criteria_scores   JSONB,
    comment           TEXT,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_evaluations_type CHECK (evaluation_type IN ('manual', 'automated')),
    CONSTRAINT chk_evaluations_score CHECK (overall_score >= 0.0 AND overall_score <= 100.0)
);

CREATE INDEX idx_evaluations_log_id ON evaluations (execution_log_id);
CREATE INDEX idx_evaluations_evaluator_id ON evaluations (evaluator_id);
CREATE INDEX idx_evaluations_score ON evaluations (overall_score);
```

### tags / prompt_tags

```sql
CREATE TABLE tags (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(50) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_tags_org_name UNIQUE (organization_id, name)
);

CREATE TABLE prompt_tags (
    prompt_id UUID NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    tag_id    UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (prompt_id, tag_id)
);

CREATE INDEX idx_prompt_tags_tag_id ON prompt_tags (tag_id);
```

---

## Layer 2: Consulting Chat

### industry_configs

```sql
CREATE TABLE industry_configs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    industry        VARCHAR(30) NOT NULL,
    enabled         BOOLEAN     NOT NULL DEFAULT TRUE,
    custom_knowledge JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_industry_configs_org_industry UNIQUE (organization_id, industry),
    CONSTRAINT chk_industry CHECK (industry IN (
        'healthcare', 'legal', 'finance',
        'customer_support', 'education', 'ecommerce', 'general'
    ))
);

CREATE INDEX idx_industry_configs_org_id ON industry_configs (organization_id);
```

### consulting_sessions

```sql
CREATE TABLE consulting_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         UUID         NOT NULL REFERENCES users(id),
    title           VARCHAR(200),
    industry        VARCHAR(30),
    prompt_id       UUID         REFERENCES prompts(id),
    status          VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_session_status CHECK (status IN ('active', 'closed')),
    CONSTRAINT chk_session_industry CHECK (
        industry IS NULL OR industry IN (
            'healthcare', 'legal', 'finance',
            'customer_support', 'education', 'ecommerce', 'general'
        )
    )
);

CREATE INDEX idx_sessions_org_id ON consulting_sessions (organization_id);
CREATE INDEX idx_sessions_user_id ON consulting_sessions (user_id);
CREATE INDEX idx_sessions_prompt_id ON consulting_sessions (prompt_id);
CREATE INDEX idx_sessions_created_at ON consulting_sessions (created_at DESC);
```

### consulting_messages

```sql
CREATE TABLE consulting_messages (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id    UUID        NOT NULL REFERENCES consulting_sessions(id) ON DELETE CASCADE,
    role          VARCHAR(20) NOT NULL,
    content       TEXT        NOT NULL,
    citations     JSONB,
    actions_taken JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_message_role CHECK (role IN ('user', 'assistant'))
);

CREATE INDEX idx_messages_session_id ON consulting_messages (session_id);
CREATE INDEX idx_messages_created_at ON consulting_messages (created_at);
```

### platform_benchmarks

```sql
CREATE TABLE platform_benchmarks (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    industry      VARCHAR(30)   NOT NULL,
    metric        VARCHAR(30)   NOT NULL,
    period        VARCHAR(7)    NOT NULL,
    p25           DECIMAL(10,2) NOT NULL,
    p50           DECIMAL(10,2) NOT NULL,
    p75           DECIMAL(10,2) NOT NULL,
    p90           DECIMAL(10,2) NOT NULL,
    sample_size   INTEGER       NOT NULL DEFAULT 0,
    calculated_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_benchmarks UNIQUE (industry, metric, period),
    CONSTRAINT chk_benchmark_metric CHECK (metric IN (
        'avg_score', 'avg_latency_ms', 'avg_tokens', 'avg_cost_per_1k'
    )),
    CONSTRAINT chk_benchmark_period CHECK (period ~ '^\d{4}-\d{2}$')
);

CREATE INDEX idx_benchmarks_industry_period ON platform_benchmarks (industry, period DESC);
```

---

## Migration Strategy

既存の `tasks` テーブルはサンプルとして残す。

### Phase 1: Foundation
```bash
make atlas-diff NAME=add_organizations_users_apikeys
```
- organizations, users, organization_members, api_keys

### Phase 2: Core
```bash
make atlas-diff NAME=add_projects_prompts_versions
```
- projects, prompts, prompt_versions

### Phase 3: Ingestion
```bash
make atlas-diff NAME=add_execution_logs_evaluations
```
- execution_logs, evaluations

### Phase 4: Tags
```bash
make atlas-diff NAME=add_tags
```
- tags, prompt_tags

### Phase 5: Consulting
```bash
make atlas-diff NAME=add_consulting_chat
```
- industry_configs, consulting_sessions, consulting_messages, platform_benchmarks

---

## Performance Notes

- **execution_logs**: 最高書き込み頻度。将来的に `executed_at` 月次パーティショニング検討
- **consulting_messages**: セッション内で時系列クエリが主。session_id + created_at のインデックスで十分
- **platform_benchmarks**: 月次バッチ集計。読み込み頻度は低い
- **JSONB**: `content`, `request_body`, `response_body` は柔軟性重視で JSONB 維持。GIN インデックスは必要に応じて追加
