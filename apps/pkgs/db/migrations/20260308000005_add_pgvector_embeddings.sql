-- Add embedding column to prompt_versions (stored as real[] for portability)
ALTER TABLE prompt_versions ADD COLUMN embedding real[];

-- Cosine similarity function for semantic search
CREATE OR REPLACE FUNCTION cosine_similarity(a real[], b real[]) RETURNS double precision AS $$
DECLARE
    dot double precision := 0;
    norm_a double precision := 0;
    norm_b double precision := 0;
    i integer;
BEGIN
    IF a IS NULL OR b IS NULL OR array_length(a, 1) != array_length(b, 1) THEN
        RETURN 0;
    END IF;
    FOR i IN 1..array_length(a, 1) LOOP
        dot := dot + a[i]::double precision * b[i]::double precision;
        norm_a := norm_a + a[i]::double precision * a[i]::double precision;
        norm_b := norm_b + b[i]::double precision * b[i]::double precision;
    END LOOP;
    IF norm_a = 0 OR norm_b = 0 THEN
        RETURN 0;
    END IF;
    RETURN dot / (sqrt(norm_a) * sqrt(norm_b));
END;
$$ LANGUAGE plpgsql IMMUTABLE STRICT;

-- Index on prompt_id + embedding IS NOT NULL for filtering
CREATE INDEX IF NOT EXISTS idx_prompt_versions_has_embedding
    ON prompt_versions (prompt_id) WHERE embedding IS NOT NULL;
