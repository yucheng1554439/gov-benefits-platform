-- Row-Level Security policies
ALTER TABLE cases ENABLE ROW LEVEL SECURITY;
ALTER TABLE applications ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE fraud_flags ENABLE ROW LEVEL SECURITY;
ALTER TABLE generated_letters ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE appeals ENABLE ROW LEVEL SECURITY;

CREATE POLICY cases_tenant ON cases
    USING (agency_id = NULLIF(current_setting('app.current_agency_id', true), '')::uuid);

CREATE POLICY cases_citizen ON cases FOR SELECT
    USING (citizen_id = NULLIF(current_setting('app.current_user_id', true), '')::uuid);

CREATE POLICY applications_tenant ON applications
    USING (agency_id = NULLIF(current_setting('app.current_agency_id', true), '')::uuid);

CREATE POLICY documents_tenant ON documents
    USING (agency_id = NULLIF(current_setting('app.current_agency_id', true), '')::uuid);

CREATE POLICY fraud_tenant ON fraud_flags
    USING (agency_id = NULLIF(current_setting('app.current_agency_id', true), '')::uuid);

CREATE POLICY letters_tenant ON generated_letters
    USING (agency_id = NULLIF(current_setting('app.current_agency_id', true), '')::uuid);

CREATE POLICY notifications_user ON notifications
    USING (user_id = NULLIF(current_setting('app.current_user_id', true), '')::uuid);

CREATE POLICY audit_tenant ON audit_logs
    USING (agency_id = NULLIF(current_setting('app.current_agency_id', true), '')::uuid);

CREATE POLICY appeals_tenant ON appeals
    USING (agency_id = NULLIF(current_setting('app.current_agency_id', true), '')::uuid);
