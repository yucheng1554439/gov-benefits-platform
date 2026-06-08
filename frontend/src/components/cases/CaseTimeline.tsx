'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { WorkflowEvent } from '@/lib/api/types';
import { formatDate, formatStatus } from '@/lib/utils';
import { Badge } from '@/components/ui/Badge';

interface CaseTimelineProps {
  caseId: string;
  refreshKey?: number;
}

export function CaseTimeline({ caseId, refreshKey = 0 }: CaseTimelineProps) {
  const [events, setEvents] = useState<WorkflowEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState('');

  useEffect(() => {
    setLoading(true);
    setLoadError('');
    api
      .get<{ data: WorkflowEvent[] }>(`/cases/${caseId}/workflow`)
      .then((res) => setEvents(res.data ?? []))
      .catch((err) => {
        setEvents([]);
        setLoadError(err instanceof Error ? err.message : 'Unable to load timeline.');
      })
      .finally(() => setLoading(false));
  }, [caseId, refreshKey]);

  if (loading) {
    return <p className="text-sm text-gov-slate">Loading timeline...</p>;
  }

  if (loadError) {
    return (
      <p className="text-sm text-gov-danger" role="alert">
        {loadError}
      </p>
    );
  }

  if (events.length === 0) {
    return <p className="text-sm text-gov-slate">No workflow events recorded.</p>;
  }

  return (
    <ol className="relative border-l-2 border-gov-border pl-6" aria-label="Case timeline">
      {events.map((event) => (
        <li key={event.id} className="mb-6 last:mb-0">
          <span className="absolute -left-[9px] mt-1.5 h-4 w-4 rounded-full border-2 border-gov-navy bg-white" />
          <div className="flex flex-wrap items-center gap-2">
            {event.from_status ? (
              <span className="text-sm font-medium text-gov-navy">
                {formatStatus(event.from_status)} → {formatStatus(event.to_status)}
              </span>
            ) : (
              <Badge variant="info">{formatStatus(event.to_status)}</Badge>
            )}
            <time className="text-xs text-gov-slate" dateTime={event.created_at}>
              {formatDate(event.created_at)}
            </time>
          </div>
          {event.actor_name && (
            <p className="mt-1 text-sm text-gov-slate">Actor: {event.actor_name}</p>
          )}
          {event.reason && (
            <p className="mt-1 text-sm text-gov-slate">Reason: {event.reason}</p>
          )}
        </li>
      ))}
    </ol>
  );
}
