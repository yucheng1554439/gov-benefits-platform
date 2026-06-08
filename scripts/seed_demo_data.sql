-- Curated demonstration dataset for portfolio recordings.
-- Run AFTER scripts/reset_demo_data.sql

BEGIN;

-- Fixed demo case IDs
-- Case A: Food Assistance — Approved (full success path)
-- Case B: Housing Assistance — Appealed (pending supervisor decision)
-- Case C: Emergency Relief — Fraud flag under worker review
-- Case D: Healthcare Assistance — Need Documents

INSERT INTO cases (id, agency_id, case_number, citizen_id, program_id, status, priority, zip_code, census_tract) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', '22222222-2222-2222-2222-222222222201', 'CASE-2026-DEMO-A', '55555555-5555-5555-5555-555555555501', '33333333-3333-3333-3333-333333333302', 'approved', 'normal', '90001', '6037400100'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', '22222222-2222-2222-2222-222222222201', 'CASE-2026-DEMO-B', '55555555-5555-5555-5555-555555555502', '33333333-3333-3333-3333-333333333301', 'appealed', 'high', '90012', '6037410200'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3', '22222222-2222-2222-2222-222222222201', 'CASE-2026-DEMO-C', '55555555-5555-5555-5555-555555555501', '33333333-3333-3333-3333-333333333304', 'under_review', 'urgent', '90001', '6037400100'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4', '22222222-2222-2222-2222-222222222201', 'CASE-2026-DEMO-D', '55555555-5555-5555-5555-555555555502', '33333333-3333-3333-3333-333333333303', 'need_documents', 'normal', '90012', '6037410200');

INSERT INTO applications (agency_id, case_id, household_size, annual_income, employment_status, form_data) VALUES
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 3, 28000, 'employed_part_time', '{}'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 2, 52000, 'employed_full_time', '{}'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3', 1, 18000, 'unemployed', '{}'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4', 4, 31000, 'employed_part_time', '{}');

INSERT INTO case_assignments (case_id, worker_id, is_active) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', '55555555-5555-5555-5555-555555555503', true),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', '55555555-5555-5555-5555-555555555503', true),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3', '55555555-5555-5555-5555-555555555504', true),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4', '55555555-5555-5555-5555-555555555504', true);

UPDATE worker_profiles SET current_case_count = 2 WHERE user_id = '55555555-5555-5555-5555-555555555503';
UPDATE worker_profiles SET current_case_count = 2 WHERE user_id = '55555555-5555-5555-5555-555555555504';

-- Case A: eligibility + benefit (Food Assistance rules from seed)
INSERT INTO eligibility_evaluations (case_id, version_id, is_eligible, evaluation_trace)
SELECT 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', erv.id, true, '[{"step":"annual_income","passed":true}]'::jsonb
FROM eligibility_rule_versions erv
JOIN eligibility_rules er ON er.id = erv.rule_id
WHERE er.program_id = '33333333-3333-3333-3333-333333333302'
ORDER BY erv.version DESC LIMIT 1;

INSERT INTO benefit_calculations (case_id, version_id, calculated_amount, calculation_trace)
SELECT 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', bcv.id, 168.00, '[{"step":"base_benefit","value":168}]'::jsonb
FROM benefit_calculation_versions bcv
JOIN benefit_calculation_rules bcr ON bcr.id = bcv.rule_id
WHERE bcr.program_id = '33333333-3333-3333-3333-333333333302'
ORDER BY bcv.version DESC LIMIT 1;

INSERT INTO generated_letters (agency_id, case_id, template_id, letter_type, file_key, generated_by)
SELECT '22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', lt.id, 'approval_notice',
       'demo/letters/case-a-approval.pdf', '55555555-5555-5555-5555-555555555503'
FROM letter_templates lt
WHERE lt.agency_id = '22222222-2222-2222-2222-222222222201' AND lt.letter_type = 'approval_notice'
LIMIT 1;

-- Case B: denial history + filed appeal (pending decision)
INSERT INTO appeals (id, agency_id, case_id, citizen_id, status, grounds)
VALUES (
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2',
    '22222222-2222-2222-2222-222222222201',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2',
    '55555555-5555-5555-5555-555555555502',
    'filed',
    'Income documentation was not considered. Attached updated lease and pay stubs support eligibility.'
);

INSERT INTO generated_letters (agency_id, case_id, template_id, letter_type, file_key, generated_by)
SELECT '22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', lt.id, 'denial_notice',
       'demo/letters/case-b-denial.pdf', '55555555-5555-5555-5555-555555555505'
FROM letter_templates lt
WHERE lt.agency_id = '22222222-2222-2222-2222-222222222201' AND lt.letter_type = 'denial_notice'
LIMIT 1;

-- Case C: fraud flag
INSERT INTO fraud_flags (agency_id, case_id, flag_type, severity, evidence, status)
VALUES (
    '22222222-2222-2222-2222-222222222201',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3',
    'duplicate_application',
    'medium',
    '{"reason":"Matching address on prior emergency claim within 30 days"}'::jsonb,
    'open'
);

-- Workflow history (representative events)
INSERT INTO workflow_events (agency_id, case_id, from_status, to_status, actor_id, reason) VALUES
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 'submitted', 'under_review', '55555555-5555-5555-5555-555555555503', 'Initial review'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 'under_review', 'eligibility_review', '55555555-5555-5555-5555-555555555503', 'Eligibility determination'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 'eligibility_review', 'approved', '55555555-5555-5555-5555-555555555505', 'Supervisor approval'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 'submitted', 'under_review', '55555555-5555-5555-5555-555555555503', 'Initial review'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 'supervisor_review', 'denied', '55555555-5555-5555-5555-555555555505', 'Income exceeds program threshold'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 'denied', 'appealed', '55555555-5555-5555-5555-555555555502', 'Citizen filed appeal'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3', 'submitted', 'under_review', '55555555-5555-5555-5555-555555555504', 'Initial review'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4', 'submitted', 'under_review', '55555555-5555-5555-5555-555555555504', 'Initial review'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4', 'under_review', 'need_documents', '55555555-5555-5555-5555-555555555504', 'Missing proof of residency');

-- Audit trail samples
INSERT INTO audit_logs (agency_id, actor_id, action, entity_type, entity_id, new_state) VALUES
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555501', 'application.created', 'event', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', '{"case_number":"CASE-2026-DEMO-A"}'::jsonb),
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555503', 'eligibility.evaluated', 'event', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', '{"is_eligible":true}'::jsonb),
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555503', 'benefit.calculated', 'event', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', '{"amount":168}'::jsonb),
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555505', 'case.status_changed', 'event', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', '{"from_status":"eligibility_review","to_status":"approved"}'::jsonb),
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555502', 'appeal.filed', 'event', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2', '{"case_id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2"}'::jsonb),
    ('22222222-2222-2222-2222-222222222201', '55555555-5555-5555-5555-555555555504', 'fraud.flagged', 'event', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3', '{"flag_type":"duplicate_application"}'::jsonb);

COMMIT;
