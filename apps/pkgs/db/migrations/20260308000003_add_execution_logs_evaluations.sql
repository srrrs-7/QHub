-- Phase 3: Execution Logs & Evaluations

CREATE TABLE public.execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES public.organizations(id) ON DELETE CASCADE,
    prompt_id UUID NOT NULL REFERENCES public.prompts(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    -- Request
    request_body JSONB NOT NULL,
    -- Response
    response_body JSONB,
    -- Metrics
    model VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL DEFAULT 'openai',
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    estimated_cost NUMERIC(10, 6) NOT NULL DEFAULT 0,
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'success' CHECK (status IN ('success', 'error', 'timeout')),
    error_message TEXT,
    -- Metadata
    environment VARCHAR(20) NOT NULL DEFAULT 'production' CHECK (environment IN ('development', 'staging', 'production')),
    metadata JSONB,
    -- Timestamps
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_execution_logs_org ON public.execution_logs(organization_id);
CREATE INDEX idx_execution_logs_prompt ON public.execution_logs(prompt_id);
CREATE INDEX idx_execution_logs_prompt_version ON public.execution_logs(prompt_id, version_number);
CREATE INDEX idx_execution_logs_executed_at ON public.execution_logs(executed_at DESC);
CREATE INDEX idx_execution_logs_status ON public.execution_logs(status);

CREATE TABLE public.evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_log_id UUID NOT NULL REFERENCES public.execution_logs(id) ON DELETE CASCADE,
    -- Scores
    overall_score NUMERIC(3, 2) CHECK (overall_score >= 0 AND overall_score <= 5),
    accuracy_score NUMERIC(3, 2) CHECK (accuracy_score >= 0 AND accuracy_score <= 5),
    relevance_score NUMERIC(3, 2) CHECK (relevance_score >= 0 AND relevance_score <= 5),
    fluency_score NUMERIC(3, 2) CHECK (fluency_score >= 0 AND fluency_score <= 5),
    safety_score NUMERIC(3, 2) CHECK (safety_score >= 0 AND safety_score <= 5),
    -- Feedback
    feedback TEXT,
    evaluator_type VARCHAR(20) NOT NULL DEFAULT 'human' CHECK (evaluator_type IN ('human', 'auto', 'llm')),
    evaluator_id VARCHAR(255),
    -- Metadata
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evaluations_log ON public.evaluations(execution_log_id);
CREATE INDEX idx_evaluations_overall ON public.evaluations(overall_score);
