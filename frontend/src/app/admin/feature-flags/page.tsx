'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { ApiListResponse, FeatureFlag } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Button } from '@/components/ui/Button';

export default function AdminFeatureFlagsPage() {
  const [flags, setFlags] = useState<FeatureFlag[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<string | null>(null);

  const load = () => {
    api
      .get<ApiListResponse<FeatureFlag>>('/feature-flags')
      .then((res) => setFlags(res.data ?? []))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    load();
  }, []);

  const toggle = async (flag: FeatureFlag) => {
    setSaving(flag.flag_key);
    try {
      await api.put('/feature-flags', {
        flag_key: flag.flag_key,
        is_enabled: !flag.is_enabled,
        rollout_pct: flag.rollout_pct ?? 100,
      });
      load();
    } finally {
      setSaving(null);
    }
  };

  const columns = [
    { key: 'key', header: 'Flag Key', render: (row: FeatureFlag) => <code className="text-sm">{row.flag_key}</code> },
    {
      key: 'enabled',
      header: 'Status',
      render: (row: FeatureFlag) => (
        <span className={row.is_enabled ? 'text-gov-success' : 'text-gov-slate'}>
          {row.is_enabled ? 'Enabled' : 'Disabled'}
        </span>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (row: FeatureFlag) => (
        <Button size="sm" variant="outline" onClick={() => toggle(row)} loading={saving === row.flag_key}>
          {row.is_enabled ? 'Disable' : 'Enable'}
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Feature Flags</h1>
        <p className="text-gov-slate">Toggle agency feature modules</p>
      </div>
      <Card>
        {loading ? <p className="text-gov-slate">Loading...</p> : <Table columns={columns} data={flags} keyField="flag_key" />}
      </Card>
    </div>
  );
}
