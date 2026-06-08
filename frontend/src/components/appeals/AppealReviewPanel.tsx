'use client';

import { useEffect, useState } from 'react';
import { ApiClientError, api } from '@/lib/api/client';
import type { Appeal } from '@/lib/api/types';
import { Textarea } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { formatDate, formatStatus } from '@/lib/utils';

interface AppealReviewPanelProps {
  appeals: Appeal[];
  onDecided?: (message: string) => void | Promise<void>;
}

export function AppealReviewPanel({ appeals, onDecided }: AppealReviewPanelProps) {
  const [selectedId, setSelectedId] = useState('');
  const [rationale, setRationale] = useState('');
  const [loading, setLoading] = useState<'approve' | 'deny' | 'remand' | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (appeals.length === 0) {
      setSelectedId('');
      return;
    }
    if (!appeals.some((a) => a.id === selectedId)) {
      setSelectedId(appeals[0]?.id ?? '');
    }
  }, [appeals, selectedId]);

  const selected = appeals.find((a) => a.id === selectedId);

  const submitDecision = async (decision: 'overturned' | 'upheld' | 'remanded', label: string) => {
    if (!selectedId) return;
    setLoading(decision === 'overturned' ? 'approve' : decision === 'upheld' ? 'deny' : 'remand');
    setError('');
    try {
      await api.post(`/appeals/${selectedId}/decide`, { decision, rationale });
      setRationale('');
      const caseLabel = selected?.case_number ? `Case ${selected.case_number}` : 'Appeal';
      await onDecided?.(`${label} recorded for ${caseLabel}.`);
    } catch (err) {
      const message =
        err instanceof ApiClientError
          ? err.message
          : 'Unable to record appeal decision. Please refresh and try again.';
      setError(message);
    } finally {
      setLoading(null);
    }
  };

  const actionsDisabled = !selectedId || loading !== null;

  if (appeals.length === 0) {
    return <p className="text-sm text-gov-slate">No appeals pending review.</p>;
  }

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <Card title="Pending Appeals">
        <ul className="space-y-3">
          {appeals.map((appeal) => (
            <li key={appeal.id}>
              <button
                type="button"
                onClick={() => setSelectedId(appeal.id)}
                className={`w-full rounded-md border p-3 text-left transition-colors ${
                  selectedId === appeal.id
                    ? 'border-gov-navy bg-gov-surface'
                    : 'border-gov-border hover:bg-gov-surface/50'
                }`}
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="font-medium text-gov-navy">
                    {appeal.case_number || 'Case record'}
                  </span>
                  <Badge variant="warning">{formatStatus(appeal.case_status || appeal.status)}</Badge>
                </div>
                {appeal.program_name && (
                  <p className="mt-1 text-sm text-gov-slate">{appeal.program_name}</p>
                )}
                {appeal.citizen_name && (
                  <p className="text-sm text-gov-slate">{appeal.citizen_name}</p>
                )}
                <p className="mt-2 text-xs text-gov-slate">Filed {formatDate(appeal.filed_at)}</p>
                <p className="mt-2 line-clamp-2 text-sm text-gov-slate">{appeal.grounds}</p>
              </button>
            </li>
          ))}
        </ul>
      </Card>

      <Card title="Appeal Decision">
        {selected && (
          <div className="mb-4 space-y-2 rounded-md bg-gov-surface p-3 text-sm">
            <p className="font-medium text-gov-navy">
              {selected.case_number} · {selected.program_name || 'Program'}
            </p>
            <p className="text-gov-slate">{selected.citizen_name}</p>
            <p className="text-gov-slate">{selected.grounds}</p>
          </div>
        )}
        <div className="space-y-4">
          <Textarea
            label="Decision rationale"
            value={rationale}
            onChange={(e) => setRationale(e.target.value)}
            rows={4}
            placeholder="Document the reason for approving or denying this appeal."
          />
          {error && (
            <p className="text-sm text-gov-danger" role="alert">
              {error}
            </p>
          )}
          <div className="flex flex-wrap gap-3">
            <Button
              onClick={() => submitDecision('overturned', 'Appeal approval')}
              loading={loading === 'approve'}
              disabled={actionsDisabled}
            >
              Approve Appeal
            </Button>
            <Button
              variant="danger"
              onClick={() => submitDecision('upheld', 'Appeal denial')}
              loading={loading === 'deny'}
              disabled={actionsDisabled}
            >
              Deny Appeal
            </Button>
            <Button
              variant="outline"
              onClick={() => submitDecision('remanded', 'Appeal remand')}
              loading={loading === 'remand'}
              disabled={actionsDisabled}
            >
              Remand for Review
            </Button>
          </div>
          <p className="text-xs text-gov-slate">
            Approve overturns the original denial. Deny upholds the original decision. Remand returns the
            case for further worker review.
          </p>
        </div>
      </Card>
    </div>
  );
}
