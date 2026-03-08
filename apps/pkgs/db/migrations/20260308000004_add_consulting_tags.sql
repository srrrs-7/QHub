-- Phase 6: Consulting Chat + Phase 7: Tags

-- Industry configs for consulting knowledge base
CREATE TABLE public.industry_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    knowledge_base JSONB NOT NULL DEFAULT '{}',
    compliance_rules JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Consulting chat sessions
CREATE TABLE public.consulting_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES public.organizations(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    industry_config_id UUID REFERENCES public.industry_configs(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_consulting_sessions_org ON public.consulting_sessions(organization_id);

-- Consulting chat messages
CREATE TABLE public.consulting_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES public.consulting_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    citations JSONB,
    actions_taken JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_consulting_messages_session ON public.consulting_messages(session_id);
CREATE INDEX idx_consulting_messages_created ON public.consulting_messages(created_at);

-- Platform benchmarks (monthly aggregates)
CREATE TABLE public.platform_benchmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    industry_config_id UUID NOT NULL REFERENCES public.industry_configs(id) ON DELETE CASCADE,
    period VARCHAR(7) NOT NULL, -- YYYY-MM
    avg_quality_score NUMERIC(3, 2),
    avg_latency_ms INTEGER,
    avg_cost_per_request NUMERIC(10, 6),
    total_executions BIGINT NOT NULL DEFAULT 0,
    p50_quality NUMERIC(3, 2),
    p90_quality NUMERIC(3, 2),
    opt_in_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (industry_config_id, period)
);

-- Tags
CREATE TABLE public.tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES public.organizations(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6B7280',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, name)
);

CREATE TABLE public.prompt_tags (
    prompt_id UUID NOT NULL REFERENCES public.prompts(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES public.tags(id) ON DELETE CASCADE,
    PRIMARY KEY (prompt_id, tag_id)
);

CREATE INDEX idx_prompt_tags_tag ON public.prompt_tags(tag_id);
