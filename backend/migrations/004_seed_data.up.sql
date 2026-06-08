-- Roles
INSERT INTO roles (id, name, description) VALUES
    ('11111111-1111-1111-1111-111111111101', 'citizen', 'Citizen applicant'),
    ('11111111-1111-1111-1111-111111111102', 'case_worker', 'Case worker'),
    ('11111111-1111-1111-1111-111111111103', 'supervisor', 'Supervisor'),
    ('11111111-1111-1111-1111-111111111104', 'admin', 'System administrator');

-- Agencies
INSERT INTO agencies (id, code, name, type, jurisdiction) VALUES
    ('22222222-2222-2222-2222-222222222201', 'LAC_DPSS', 'LA County DPSS', 'county', 'Los Angeles County'),
    ('22222222-2222-2222-2222-222222222202', 'LAC_DHS', 'LA County DHS', 'county', 'Los Angeles County'),
    ('22222222-2222-2222-2222-222222222203', 'CITY_LA', 'City of Los Angeles', 'city', 'City of Los Angeles'),
    ('22222222-2222-2222-2222-222222222204', 'CA_DSS', 'California DSS', 'state', 'California');

-- Programs
INSERT INTO programs (id, code, name, description) VALUES
    ('33333333-3333-3333-3333-333333333301', 'housing_assistance', 'Housing Assistance', 'Rental and housing support'),
    ('33333333-3333-3333-3333-333333333302', 'food_assistance', 'Food Assistance', 'Nutrition and food benefits'),
    ('33333333-3333-3333-3333-333333333303', 'healthcare_assistance', 'Healthcare Assistance', 'Medical coverage assistance'),
    ('33333333-3333-3333-3333-333333333304', 'emergency_relief', 'Emergency Relief', 'Emergency financial assistance');

-- Document types
INSERT INTO document_types (id, code, name) VALUES
    ('44444444-4444-4444-4444-444444444401', 'government_id', 'Government ID'),
    ('44444444-4444-4444-4444-444444444402', 'pay_stubs', 'Pay Stubs'),
    ('44444444-4444-4444-4444-444444444403', 'tax_forms', 'Tax Forms'),
    ('44444444-4444-4444-4444-444444444404', 'utility_bills', 'Utility Bills');

-- Agency programs (all agencies get all programs)
INSERT INTO agency_programs (agency_id, program_id, is_enabled)
SELECT a.id, p.id, true FROM agencies a CROSS JOIN programs p;

-- Password hash for Password123! (bcrypt cost 12)
-- Demo password: Password123!
-- Hash: $2a$12$5pDrK0SKqeHdwLh4lkjlv.zZYeHOJegXzaHS/8K1wGgkHq7WDT1i.

-- Users (password: Password123!)
INSERT INTO users (id, email, password_hash, status) VALUES
    ('55555555-5555-5555-5555-555555555501', 'citizen1@example.com', '$2a$12$5pDrK0SKqeHdwLh4lkjlv.zZYeHOJegXzaHS/8K1wGgkHq7WDT1i.', 'active'),
    ('55555555-5555-5555-5555-555555555502', 'citizen2@example.com', '$2a$12$5pDrK0SKqeHdwLh4lkjlv.zZYeHOJegXzaHS/8K1wGgkHq7WDT1i.', 'active'),
    ('55555555-5555-5555-5555-555555555503', 'worker1@dpss.lacounty.gov', '$2a$12$5pDrK0SKqeHdwLh4lkjlv.zZYeHOJegXzaHS/8K1wGgkHq7WDT1i.', 'active'),
    ('55555555-5555-5555-5555-555555555504', 'worker2@dpss.lacounty.gov', '$2a$12$5pDrK0SKqeHdwLh4lkjlv.zZYeHOJegXzaHS/8K1wGgkHq7WDT1i.', 'active'),
    ('55555555-5555-5555-5555-555555555505', 'supervisor1@dpss.lacounty.gov', '$2a$12$5pDrK0SKqeHdwLh4lkjlv.zZYeHOJegXzaHS/8K1wGgkHq7WDT1i.', 'active'),
    ('55555555-5555-5555-5555-555555555506', 'admin@dpss.lacounty.gov', '$2a$12$5pDrK0SKqeHdwLh4lkjlv.zZYeHOJegXzaHS/8K1wGgkHq7WDT1i.', 'active');

