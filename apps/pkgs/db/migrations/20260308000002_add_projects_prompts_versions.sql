-- projects table
CREATE TABLE IF NOT EXISTS public.projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES public.organizations(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    slug            VARCHAR(50)  NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_projects_org_slug UNIQUE (organization_id, slug),
    CONSTRAINT chk_projects_slug CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$')
);

CREATE INDEX IF NOT EXISTS idx_projects_org_id ON public.projects (organization_id);

-- prompts table
CREATE TABLE IF NOT EXISTS public.prompts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID         NOT NULL REFERENCES public.projects(id) ON DELETE CASCADE,
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

CREATE INDEX IF NOT EXISTS idx_prompts_project_id ON public.prompts (project_id);

-- prompt_versions table
CREATE TABLE IF NOT EXISTS public.prompt_versions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prompt_id           UUID         NOT NULL REFERENCES public.prompts(id) ON DELETE CASCADE,
    version_number      INTEGER      NOT NULL,
    status              VARCHAR(20)  NOT NULL DEFAULT 'draft',
    content             JSONB        NOT NULL,
    variables           JSONB,
    change_description  VARCHAR(500),
    semantic_diff       JSONB,
    lint_result         JSONB,
    author_id           UUID         NOT NULL REFERENCES public.users(id),
    published_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_prompt_versions_prompt_number UNIQUE (prompt_id, version_number),
    CONSTRAINT chk_prompt_versions_number CHECK (version_number >= 1),
    CONSTRAINT chk_prompt_versions_status CHECK (status IN ('draft', 'review', 'production', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_prompt_versions_prompt_id ON public.prompt_versions (prompt_id);
CREATE INDEX IF NOT EXISTS idx_prompt_versions_status ON public.prompt_versions (status);
CREATE INDEX IF NOT EXISTS idx_prompt_versions_author_id ON public.prompt_versions (author_id);
CREATE INDEX IF NOT EXISTS idx_prompt_versions_created_at ON public.prompt_versions (created_at DESC);
