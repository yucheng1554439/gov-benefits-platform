-- Reset transactional demo data while preserving configuration and accounts.
-- Safe to run repeatedly before a recorded demonstration.

BEGIN;

-- Case-scoped transactional data (order respects FK dependencies)
DELETE FROM appeal_documents;
DELETE FROM appeal_decisions;
DELETE FROM appeals;
DELETE FROM fraud_reviews;
DELETE FROM fraud_flags;
DELETE FROM generated_letters;
DELETE FROM benefit_calculations;
DELETE FROM eligibility_evaluations;
DELETE FROM documents;
DELETE FROM case_sla_tracking;
DELETE FROM workflow_events;
DELETE FROM case_notes;
DELETE FROM assignment_history;
DELETE FROM case_assignments;
DELETE FROM applications;
DELETE FROM cases;

-- Agency-wide transactional data (audit logs are immutable by default)
ALTER TABLE audit_logs DISABLE TRIGGER audit_logs_no_delete;
DELETE FROM audit_logs;
ALTER TABLE audit_logs ENABLE TRIGGER audit_logs_no_delete;
DELETE FROM notifications;
DELETE FROM reports;

-- Reset worker assignment counters
UPDATE worker_profiles SET current_case_count = 0;

COMMIT;