INSERT INTO user_profiles (user_id, first_name, last_name, phone, address) VALUES
    ('55555555-5555-5555-5555-555555555501', 'Maria', 'Garcia', '213-555-0101', '{"street":"123 Main St","city":"Los Angeles","state":"CA","zip":"90001"}'),
    ('55555555-5555-5555-5555-555555555502', 'James', 'Wilson', '213-555-0102', '{"street":"456 Oak Ave","city":"Los Angeles","state":"CA","zip":"90012"}'),
    ('55555555-5555-5555-5555-555555555503', 'Sarah', 'Chen', '213-555-0201', '{}'),
    ('55555555-5555-5555-5555-555555555504', 'David', 'Martinez', '213-555-0202', '{}'),
    ('55555555-5555-5555-5555-555555555505', 'Patricia', 'Johnson', '213-555-0301', '{}'),
    ('55555555-5555-5555-5555-555555555506', 'Robert', 'Admin', '213-555-0401', '{}');

INSERT INTO user_roles (user_id, role_id) VALUES
    ('55555555-5555-5555-5555-555555555501', '11111111-1111-1111-1111-111111111101'),
    ('55555555-5555-5555-5555-555555555502', '11111111-1111-1111-1111-111111111101'),
    ('55555555-5555-5555-5555-555555555503', '11111111-1111-1111-1111-111111111102'),
    ('55555555-5555-5555-5555-555555555504', '11111111-1111-1111-1111-111111111102'),
    ('55555555-5555-5555-5555-555555555505', '11111111-1111-1111-1111-111111111103'),
    ('55555555-5555-5555-5555-555555555506', '11111111-1111-1111-1111-111111111104');

INSERT INTO agency_users (agency_id, user_id, agency_role, is_primary) VALUES
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555501', 'citizen', true),
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555502', 'citizen', true),
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555503', 'worker', true),
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555504', 'worker', true),
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555505', 'supervisor', true),
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555506', 'admin', true);

INSERT INTO worker_profiles (user_id, agency_id, specializations, max_active_cases, current_case_count) VALUES
    ('55555555-5555-5555-5555-555555555503', '22222222-2222-2222-2222-222222222201', ARRAY['food_assistance','housing_assistance'], 50, 0),
    ('55555555-5555-5555-5555-555555555504', '22222222-2222-2222-2222-222222222201', ARRAY['emergency_relief','healthcare_assistance'], 50, 0);

-- Workflow transitions
INSERT INTO workflow_transitions (agency_id, from_status, to_status, required_role) VALUES
    ('22222222-2222-2222-2222-222222222201', 'submitted', 'under_review', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'under_review', 'need_documents', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'under_review', 'eligibility_review', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'under_review', 'denied', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'need_documents', 'under_review', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'eligibility_review', 'supervisor_review', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'eligibility_review', 'approved', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'eligibility_review', 'denied', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'supervisor_review', 'approved', 'supervisor'),
    ('22222222-2222-2222-2222-222222222201', 'supervisor_review', 'denied', 'supervisor'),
    ('22222222-2222-2222-2222-222222222201', 'approved', 'closed', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'denied', 'appealed', 'citizen'),
    ('22222222-2222-2222-2222-222222222201', 'denied', 'closed', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'appealed', 'appeal_review', 'case_worker'),
    ('22222222-2222-2222-2222-222222222201', 'appeal_review', 'appeal_approved', 'supervisor'),
    ('22222222-2222-2222-2222-222222222201', 'appeal_review', 'appeal_denied', 'supervisor'),
    ('22222222-2222-2222-2222-222222222201', 'appeal_approved', 'approved', 'supervisor'),
    ('22222222-2222-2222-2222-222222222201', 'appeal_denied', 'closed', 'supervisor');

-- SLA policies
INSERT INTO sla_policies (agency_id, program_id, target_days, warning_threshold_pct, business_days_only) VALUES
    ('22222222-2222-2222-2222-222222222201', '33333333-3333-3333-3333-333333333302', 30, 80, false),
    ('22222222-2222-2222-2222-222222222201', '33333333-3333-3333-3333-333333333304', 1, 50, false),
    ('22222222-2222-2222-2222-222222222201', '33333333-3333-3333-3333-333333333301', 45, 80, false),
    ('22222222-2222-2222-2222-222222222201', '33333333-3333-3333-3333-333333333303', 14, 80, false);

-- Eligibility rules
INSERT INTO eligibility_rules (id, agency_id, program_id, name, is_active) VALUES
    ('66666666-6666-6666-6666-666666666601', '22222222-2222-2222-2222-222222222201', '33333333-3333-3333-3333-333333333302', 'Food Assistance Income Rule', true);

