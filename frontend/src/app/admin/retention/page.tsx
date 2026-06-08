'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { ApiListResponse, RetentionPolicy } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { useFeatureFlag } from '@/lib/feature-flags/useFeatureFlag';

export default function AdminRetentionPage() {
  const enabled = useFeatureFlag('retention_policies');
  const [policies, setPolicies] = useState<RetentionPolicy[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .get<ApiListResponse<RetentionPolicy>>('/admin/retention-policies')
      .then((res) => setPolicies(res.data ?? []))
      .catch(() => setPolicies([]))
      .finally(() => setLoading(false));
  }, []);

  const columns = [
    { key: 'entity', header: 'Entity Type', render: (row: RetentionPolicy) => row.entity_type.replace(/_/g, ' ') },
    { key: 'years', header: 'Retention', render: (row: RetentionPolicy) => `${row.retention_years} years` },
    { key: 'action', header: 'Disposition', render: (row: RetentionPolicy) => <Badge>{row.disposition_action}</Badge> },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Retention Policies</h1>
        <p className="text-gov-slate">
          Data retention and disposition rules {!enabled && '(feature disabled)'}
        </p>
      </div>
      <Card>
        {loading ? <p className="text-gov-slate">Loading...</p> : <Table columns={columns} data={policies} keyField="id" />}
      </Card>
    </div>
  );
}
