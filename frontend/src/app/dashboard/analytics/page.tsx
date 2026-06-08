'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { AnalyticsSummary } from '@/lib/api/types';
import { KPIChart } from '@/components/charts/KPIChart';
import { StatusPieChart } from '@/components/charts/StatusPieChart';
import { SLAGauge } from '@/components/charts/SLAGauge';
import { Card } from '@/components/ui/Card';

export default function AnalyticsPage() {
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .get<AnalyticsSummary>('/analytics/summary')
      .then(setSummary)
      .catch(() => setSummary(null))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p className="text-gov-slate">Loading analytics...</p>;

  const statusCounts = summary?.case_status_counts ?? {};
  const totalCases = Object.values(statusCounts).reduce((a, b) => a + b, 0);
  const onTrack = statusCounts['under_review'] ?? 0;
  const warning = statusCounts['need_documents'] ?? 0;
  const breached = statusCounts['denied'] ?? 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Analytics Dashboard</h1>
        <p className="text-gov-slate">Agency performance metrics and KPIs</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-4">
        <Card><p className="text-sm text-gov-slate">Total Cases</p><p className="text-3xl font-bold">{totalCases}</p></Card>
        <Card><p className="text-sm text-gov-slate">Open Fraud Flags</p><p className="text-3xl font-bold text-gov-warning">{summary?.open_fraud_flags ?? 0}</p></Card>
        <Card><p className="text-sm text-gov-slate">Under Review</p><p className="text-3xl font-bold">{onTrack}</p></Card>
        <Card><p className="text-sm text-gov-slate">ZIP Regions</p><p className="text-3xl font-bold">{Object.keys(summary?.cases_by_zip ?? {}).length}</p></Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <StatusPieChart title="Cases by Status" data={statusCounts} />
        <SLAGauge onTrack={onTrack} warning={warning} breached={breached} />
      </div>

      {summary?.cases_by_zip && (
        <KPIChart title="Cases by ZIP Code" data={summary.cases_by_zip} color="#c9a227" />
      )}
    </div>
  );
}
