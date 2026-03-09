-- Seed data for industry configurations
-- Run: psql $DB_URI -f apps/pkgs/db/seeds/industries.sql

INSERT INTO industry_configs (slug, name, description, knowledge_base, compliance_rules)
VALUES
(
    'healthcare',
    'Healthcare',
    'Healthcare and medical industry prompts',
    '{
        "best_practices": [
            "Always include disclaimers about not replacing professional medical advice",
            "Use precise medical terminology with lay-person explanations",
            "Include patient safety considerations",
            "Avoid definitive diagnoses in AI responses",
            "Reference clinical guidelines when applicable"
        ],
        "common_variables": ["patient_context", "medical_specialty", "symptom_description"],
        "templates": ["clinical_summary", "patient_education", "triage_assessment"]
    }'::jsonb,
    '{
        "rules": [
            {"id": "hipaa-phi", "name": "HIPAA PHI Protection", "description": "Must not include or request Protected Health Information directly", "severity": "error"},
            {"id": "medical-disclaimer", "name": "Medical Disclaimer", "description": "Must include appropriate medical disclaimers", "severity": "warning"},
            {"id": "evidence-based", "name": "Evidence-Based References", "description": "Should reference evidence-based guidelines", "severity": "info"}
        ]
    }'::jsonb
),
(
    'customer-support',
    'Customer Support',
    'Customer service and support prompts',
    '{
        "best_practices": [
            "Maintain empathetic and professional tone",
            "Follow escalation procedures for complex issues",
            "Include clear next steps in responses",
            "Personalize responses using customer context",
            "Track resolution metrics"
        ],
        "common_variables": ["customer_name", "issue_type", "product_name", "account_id"],
        "templates": ["initial_response", "escalation", "resolution_summary", "follow_up"]
    }'::jsonb,
    '{
        "rules": [
            {"id": "pii-protection", "name": "PII Protection", "description": "Must not expose customer PII in responses", "severity": "error"},
            {"id": "brand-tone", "name": "Brand Tone Consistency", "description": "Must maintain brand voice guidelines", "severity": "warning"},
            {"id": "response-time", "name": "Response Completeness", "description": "Must address all customer concerns", "severity": "info"}
        ]
    }'::jsonb
),
(
    'finance',
    'Financial Services',
    'Banking, insurance, and financial advisory prompts',
    '{
        "best_practices": [
            "Include appropriate financial disclaimers",
            "Avoid providing specific investment advice",
            "Reference regulatory requirements",
            "Use precise financial terminology",
            "Include risk disclosures"
        ],
        "common_variables": ["account_type", "risk_profile", "investment_horizon", "regulatory_jurisdiction"],
        "templates": ["market_analysis", "risk_assessment", "compliance_review", "client_advisory"]
    }'::jsonb,
    '{
        "rules": [
            {"id": "financial-disclaimer", "name": "Financial Disclaimer", "description": "Must include investment risk disclaimers", "severity": "error"},
            {"id": "regulatory-compliance", "name": "Regulatory Compliance", "description": "Must comply with applicable financial regulations", "severity": "error"},
            {"id": "suitability", "name": "Suitability Check", "description": "Should consider client suitability", "severity": "warning"}
        ]
    }'::jsonb
),
(
    'legal',
    'Legal',
    'Legal industry and law firm prompts',
    '{
        "best_practices": [
            "Include legal disclaimers about not constituting legal advice",
            "Reference applicable jurisdiction",
            "Use precise legal terminology",
            "Cite relevant statutes and case law when applicable",
            "Maintain attorney-client privilege considerations"
        ],
        "common_variables": ["jurisdiction", "practice_area", "case_type", "client_context"],
        "templates": ["legal_research", "contract_review", "case_summary", "client_memo"]
    }'::jsonb,
    '{
        "rules": [
            {"id": "legal-disclaimer", "name": "Legal Disclaimer", "description": "Must not constitute legal advice without proper context", "severity": "error"},
            {"id": "jurisdiction-specific", "name": "Jurisdiction Specificity", "description": "Must specify applicable jurisdiction", "severity": "warning"},
            {"id": "privilege-awareness", "name": "Privilege Awareness", "description": "Should be mindful of attorney-client privilege", "severity": "info"}
        ]
    }'::jsonb
),
(
    'education',
    'Education',
    'Educational and e-learning prompts',
    '{
        "best_practices": [
            "Adapt content to target age/level",
            "Use scaffolded learning approaches",
            "Include assessment criteria",
            "Promote critical thinking over rote answers",
            "Support multiple learning styles"
        ],
        "common_variables": ["grade_level", "subject", "learning_objective", "student_context"],
        "templates": ["lesson_plan", "assessment", "study_guide", "feedback"]
    }'::jsonb,
    '{
        "rules": [
            {"id": "age-appropriate", "name": "Age-Appropriate Content", "description": "Content must be appropriate for target age group", "severity": "error"},
            {"id": "educational-accuracy", "name": "Educational Accuracy", "description": "Must be factually accurate and up-to-date", "severity": "error"},
            {"id": "inclusive-language", "name": "Inclusive Language", "description": "Should use inclusive and accessible language", "severity": "warning"}
        ]
    }'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    knowledge_base = EXCLUDED.knowledge_base,
    compliance_rules = EXCLUDED.compliance_rules,
    updated_at = NOW();
