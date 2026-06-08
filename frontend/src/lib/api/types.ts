export interface User {
  id: string;
  email: string;
  status: string;
  created_at: string;
}

export interface UserProfile {
  id: string;
  user_id: string;
  first_name: string;
  last_name: string;
  phone?: string;
  address?: Record<string, string>;
}

export interface AuthUser {
  user: User;
  profile?: UserProfile;
  roles: string[];
  agency_id: string;
  agency_role: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  user: AuthUser;
}

export interface Agency {
  id: string;
  code: string;
  name: string;
  type: string;
  jurisdiction: string;
}

export interface Program {
  id: string;
  code: string;
  name: string;
  description?: string;
}

export interface Case {
  id: string;
  agency_id: string;
  case_number: string;
  citizen_id: string;
  program_id: string;
  status: string;
  priority: string;
  zip_code?: string;
  census_tract?: string;
  submitted_at: string;
  closed_at?: string;
  created_at: string;
  updated_at: string;
  program?: Program;
  application?: Application;
}

export interface Application {
  id: string;
  agency_id: string;
  case_id: string;
  household_size: number;
  annual_income: number;
  employment_status: string;
  form_data?: Record<string, unknown>;
}

export interface Appeal {
  id: string;
  agency_id: string;
  case_id: string;
  citizen_id: string;
  status: string;
  grounds: string;
  filed_at: string;
  hearing_date?: string;
  case_number?: string;
  program_name?: string;
  citizen_name?: string;
  case_status?: string;
}

export interface FraudFlag {
  id: string;
  agency_id: string;
  case_id: string;
  flag_type: string;
  severity: string;
  evidence?: Record<string, unknown>;
  status: string;
  created_at: string;
}

export interface CaseSLATracking {
  id: string;
  case_id: string;
  sla_policy_id: string;
  due_at: string;
  status: string;
  elapsed_days: number;
  breached_at?: string;
}

export interface WorkflowEvent {
  id: string;
  agency_id: string;
  case_id: string;
  from_status?: string;
  to_status: string;
  actor_id?: string;
  actor_name?: string;
  reason?: string;
  created_at: string;
}

export interface WorkflowTransition {
  id: string;
  agency_id?: string;
  from_status: string;
  to_status: string;
  required_role: string;
}

export interface Notification {
  id: string;
  user_id: string;
  title: string;
  body: string;
  is_read: boolean;
  created_at: string;
}

export interface Letter {
  id: string;
  case_id: string;
  letter_type: string;
  file_key?: string;
  generated_at: string;
}

export interface FeatureFlag {
  id?: string;
  agency_id: string;
  flag_key: string;
  is_enabled: boolean;
  rollout_pct?: number;
}

export interface RetentionPolicy {
  id: string;
  agency_id: string;
  entity_type: string;
  retention_years: number;
  disposition_action: string;
}

export interface Report {
  id: string;
  agency_id: string;
  report_type: string;
  status: string;
  created_at: string;
}

export interface AnalyticsSummary {
  case_status_counts: Record<string, number>;
  cases_by_zip: Record<string, number>;
  open_fraud_flags: number;
}

export interface ApiListResponse<T> {
  data: T[];
}

export interface ApiError {
  error: string;
}
