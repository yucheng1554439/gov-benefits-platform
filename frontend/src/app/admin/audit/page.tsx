'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { api } from '@/lib/api/client';
import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { formatDate, formatStatus } from '@/lib/utils';

interface AuditLog {
  id: string;
  action: string;
  entity_type: string;
  entity_id?: string;
  actor_name?: string;
  new_state?: Record<string, unknown>;
  created_at: string;
}

interface AuditResponse {
  data: AuditLog[];
  total: number;
  offset: number;
  limit: number;
}

const PAGE_SIZE = 25;

const ACTION_OPTIONS = [
  '',
  'case.created',
  'case.status_changed',
  'application.created',
  'eligibility.evaluated',
  'benefit.calculated',
  'appeal.filed',
  'appeal.decided',
  'letter.generated',
  'document.uploaded',
];

function actionBadgeVariant(action: string): 'info' | 'success' | 'warning' | 'danger' | 'default' {
  if (action.includes('denied') || action.includes('fraud')) return 'danger';
  if (action.includes('approved') || action.includes('created')) return 'success';
  if (action.includes('appeal') || action.includes('status')) return 'warning';
  return 'info';
}

export default function AuditLogPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [search, setSearch] = useState('');
  const [actionFilter, setActionFilter] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const loadLogs = useCallback(async () => {
    setLoading(true);
    setError('');
    const params = new URLSearchParams({
      limit: String(PAGE_SIZE),
      offset: String(offset),
    });
    if (search.trim()) params.set('search', search.trim());
    if (actionFilter) params.set('action', actionFilter);

    try {
      const res = await api.get<AuditResponse>(`/audit-logs?${params.toString()}`);
      setLogs(res.data ?? []);
      setTotal(res.total ?? 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load audit logs.');
      setLogs([]);
    } finally {
      setLoading(false);
    }
  }, [offset, search, actionFilter]);

  useEffect(() => {
    loadLogs();
  }, [loadLogs]);

  const page = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const columns = useMemo(
    () => [
      {
        key: 'action',
        header: 'Event',
        render: (row: AuditLog) => (
          <Badge variant={actionBadgeVariant(row.action)}>{row.action.replace(/\./g, ' ')}</Badge>
        ),
      },
      {
        key: 'actor',
        header: 'Actor',
        render: (row: AuditLog) => row.actor_name || 'System',
      },
      {
        key: 'details',
        header: 'Details',
        render: (row: AuditLog) => {
          if (row.new_state?.to_status) {
            const from = row.new_state.from_status ? formatStatus(String(row.new_state.from_status)) : '—';
            return `${from} → ${formatStatus(String(row.new_state.to_status))}`;
          }
          if (row.new_state?.is_eligible != null) {
            return row.new_state.is_eligible ? 'Eligible' : 'Not eligible';
          }
          if (row.new_state?.amount != null) {
            return `Amount: $${row.new_state.amount}`;
          }
          return row.entity_id?.slice(0, 8) ?? '—';
        },
      },
      { key: 'created_at', header: 'When', render: (row: AuditLog) => formatDate(row.created_at) },
    ],
    [],
  );

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Audit Trail</h1>
        <p className="text-gov-slate">Immutable record of platform events for this agency</p>
      </div>

      <Card>
        <div className="mb-4 flex flex-wrap gap-3">
          <input
            type="search"
            placeholder="Search actions, actors, details..."
            value={search}
            onChange={(e) => {
              setOffset(0);
              setSearch(e.target.value);
            }}
            className="min-w-[220px] flex-1 rounded border border-gov-border px-3 py-2 text-sm"
            aria-label="Search audit logs"
          />
          <select
            value={actionFilter}
            onChange={(e) => {
              setOffset(0);
              setActionFilter(e.target.value);
            }}
            className="rounded border border-gov-border px-3 py-2 text-sm"
            aria-label="Filter by event type"
          >
            <option value="">All events</option>
            {ACTION_OPTIONS.filter(Boolean).map((action) => (
              <option key={action} value={action}>
                {action}
              </option>
            ))}
          </select>
          <Button size="sm" variant="secondary" onClick={loadLogs}>
            Refresh
          </Button>
        </div>

        {loading && <p className="text-gov-slate">Loading audit logs...</p>}
        {error && (
          <p className="text-gov-danger" role="alert">
            {error}
          </p>
        )}
        {!loading && !error && (
          <>
            <Table columns={columns} data={logs} keyField="id" />
            <div className="mt-4 flex items-center justify-between text-sm text-gov-slate">
              <span>
                Page {page} of {totalPages} ({total} events)
              </span>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="secondary"
                  disabled={offset === 0}
                  onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                >
                  Previous
                </Button>
                <Button
                  size="sm"
                  variant="secondary"
                  disabled={offset + PAGE_SIZE >= total}
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                >
                  Next
                </Button>
              </div>
            </div>
          </>
        )}
      </Card>
    </div>
  );
}
