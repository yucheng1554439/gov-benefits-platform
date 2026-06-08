'use client';

import { use, useEffect, useState } from 'react';
import { ApiClientError, api } from '@/lib/api/client';
import { useCase } from '@/lib/hooks/useCases';
import { Card } from '@/components/ui/Card';
import { Badge, statusBadgeVariant } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { CaseTimeline } from '@/components/cases/CaseTimeline';
import { FraudFlagBanner } from '@/components/cases/FraudFlagBanner';
import { SLABadge } from '@/components/cases/SLABadge';
import { BenefitAmountCard } from '@/components/cases/BenefitAmountCard';
import { AppealReviewPanel } from '@/components/appeals/AppealReviewPanel';
import type { Appeal, ApiListResponse } from '@/lib/api/types';
import { isPendingAppeal } from '@/lib/appeals';
import { formatDate, formatStatus } from '@/lib/utils';

function formatActionError(err: unknown, fallback: string): string {
  if (err instanceof ApiClientError) return err.message;
  if (err instanceof Error && err.message) return err.message;
  return fallback;
}

export default function SupervisorCaseDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { caseData, loading, error, refetch } = useCase(id);
  const [appeals, setAppeals] = useState<Appeal[]>([]);
  const [updating, setUpdating] = useState(false);
  const [actionMessage, setActionMessage] = useState('');
  const [actionError, setActionError] = useState('');
  const [refreshKey, setRefreshKey] = useState(0);

  const loadAppeals = () => {
    api
      .get<ApiListResponse<Appeal>>(`/cases/${id}/appeals`)
      .then((res) => setAppeals((res.data ?? []).filter(isPendingAppeal)));
  };

  const handleAppealDecided = async (message: string) => {
    setActionMessage(message);
    loadAppeals();
    await refetch();
    setRefreshKey((key) => key + 1);
  };

  useEffect(() => {
    loadAppeals();
  }, [id]);

  const decide = async (status: string) => {
    setUpdating(true);
    setActionMessage('');
    setActionError('');
    try {
      await api.patch(`/cases/${id}/status`, { to_status: status });
      setActionMessage(`Case status updated to ${formatStatus(status)}.`);
      await refetch();
      setRefreshKey((key) => key + 1);
    } catch (err) {
      setActionError(formatActionError(err, 'Unable to update case status.'));
    } finally {
      setUpdating(false);
    }
  };

  if (loading) return <p className="text-gov-slate">Loading...</p>;
  if (error || !caseData) return <p className="text-gov-danger">{error ?? 'Not found'}</p>;

  return (
    <div className="space-y-6">
      <FraudFlagBanner caseId={id} refreshKey={refreshKey} />
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gov-navy">{caseData.case_number}</h1>
          <p className="text-gov-slate">Supervisor Review</p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant={statusBadgeVariant(caseData.status)}>{formatStatus(caseData.status)}</Badge>
          <SLABadge caseId={id} />
        </div>
      </div>

      {(actionMessage || actionError) && (
        <div
          className={`rounded-md border px-4 py-3 text-sm ${
            actionError
              ? 'border-gov-danger/30 bg-red-50 text-gov-danger'
              : 'border-green-300 bg-green-50 text-green-800'
          }`}
          role="alert"
          aria-live="polite"
        >
          {actionError || actionMessage}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <Card title="Case Summary">
            <dl className="grid gap-3 sm:grid-cols-2 text-sm">
              <div><dt className="text-gov-slate">Program</dt><dd>{caseData.program?.name}</dd></div>
              <div><dt className="text-gov-slate">Submitted</dt><dd>{formatDate(caseData.submitted_at)}</dd></div>
            </dl>
            {caseData.status === 'supervisor_review' && (
              <div className="mt-4 flex gap-2">
                <Button onClick={() => decide('approved')} loading={updating}>Approve</Button>
                <Button variant="danger" onClick={() => decide('denied')} loading={updating}>Deny</Button>
              </div>
            )}
          </Card>
          <Card title="Timeline"><CaseTimeline caseId={id} refreshKey={refreshKey} /></Card>
          {(caseData.status === 'appealed' || caseData.status === 'appeal_review') && (
            <Card title="Appeal Review">
              <AppealReviewPanel appeals={appeals} onDecided={handleAppealDecided} />
            </Card>
          )}
          {appeals.length === 0 &&
            (caseData.status === 'appeal_approved' || caseData.status === 'appeal_denied') && (
              <Card title="Appeal Review">
                <p className="text-sm text-gov-slate">
                  Appeal decision recorded. Case status: {formatStatus(caseData.status)}.
                </p>
              </Card>
            )}
        </div>
        <BenefitAmountCard caseId={id} canCalculate />
      </div>
    </div>
  );
}
