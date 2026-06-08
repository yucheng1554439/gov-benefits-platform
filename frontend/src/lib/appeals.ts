import type { Appeal } from '@/lib/api/types';

const TERMINAL_CASE_STATUSES = new Set([
  'appeal_approved',
  'appeal_denied',
  'approved',
  'closed',
]);

export function isPendingAppeal(appeal: Appeal): boolean {
  if (appeal.status === 'decided') return false;
  if (TERMINAL_CASE_STATUSES.has(appeal.case_status ?? '')) return false;
  return appeal.status === 'filed' || appeal.status === 'pending';
}

export function isDecidedAppeal(appeal: Appeal): boolean {
  return appeal.status === 'decided' || TERMINAL_CASE_STATUSES.has(appeal.case_status ?? '');
}
