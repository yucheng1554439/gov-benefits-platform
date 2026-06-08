'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { CaseSLATracking } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { formatDate } from '@/lib/utils';

export default function AdminSLAPage() {
  const [breached, setBreached] = useState<CaseSLATracking[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .get<{ data: CaseSLATracking[] }>('/sla/breached')
      .then((res) => setBreached(res.data ?? []))
      .catch(() => setBreached([]))
      .finally(() => setLoading(false));
  }, []);

  const columns = [
    { key: 'case', header: 'Case ID', render: (row: CaseSLATracking) => row.case_id.slice(0, 8) + '...' },
    { key: 'status', header: 'SLA Status', render: (row: CaseSLATracking) => <Badge variant="danger">{row.status}</Badge> },
    { key: 'elapsed', header: 'Elapsed Days', render: (row: CaseSLATracking) => row.elapsed_days },
    { key: 'due', header: 'Due Date', render: (row: CaseSLATracking) => formatDate(row.due_at) },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">SLA Management</h1>
        <p className="text-gov-slate">Monitor service level agreement compliance</p>
      </div>
      <Card title="Breached SLAs">
        {loading ? <p className="text-gov-slate">Loading...</p> : <Table columns={columns} data={breached} keyField="id" emptyMessage="No breached SLAs." />}
      </Card>
    </div>
  );
}
