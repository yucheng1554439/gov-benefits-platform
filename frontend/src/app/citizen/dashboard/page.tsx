'use client';

import Link from 'next/link';
import { useCases } from '@/lib/hooks/useCases';
import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Badge, statusBadgeVariant } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { formatDate, formatStatus } from '@/lib/utils';
import type { Case } from '@/lib/api/types';

export default function CitizenDashboardPage() {
  const { cases, loading } = useCases();

  const columns = [
    {
      key: 'case_number',
      header: 'Case #',
      render: (row: Case) => (
        <Link href={`/citizen/cases/${row.id}`} className="font-medium text-gov-navy underline">
          {row.case_number}
        </Link>
      ),
    },
    {
      key: 'program',
      header: 'Program',
      render: (row: Case) => row.program?.name ?? '—',
    },
    {
      key: 'status',
      header: 'Status',
      render: (row: Case) => <Badge variant={statusBadgeVariant(row.status)}>{formatStatus(row.status)}</Badge>,
    },
    {
      key: 'submitted',
      header: 'Submitted',
      render: (row: Case) => formatDate(row.submitted_at),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gov-navy">My Dashboard</h1>
          <p className="text-gov-slate">Track your benefit applications and cases</p>
        </div>
        <Link href="/citizen/apply">
          <Button>New Application</Button>
        </Link>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <p className="text-sm text-gov-slate">Active Cases</p>
          <p className="text-3xl font-bold text-gov-navy">{cases.filter((c) => c.status !== 'closed').length}</p>
        </Card>
        <Card>
          <p className="text-sm text-gov-slate">Total Cases</p>
          <p className="text-3xl font-bold text-gov-navy">{cases.length}</p>
        </Card>
        <Card>
          <p className="text-sm text-gov-slate">Approved</p>
          <p className="text-3xl font-bold text-gov-success">
            {cases.filter((c) => c.status === 'approved').length}
          </p>
        </Card>
      </div>

      <Card title="My Cases">
        {loading ? (
          <p className="text-gov-slate">Loading...</p>
        ) : (
          <Table columns={columns} data={cases} keyField="id" />
        )}
      </Card>
    </div>
  );
}
