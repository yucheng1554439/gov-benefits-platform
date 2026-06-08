'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { ApiListResponse, Appeal } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { AppealForm } from '@/components/appeals/AppealForm';
import { Badge } from '@/components/ui/Badge';
import { formatDate } from '@/lib/utils';

export default function CitizenAppealPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const [appeals, setAppeals] = useState<Appeal[]>([]);

  const loadAppeals = () => {
    api
      .get<ApiListResponse<Appeal>>(`/cases/${id}/appeals`)
      .then((res) => setAppeals(res.data ?? []))
      .catch(() => setAppeals([]));
  };

  useEffect(() => {
    loadAppeals();
  }, [id]);

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <Link href={`/citizen/cases/${id}`} className="text-sm text-gov-navy underline">
          ← Back to Case
        </Link>
        <h1 className="mt-2 text-2xl font-bold text-gov-navy">File an Appeal</h1>
      </div>

      {appeals.length > 0 && (
        <Card title="Previous Appeals">
          <ul className="space-y-2">
            {appeals.map((a) => (
              <li key={a.id} className="flex items-center justify-between text-sm">
                <span>{a.grounds.slice(0, 60)}...</span>
                <div className="flex items-center gap-2">
                  <Badge variant="info">{a.status}</Badge>
                  <span className="text-gov-slate">{formatDate(a.filed_at)}</span>
                </div>
              </li>
            ))}
          </ul>
        </Card>
      )}

      <Card title="New Appeal">
        <AppealForm caseId={id} onSubmitted={loadAppeals} />
      </Card>
    </div>
  );
}
