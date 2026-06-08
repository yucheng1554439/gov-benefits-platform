'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api } from '@/lib/api/client';
import { useCases } from '@/lib/hooks/useCases';
import type { ApiListResponse, FraudFlag } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Textarea } from '@/components/ui/Input';
import { useFeatureFlag } from '@/lib/feature-flags/useFeatureFlag';

export default function FraudReviewPage() {
  const fraudEnabled = useFeatureFlag('fraud_detection');
  const { cases } = useCases();
  const [flags, setFlags] = useState<(FraudFlag & { case_number?: string })[]>([]);
  const [notes, setNotes] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState<Record<string, boolean>>({});

  useEffect(() => {
    async function loadFlags() {
      const all: (FraudFlag & { case_number?: string })[] = [];
      for (const c of cases) {
        try {
          const res = await api.get<ApiListResponse<FraudFlag>>(`/cases/${c.id}/fraud`);
          (res.data ?? [])
            .filter((f) => f.status === 'open')
            .forEach((f) => all.push({ ...f, case_number: c.case_number }));
        } catch {
          // skip
        }
      }
      setFlags(all);
    }
    if (cases.length) loadFlags();
  }, [cases]);

  const review = async (flagId: string, disposition: string) => {
    setLoading((l) => ({ ...l, [flagId]: true }));
    try {
      await api.post(`/fraud/${flagId}/review`, { disposition, notes: notes[flagId] ?? '' });
      setFlags((prev) => prev.filter((f) => f.id !== flagId));
    } finally {
      setLoading((l) => ({ ...l, [flagId]: false }));
    }
  };

  if (!fraudEnabled) {
    return <p className="text-gov-slate">Fraud detection module is disabled for this agency.</p>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Fraud Review</h1>
        <p className="text-gov-slate">Review open fraud flags across cases</p>
      </div>
      {flags.length === 0 ? (
        <Card><p className="text-gov-slate">No open fraud flags.</p></Card>
      ) : (
        flags.map((flag) => (
          <Card key={flag.id} title={flag.flag_type.replace(/_/g, ' ')}>
            <div className="mb-3 flex items-center gap-2">
              <Badge variant={flag.severity === 'high' ? 'danger' : 'warning'}>{flag.severity}</Badge>
              <Link href={`/worker/cases/${flag.case_id}`} className="text-sm text-gov-navy underline">
                {flag.case_number}
              </Link>
            </div>
            <Textarea
              label="Review Notes"
              value={notes[flag.id] ?? ''}
              onChange={(e) => setNotes((n) => ({ ...n, [flag.id]: e.target.value }))}
              rows={2}
            />
            <div className="mt-3 flex gap-2">
              <Button size="sm" onClick={() => review(flag.id, 'confirmed')} loading={loading[flag.id]}>Confirm</Button>
              <Button size="sm" variant="outline" onClick={() => review(flag.id, 'dismissed')} loading={loading[flag.id]}>Dismiss</Button>
            </div>
          </Card>
        ))
      )}
    </div>
  );
}
