CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE agencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    jurisdiction VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(50),
    ssn_hash VARCHAR(255),
    address JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE agency_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    agency_role VARCHAR(50) DEFAULT 'member',
    is_primary BOOLEAN DEFAULT false,
    UNIQUE(agency_id, user_id)
);

CREATE TABLE programs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT
);

CREATE TABLE agency_programs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    program_id UUID NOT NULL REFERENCES programs(id),
    is_enabled BOOLEAN DEFAULT true,
    UNIQUE(agency_id, program_id)
);

CREATE TABLE worker_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    agency_id UUID NOT NULL REFERENCES agencies(id),
    specializations TEXT[] DEFAULT '{}',
    max_active_cases INT DEFAULT 50,
    current_case_count INT DEFAULT 0
);

CREATE TABLE document_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    case_number VARCHAR(50) UNIQUE NOT NULL,
    citizen_id UUID NOT NULL REFERENCES users(id),
    program_id UUID NOT NULL REFERENCES programs(id),
    status VARCHAR(50) NOT NULL DEFAULT 'submitted',
    priority VARCHAR(20) DEFAULT 'normal',
    zip_code VARCHAR(20),
    census_tract VARCHAR(50),
    submitted_at TIMESTAMPTZ DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    case_id UUID UNIQUE NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    household_size INT NOT NULL DEFAULT 1,
    annual_income DECIMAL(12,2) DEFAULT 0,
    employment_status VARCHAR(50),
    form_data JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE case_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    worker_id UUID NOT NULL REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    assigned_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE assignment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    case_id UUID NOT NULL REFERENCES cases(id),
    from_worker_id UUID REFERENCES users(id),
    to_worker_id UUID NOT NULL REFERENCES users(id),
    assigned_by UUID REFERENCES users(id),
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE case_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    document_type_id UUID REFERENCES document_types(id),
    file_key VARCHAR(500) NOT NULL,
    original_name VARCHAR(255),
    mime_type VARCHAR(100),
    file_size BIGINT DEFAULT 0,
    verification_status VARCHAR(50) DEFAULT 'pending',
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE workflow_transitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID REFERENCES agencies(id),
    from_status VARCHAR(50) NOT NULL,
    to_status VARCHAR(50) NOT NULL,
    required_role VARCHAR(50) NOT NULL,
    UNIQUE(agency_id, from_status, to_status)
);

CREATE TABLE workflow_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    actor_id UUID REFERENCES users(id),
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE workflow_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    name VARCHAR(255) NOT NULL,
    nodes JSONB DEFAULT '[]',
    edges JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE eligibility_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    program_id UUID NOT NULL REFERENCES programs(id),
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE eligibility_rule_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES eligibility_rules(id) ON DELETE CASCADE,
    version INT NOT NULL,
    conditions JSONB NOT NULL DEFAULT '{}',
    actions JSONB DEFAULT '{}',
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,
    effective_to DATE,
    UNIQUE(rule_id, version)
);

CREATE TABLE eligibility_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    version_id UUID REFERENCES eligibility_rule_versions(id),
    is_eligible BOOLEAN NOT NULL,
    evaluation_trace JSONB DEFAULT '[]',
    evaluated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE benefit_calculation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    program_id UUID NOT NULL REFERENCES programs(id),
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE benefit_calculation_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES benefit_calculation_rules(id) ON DELETE CASCADE,
    version INT NOT NULL,
    formula JSONB NOT NULL DEFAULT '{}',
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,
    effective_to DATE,
    UNIQUE(rule_id, version)
);

CREATE TABLE benefit_calculations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    version_id UUID REFERENCES benefit_calculation_versions(id),
    calculated_amount DECIMAL(12,2) NOT NULL,
    approved_amount DECIMAL(12,2),
    calculation_trace JSONB DEFAULT '[]',
    calculated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE sla_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    program_id UUID NOT NULL REFERENCES programs(id),
    target_days INT NOT NULL,
    warning_threshold_pct INT DEFAULT 80,
    business_days_only BOOLEAN DEFAULT false,
    UNIQUE(agency_id, program_id)
);

CREATE TABLE case_sla_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID UNIQUE NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    sla_policy_id UUID NOT NULL REFERENCES sla_policies(id),
    due_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(50) DEFAULT 'on_track',
    elapsed_days INT DEFAULT 0,
    breached_at TIMESTAMPTZ
);

CREATE TABLE fraud_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    flag_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) DEFAULT 'medium',
    evidence JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'open',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE fraud_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fraud_flag_id UUID UNIQUE NOT NULL REFERENCES fraud_flags(id) ON DELETE CASCADE,
    reviewer_id UUID NOT NULL REFERENCES users(id),
    disposition VARCHAR(50) NOT NULL,
    notes TEXT,
    reviewed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE letter_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    letter_type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    body_template TEXT NOT NULL,
    merge_fields JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE generated_letters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    template_id UUID REFERENCES letter_templates(id),
    letter_type VARCHAR(100) NOT NULL,
    file_key VARCHAR(500),
    generated_by UUID REFERENCES users(id),
    generated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE appeals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    citizen_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'filed',
    grounds TEXT NOT NULL,
    filed_at TIMESTAMPTZ DEFAULT NOW(),
    hearing_date TIMESTAMPTZ
);

CREATE TABLE appeal_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appeal_id UUID NOT NULL REFERENCES appeals(id) ON DELETE CASCADE,
    file_key VARCHAR(500) NOT NULL,
    document_type VARCHAR(100),
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE appeal_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appeal_id UUID UNIQUE NOT NULL REFERENCES appeals(id) ON DELETE CASCADE,
    reviewer_id UUID NOT NULL REFERENCES users(id),
    decision VARCHAR(50) NOT NULL,
    rationale TEXT,
    decided_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE retention_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    entity_type VARCHAR(100) NOT NULL,
    retention_years INT NOT NULL,
    disposition_action VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    UNIQUE(agency_id, entity_type)
);

CREATE TABLE feature_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    flag_key VARCHAR(100) NOT NULL,
    is_enabled BOOLEAN DEFAULT true,
    rollout_pct INT DEFAULT 100,
    metadata JSONB DEFAULT '{}',
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(agency_id, flag_key)
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID REFERENCES agencies(id),
    actor_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID,
    previous_state JSONB,
    new_state JSONB,
    ip_address VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID REFERENCES agencies(id),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(50) DEFAULT 'in_app',
    event_type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    report_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    file_key VARCHAR(500),
    params JSONB DEFAULT '{}',
    requested_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_cases_agency ON cases(agency_id);
CREATE INDEX idx_cases_status ON cases(status);
CREATE INDEX idx_cases_citizen ON cases(citizen_id);
CREATE INDEX idx_cases_number ON cases(case_number);
CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);
CREATE INDEX idx_fraud_case ON fraud_flags(case_id);
CREATE INDEX idx_sla_status ON case_sla_tracking(status);
