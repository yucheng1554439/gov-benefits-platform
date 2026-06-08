'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { api } from '@/lib/api/client';
import type { Appeal, ApiListResponse } from '@/lib/api/types';
import { AppealReviewPanel } from '@/components/appeals/AppealReviewPanel';
import { isPendingAppeal } from '@/lib/appeals';
import { Card } from '@/components/ui/Card';

export default function SupervisorAppealsPage() {
  const [appeals, setAppeals] = useState<Appeal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  const loadAppeals = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await api.get<ApiListResponse<Appeal>>('/appeals?pending=true');
      setAppeals(res.data ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load appeals.');
      setAppeals([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadAppeals();
  }, [loadAppeals]);

  const pendingAppeals = useMemo(() => appeals.filter(isPendingAppeal), [appeals]);

  const handleDecided = async (message: string) => {
    setSuccessMessage(message);
    await loadAppeals();
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Appeal Review</h1>
        <p className="text-gov-slate">
          Review citizen appeals and approve or deny decisions. Pending: {pendingAppeals.length}
        </p>
      </div>

      {successMessage && (
        <div
          className="rounded-md border border-green-300 bg-green-50 px-4 py-3 text-sm text-green-800"
          role="status"
          aria-live="polite"
        >
          {successMessage}
        </div>
      )}

      {loading && <p className="text-gov-slate">Loading appeals...</p>}
      {error && (
        <p className="text-gov-danger" role="alert">
          {error}
        </p>
      )}

      {!loading && !error && pendingAppeals.length === 0 ? (
        <Card>
          <p className="text-gov-slate">No appeals awaiting supervisor decision.</p>
        </Card>
      ) : (
        !loading &&
        !error && <AppealReviewPanel appeals={pendingAppeals} onDecided={handleDecided} />
      )}
    </div>
  );
}
