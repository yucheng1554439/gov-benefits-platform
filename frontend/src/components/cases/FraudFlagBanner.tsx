'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { ApiListResponse, FraudFlag } from '@/lib/api/types';
import { Badge } from '@/components/ui/Badge';

interface FraudFlagBannerProps {
  caseId: string;
  refreshKey?: number;
}

export function FraudFlagBanner({ caseId, refreshKey = 0 }: FraudFlagBannerProps) {
  const [flags, setFlags] = useState<FraudFlag[]>([]);

  useEffect(() => {
    api
      .get<ApiListResponse<FraudFlag>>(`/cases/${caseId}/fraud`)
      .then((res) => setFlags((res.data ?? []).filter((f) => f.status === 'open')))
      .catch(() => setFlags([]));
  }, [caseId, refreshKey]);

  if (flags.length === 0) return null;

  return (
    <div
      className="mb-4 rounded-lg border border-amber-300 bg-amber-50 px-4 py-3"
      role="alert"
      aria-live="polite"
    >
      <p className="font-semibold text-gov-warning">Fraud flags detected</p>
      <ul className="mt-2 space-y-1">
        {flags.map((flag) => (
          <li key={flag.id} className="flex items-center gap-2 text-sm text-gov-slate">
            <Badge variant={flag.severity === 'high' ? 'danger' : 'warning'}>
              {flag.severity}
            </Badge>
            <span>{flag.flag_type.replace(/_/g, ' ')}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
