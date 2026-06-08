'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { ApiListResponse, Report } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { formatDate } from '@/lib/utils';

export default function AdminReportsPage() {
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);

  const load = () => {
    api
      .get<ApiListResponse<Report>>('/reports')
      .then((res) => setReports(res.data ?? []))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    load();
  }, []);

  const createReport = async (type: string) => {
    setCreating(true);
    try {
      await api.post('/reports', { report_type: type, params: {} });
      load();
    } finally {
      setCreating(false);
    }
  };

  const columns = [
    { key: 'type', header: 'Type', render: (row: Report) => row.report_type },
    { key: 'status', header: 'Status', render: (row: Report) => <Badge>{row.status}</Badge> },
    { key: 'created', header: 'Requested', render: (row: Report) => formatDate(row.created_at) },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gov-navy">Reports</h1>
          <p className="text-gov-slate">Generate and track agency reports</p>
        </div>
        <div className="flex gap-2">
          <Button size="sm" onClick={() => createReport('case_summary')} loading={creating}>Case Summary</Button>
          <Button size="sm" variant="outline" onClick={() => createReport('sla_compliance')} loading={creating}>SLA Report</Button>
        </div>
      </div>
      <Card>
        {loading ? <p className="text-gov-slate">Loading...</p> : <Table columns={columns} data={reports} keyField="id" />}
      </Card>
    </div>
  );
}