INSERT INTO eligibility_rule_versions (rule_id, version, conditions, effective_from) VALUES
    ('66666666-6666-6666-6666-666666666601', 1, '{"operator":"AND","rules":[{"field":"annual_income","op":"lt","value":35000},{"field":"household_size","op":"gte","value":1}]}', CURRENT_DATE);

-- Benefit calculation rules
INSERT INTO benefit_calculation_rules (id, agency_id, program_id, name, is_active) VALUES
    ('77777777-7777-7777-7777-777777777701', '22222222-2222-2222-2222-222222222201', '33333333-3333-3333-3333-333333333302', 'Food Assistance Benefit Formula', true);

INSERT INTO benefit_calculation_versions (rule_id, version, formula, effective_from) VALUES
    ('77777777-7777-7777-7777-777777777701', 1, '{"baseBenefit":350,"householdMultiplier":1.4,"maxBenefit":1200}', CURRENT_DATE);

-- Letter templates
INSERT INTO letter_templates (agency_id, letter_type, name, body_template, merge_fields) VALUES
    ('22222222-2222-2222-2222-222222222201', 'approval_notice', 'Approval Notice',
     'Dear {{.CitizenName}},\n\nYour application for {{.ProgramName}} has been APPROVED.\nBenefit Amount: ${{.BenefitAmount}}/month\nCase Number: {{.CaseNumber}}\n\nSincerely,\nLA County DPSS',
     '["CitizenName","ProgramName","BenefitAmount","CaseNumber"]'),
    ('22222222-2222-2222-2222-222222222201', 'denial_notice', 'Denial Notice',
     'Dear {{.CitizenName}},\n\nYour application for {{.ProgramName}} has been DENIED.\nReason: {{.DenialReason}}\n\nYou have 90 days to file an appeal.\nCase Number: {{.CaseNumber}}\n\nSincerely,\nLA County DPSS',
     '["CitizenName","ProgramName","DenialReason","CaseNumber"]');

-- Retention policies
INSERT INTO retention_policies (agency_id, entity_type, retention_years, disposition_action) VALUES
    ('22222222-2222-2222-2222-222222222201', 'case_records', 7, 'archive'),
    ('22222222-2222-2222-2222-222222222201', 'audit_logs', 10, 'retain'),
    ('22222222-2222-2222-2222-222222222201', 'documents', 7, 'purge'),
    ('22222222-2222-2222-2222-222222222201', 'generated_letters', 7, 'archive');

-- Feature flags
INSERT INTO feature_flags (agency_id, flag_key, is_enabled) VALUES
    ('22222222-2222-2222-2222-222222222201', 'new_workflow_engine', true),
    ('22222222-2222-2222-2222-222222222201', 'appeals_module', true),
    ('22222222-2222-2222-2222-222222222201', 'fraud_detection', true),
    ('22222222-2222-2222-2222-222222222201', 'geo_analytics', true),
    ('22222222-2222-2222-2222-222222222201', 'benefit_calculation', true),
    ('22222222-2222-2222-2222-222222222201', 'retention_policies', true),
    ('22222222-2222-2222-2222-222222222204', 'geo_analytics', false);

-- Sample cases
INSERT INTO cases (id, agency_id, case_number, citizen_id, program_id, status, priority, zip_code, census_tract) VALUES
    ('88888888-8888-8888-8888-888888888801', '22222222-2222-2222-2222-222222222201', 'CASE-2026-000001', '55555555-5555-5555-5555-555555555501', '33333333-3333-3333-3333-333333333302', 'under_review', 'normal', '90001', '6037400100'),
    ('88888888-8888-8888-8888-888888888802', '22222222-2222-2222-2222-222222222201', 'CASE-2026-000002', '55555555-5555-5555-5555-555555555502', '33333333-3333-3333-3333-333333333301', 'submitted', 'high', '90012', '6037410200');

INSERT INTO applications (agency_id, case_id, household_size, annual_income, employment_status, form_data) VALUES
    ('22222222-2222-2222-2222-222222222201', '88888888-8888-8888-8888-888888888801', 3, 28000, 'employed_part_time', '{}'),
    ('22222222-2222-2222-2222-222222222201', '88888888-8888-8888-8888-888888888802', 2, 45000, 'employed_full_time', '{}');

INSERT INTO case_assignments (case_id, worker_id, is_active) VALUES
    ('88888888-8888-8888-8888-888888888801', '55555555-5555-5555-5555-555555555503', true);

UPDATE worker_profiles SET current_case_count = 1 WHERE user_id = '55555555-5555-5555-5555-555555555503';
