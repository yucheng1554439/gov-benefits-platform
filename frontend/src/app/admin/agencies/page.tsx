'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { Agency, ApiListResponse } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';

export default function AdminAgenciesPage() {
  const [agencies, setAgencies] = useState<Agency[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .get<ApiListResponse<Agency>>('/agencies', { skipAuth: true })
      .then((res) => setAgencies(res.data ?? []))
      .finally(() => setLoading(false));
  }, []);

  const columns = [
    { key: 'code', header: 'Code', render: (row: Agency) => row.code },
    { key: 'name', header: 'Name', render: (row: Agency) => row.name },
    { key: 'type', header: 'Type', render: (row: Agency) => <Badge>{row.type}</Badge> },
    { key: 'jurisdiction', header: 'Jurisdiction', render: (row: Agency) => row.jurisdiction },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Agencies</h1>
        <p className="text-gov-slate">Participating government agencies</p>
      </div>
      <Card>
        {loading ? <p className="text-gov-slate">Loading...</p> : <Table columns={columns} data={agencies} keyField="id" />}
      </Card>
    </div>
  );
}
