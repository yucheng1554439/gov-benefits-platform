'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { CaseSLATracking } from '@/lib/api/types';
import { Badge } from '@/components/ui/Badge';
import { formatDate } from '@/lib/utils';

interface SLABadgeProps {
  caseId: string;
}

function slaVariant(status: string): 'default' | 'success' | 'warning' | 'danger' | 'info' {
  if (status === 'breached') return 'danger';
  if (status === 'warning') return 'warning';
  if (status === 'on_track') return 'success';
  return 'default';
}

export function SLABadge({ caseId }: SLABadgeProps) {
  const [sla, setSla] = useState<CaseSLATracking | null>(null);

  useEffect(() => {
    api
      .get<CaseSLATracking>(`/cases/${caseId}/sla`)
      .then(setSla)
      .catch(() => setSla(null));
  }, [caseId]);

  if (!sla) return null;

  return (
    <div className="inline-flex items-center gap-2">
      <Badge variant={slaVariant(sla.status)}>SLA: {sla.status.replace(/_/g, ' ')}</Badge>
      <span className="text-xs text-gov-slate">
        Due {formatDate(sla.due_at)} · {sla.elapsed_days} days elapsed
      </span>
    </div>
  );
}
