'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api } from '@/lib/api/client';
import type { Appeal, ApiListResponse } from '@/lib/api/types';
import { isPendingAppeal } from '@/lib/appeals';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { useFeatureFlag } from '@/lib/feature-flags/useFeatureFlag';
import { formatStatus } from '@/lib/utils';

export default function WorkerAppealsPage() {
  const appealsEnabled = useFeatureFlag('appeals_module');
  const [appeals, setAppeals] = useState<Appeal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api
      .get<ApiListResponse<Appeal>>('/appeals?pending=true')
      .then((res) => setAppeals((res.data ?? []).filter(isPendingAppeal)))
      .catch((err) => setError(err instanceof Error ? err.message : 'Unable to load appeals.'))
      .finally(() => setLoading(false));
  }, []);

  if (!appealsEnabled) {
    return <p className="text-gov-slate">Appeals module is disabled.</p>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Appeals Queue</h1>
        <p className="text-gov-slate">
          Appeals awaiting supervisor decision ({appeals.length} pending). Supervisors use the Appeals
          menu to approve or deny.
        </p>
      </div>
      <Card>
        {loading && <p className="text-gov-slate">Loading appeals...</p>}
        {error && (
          <p className="text-gov-danger" role="alert">
            {error}
          </p>
        )}
        {!loading && !error && appeals.length === 0 ? (
          <p className="text-gov-slate">No pending appeals.</p>
        ) : (
          <ul className="divide-y divide-gov-border">
            {appeals.map((a) => (
              <li key={a.id} className="py-4">
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <div>
                    <Link
                      href={`/worker/cases/${a.case_id}`}
                      className="font-semibold text-gov-navy hover:underline"
                    >
                      {a.case_number || 'Case record'}
                    </Link>
                    {a.program_name && <p className="text-sm text-gov-slate">{a.program_name}</p>}
                    {a.citizen_name && <p className="text-sm text-gov-slate">{a.citizen_name}</p>}
                  </div>
                  <Badge variant="warning">{formatStatus(a.case_status || a.status)}</Badge>
                </div>
                <p className="mt-2 line-clamp-2 text-sm text-gov-slate">{a.grounds}</p>
              </li>
            ))}
          </ul>
        )}
      </Card>
    </div>
  );
}
