'use client';

import { use, useEffect, useState } from 'react';
import { ApiClientError, api } from '@/lib/api/client';
import { useCase } from '@/lib/hooks/useCases';
import { Card } from '@/components/ui/Card';
import { Badge, statusBadgeVariant } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Input';
import { CaseTimeline } from '@/components/cases/CaseTimeline';
import { FraudFlagBanner } from '@/components/cases/FraudFlagBanner';
import { SLABadge } from '@/components/cases/SLABadge';
import { EligibilityCard } from '@/components/cases/EligibilityCard';
import { BenefitAmountCard } from '@/components/cases/BenefitAmountCard';
import { LetterActionsCard } from '@/components/cases/LetterActionsCard';
import { formatDate, formatStatus } from '@/lib/utils';

function formatActionError(err: unknown, fallback: string): string {
  if (err instanceof ApiClientError) return err.message;
  if (err instanceof Error && err.message) return err.message;
  return fallback;
}

export default function WorkerCaseDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { caseData, loading, error, refetch } = useCase(id);
  const [transitions, setTransitions] = useState<string[]>([]);
  const [newStatus, setNewStatus] = useState('');
  const [updating, setUpdating] = useState(false);
  const [scanning, setScanning] = useState(false);
  const [actionMessage, setActionMessage] = useState('');
  const [actionError, setActionError] = useState('');
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    if (!id || !caseData) return;
    api
      .get<{ data: string[] }>(`/cases/${id}/transitions`)
      .then((res) => {
        const options = res.data ?? [];
        setTransitions(options);
        setNewStatus((current) => (options.includes(current) ? current : options[0] ?? ''));
      })
      .catch(() => setTransitions([]));
  }, [id, caseData?.status]);

  const updateStatus = async () => {
    if (!newStatus) {
      setActionError('Select a status to continue.');
      return;
    }
    setUpdating(true);
    setActionMessage('');
    setActionError('');
    try {
      await api.patch(`/cases/${id}/status`, { to_status: newStatus });
      setActionMessage(`Case status updated to ${formatStatus(newStatus)}.`);
      await refetch();
      setRefreshKey((key) => key + 1);
    } catch (err) {
      setActionError(formatActionError(err, 'Unable to update case status.'));
    } finally {
      setUpdating(false);
    }
  };

  const runFraudScan = async () => {
    setScanning(true);
    setActionMessage('');
    setActionError('');
    try {
      const res = await api.post<{ data: unknown[]; count?: number }>(`/cases/${id}/fraud/scan`);
      const count = res.count ?? res.data?.length ?? 0;
      setActionMessage(count > 0 ? `Fraud scan found ${count} issue${count === 1 ? '' : 's'}.` : 'Fraud scan complete. No issues found.');
      setRefreshKey((key) => key + 1);
    } catch (err) {
      setActionError(formatActionError(err, 'Unable to run fraud scan.'));
    } finally {
      setScanning(false);
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
          <p className="text-gov-slate">{caseData.program?.name}</p>
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
          <Card title="Case Information">
            <dl className="grid gap-3 sm:grid-cols-2 text-sm">
              <div><dt className="text-gov-slate">Submitted</dt><dd>{formatDate(caseData.submitted_at)}</dd></div>
              <div><dt className="text-gov-slate">Priority</dt><dd className="capitalize">{caseData.priority}</dd></div>
              {caseData.application && (
                <>
                  <div><dt className="text-gov-slate">Household</dt><dd>{caseData.application.household_size}</dd></div>
                  <div><dt className="text-gov-slate">Income</dt><dd>${caseData.application.annual_income.toLocaleString()}</dd></div>
                </>
              )}
            </dl>
          </Card>
          <Card title="Update Status">
            {transitions.length === 0 ? (
              <p className="text-sm text-gov-slate">No status changes are available for the current workflow state.</p>
            ) : (
              <div className="flex flex-wrap gap-3">
                <Select
                  value={newStatus}
                  onChange={(e) => setNewStatus(e.target.value)}
                  options={transitions.map((s) => ({ value: s, label: formatStatus(s) }))}
                  className="min-w-[200px]"
                />
                <Button onClick={updateStatus} loading={updating} disabled={!newStatus}>
                  Update
                </Button>
                <Button variant="outline" onClick={runFraudScan} loading={scanning}>
                  Fraud Scan
                </Button>
              </div>
            )}
          </Card>
          <Card title="Timeline"><CaseTimeline caseId={id} refreshKey={refreshKey} /></Card>
        </div>
        <div className="space-y-6">
          <EligibilityCard caseId={id} canEvaluate />
          <BenefitAmountCard caseId={id} canCalculate />
          <LetterActionsCard caseId={id} caseStatus={caseData.status} />
        </div>
      </div>
    </div>
  );
}
