'use client';

import dynamic from 'next/dynamic';
import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { AnalyticsSummary } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { useFeatureFlag } from '@/lib/feature-flags/useFeatureFlag';

const GeoHeatmap = dynamic(() => import('@/components/maps/GeoHeatmap').then((m) => m.GeoHeatmap), {
  ssr: false,
  loading: () => <div className="flex h-[500px] items-center justify-center text-gov-slate">Loading map...</div>,
});

export default function GeographyPage() {
  const geoEnabled = useFeatureFlag('geo_analytics');
  const [zipData, setZipData] = useState<Record<string, number>>({});

  useEffect(() => {
    api
      .get<AnalyticsSummary>('/analytics/summary')
      .then((s) => setZipData(s.cases_by_zip ?? {}))
      .catch(() => setZipData({}));
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Geographic Analytics</h1>
        <p className="text-gov-slate">Case distribution heatmap by ZIP code</p>
      </div>
      {!geoEnabled ? (
        <Card><p className="text-gov-slate">Geographic analytics is disabled for this agency.</p></Card>
      ) : (
        <Card title="Los Angeles County Case Heatmap">
          <GeoHeatmap data={zipData} />
        </Card>
      )}
    </div>
  );
}
