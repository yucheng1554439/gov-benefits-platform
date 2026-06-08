# Entity Relationship Diagram

## Core Entities

### Tenancy
- **agencies** — LA County DPSS, LA County DHS, City of Los Angeles, California DSS
- **agency_users** — user membership in agencies
- **agency_programs** — programs enabled per agency

### Users & Auth
- **users** — email, password_hash, status
- **roles** — citizen, case_worker, supervisor, admin
- **user_roles** — many-to-many
- **user_profiles** — name, phone, address, ssn_hash
- **worker_profiles** — specializations, workload

### Case Management
- **cases** — case_number, status, priority, zip_code, census_tract
- **applications** — household_size, annual_income, employment_status
- **case_assignments** — worker assignment
- **assignment_history** — reassignment audit trail
- **case_notes** — worker notes
- **workflow_events** — state transition history
- **workflow_transitions** — allowed transitions
- **workflow_definitions** — React Flow JSON

### Documents
- **document_types** — government_id, pay_stubs, tax_forms, utility_bills
- **documents** — file metadata, verification status

### Eligibility & Benefits
- **eligibility_rules** + **eligibility_rule_versions** — JSON condition trees
- **eligibility_evaluations** — per-case results
- **benefit_calculation_rules** + **benefit_calculation_versions** — formula JSON
- **benefit_calculations** — calculated and approved amounts

### Compliance
- **sla_policies** — target days per program
- **case_sla_tracking** — on_track, at_risk, overdue, met, breached
- **fraud_flags** + **fraud_reviews** — duplicate detection
- **retention_policies** — entity retention schedules
- **feature_flags** — per-agency module toggles

### Appeals
- **appeals** — grounds, status, hearing_date
- **appeal_documents** — supporting evidence
- **appeal_decisions** — reviewer decision (immutable)

### Letters & Reports
- **letter_templates** — Go text/template bodies
- **generated_letters** — PDF file references
- **reports** — generated report jobs

### System
- **audit_logs** — immutable action log
- **notifications** — in-app and email

## Key Relationships

```
agencies 1──* cases
cases 1──1 applications
cases 1──* documents
cases 1──* workflow_events
cases 1──* benefit_calculations
cases 1──* appeals
cases 1──* fraud_flags
cases 1──1 case_sla_tracking
users 1──* cases (as citizen)
users 1──* case_assignments (as worker)
```

## Case Status Flow

```
submitted → under_review → need_documents ↔ under_review
                        → eligibility_review → supervisor_review → approved/denied
denied → appealed → appeal_review → appeal_approved/appeal_denied
approved/denied → closed
```
