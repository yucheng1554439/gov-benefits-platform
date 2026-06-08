'use client';

import Link from 'next/link';
import { useEffect, useMemo, useState } from 'react';
import { api } from '@/lib/api/client';
import { useCases } from '@/lib/hooks/useCases';
import type { CaseSLATracking } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Badge, statusBadgeVariant } from '@/components/ui/Badge';
import { formatDate, formatStatus } from '@/lib/utils';
import type { Case } from '@/lib/api/types';

const ESCALATED_STATUSES = ['supervisor_review', 'appealed', 'appeal_review'];

export default function EscalationsPage() {
  const { cases, loading } = useCases();
  const [breached, setBreached] = useState<CaseSLATracking[]>([]);

  useEffect(() => {
    api
      .get<{ data: CaseSLATracking[] }>('/sla/breached')
      .then((res) => setBreached(res.data ?? []))
      .catch(() => setBreached([]));
  }, []);

  const escalated = useMemo(
    () => cases.filter((c) => ESCALATED_STATUSES.includes(c.status)),
    [cases],
  );

  const supervisorReviewCount = cases.filter((c) => c.status === 'supervisor_review').length;
  const appealCount = cases.filter((c) => c.status === 'appealed' || c.status === 'appeal_review').length;

  const columns = [
    {
      key: 'case_number',
      header: 'Case #',
      render: (row: Case) => (
        <Link href={`/supervisor/cases/${row.id}`} className="font-medium text-gov-navy underline">
          {row.case_number}
        </Link>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (row: Case) => (
        <Badge variant={statusBadgeVariant(row.status)}>{formatStatus(row.status)}</Badge>
      ),
    },
    {
      key: 'action',
      header: 'Action',
      render: (row: Case) =>
        row.status === 'appealed' || row.status === 'appeal_review' ? (
          <Link href="/supervisor/appeals" className="text-sm font-medium text-gov-navy underline">
            Review appeal
          </Link>
        ) : (
          <span className="text-sm text-gov-slate">Review case</span>
        ),
    },
    { key: 'priority', header: 'Priority', render: (row: Case) => <span className="capitalize">{row.priority}</span> },
    { key: 'submitted', header: 'Submitted', render: (row: Case) => formatDate(row.submitted_at) },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Escalations</h1>
        <p className="text-gov-slate">Cases requiring supervisor attention</p>
      </div>
      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <p className="text-sm text-gov-slate">Supervisor Review</p>
          <p className="text-3xl font-bold text-gov-navy">{supervisorReviewCount}</p>
        </Card>
        <Card>
          <p className="text-sm text-gov-slate">Appeals Pending</p>
          <p className="text-3xl font-bold text-gov-navy">{appealCount}</p>
          {appealCount > 0 && (
            <Link href="/supervisor/appeals" className="mt-2 inline-block text-sm text-gov-navy underline">
              Open appeal review
            </Link>
          )}
        </Card>
        <Card>
          <p className="text-sm text-gov-slate">SLA Breached</p>
          <p className="text-3xl font-bold text-gov-danger">{breached.length}</p>
        </Card>
      </div>
      <Card title="Escalated Cases">
        {loading ? <p className="text-gov-slate">Loading...</p> : <Table columns={columns} data={escalated} keyField="id" />}
      </Card>
    </div>
  );
}
