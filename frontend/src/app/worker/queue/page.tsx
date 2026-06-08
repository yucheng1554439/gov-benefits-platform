'use client';

import Link from 'next/link';
import { useCases } from '@/lib/hooks/useCases';
import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Badge, statusBadgeVariant } from '@/components/ui/Badge';
import { formatDate, formatStatus } from '@/lib/utils';
import type { Case } from '@/lib/api/types';

export default function WorkerQueuePage() {
  const { cases, loading } = useCases();

  const activeCases = cases.filter((c) => !['closed', 'approved'].includes(c.status));

  const columns = [
    {
      key: 'case_number',
      header: 'Case #',
      render: (row: Case) => (
        <Link href={`/worker/cases/${row.id}`} className="font-medium text-gov-navy underline">
          {row.case_number}
        </Link>
      ),
    },
    { key: 'program', header: 'Program', render: (row: Case) => row.program?.name ?? '—' },
    {
      key: 'status',
      header: 'Status',
      render: (row: Case) => <Badge variant={statusBadgeVariant(row.status)}>{formatStatus(row.status)}</Badge>,
    },
    { key: 'priority', header: 'Priority', render: (row: Case) => <span className="capitalize">{row.priority}</span> },
    { key: 'submitted', header: 'Submitted', render: (row: Case) => formatDate(row.submitted_at) },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Case Queue</h1>
        <p className="text-gov-slate">Cases assigned to your workload</p>
      </div>
      <div className="grid gap-4 sm:grid-cols-3">
        <Card><p className="text-sm text-gov-slate">In Queue</p><p className="text-3xl font-bold">{activeCases.length}</p></Card>
        <Card><p className="text-sm text-gov-slate">High Priority</p><p className="text-3xl font-bold text-gov-warning">{activeCases.filter((c) => c.priority === 'high' || c.priority === 'urgent').length}</p></Card>
        <Card><p className="text-sm text-gov-slate">Total Cases</p><p className="text-3xl font-bold">{cases.length}</p></Card>
      </div>
      <Card title="Active Cases">
        {loading ? <p className="text-gov-slate">Loading...</p> : <Table columns={columns} data={activeCases} keyField="id" />}
      </Card>
    </div>
  );
}
