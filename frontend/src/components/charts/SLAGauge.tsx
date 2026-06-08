'use client';

import { PieChart, Pie, Cell, ResponsiveContainer } from 'recharts';
import { Card } from '@/components/ui/Card';

interface SLAGaugeProps {
  onTrack: number;
  warning: number;
  breached: number;
}

export function SLAGauge({ onTrack, warning, breached }: SLAGaugeProps) {
  const total = onTrack + warning + breached || 1;
  const compliance = Math.round((onTrack / total) * 100);

  const data = [
    { name: 'On Track', value: onTrack, color: '#166534' },
    { name: 'Warning', value: warning, color: '#b45309' },
    { name: 'Breached', value: breached, color: '#b91c1c' },
  ];

  return (
    <Card title="SLA Compliance">
      <div className="relative h-48" role="img" aria-label={`SLA compliance ${compliance} percent`}>
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={data}
              cx="50%"
              cy="70%"
              startAngle={180}
              endAngle={0}
              innerRadius={60}
              outerRadius={90}
              dataKey="value"
            >
              {data.map((entry) => (
                <Cell key={entry.name} fill={entry.color} />
              ))}
            </Pie>
          </PieChart>
        </ResponsiveContainer>
        <div className="absolute inset-0 flex flex-col items-center justify-end pb-4">
          <span className="text-3xl font-bold text-gov-navy">{compliance}%</span>
          <span className="text-sm text-gov-slate">On Track</span>
        </div>
      </div>
      <div className="mt-2 flex justify-center gap-4 text-xs text-gov-slate">
        {data.map((d) => (
          <span key={d.name} className="flex items-center gap-1">
            <span className="inline-block h-2 w-2 rounded-full" style={{ background: d.color }} />
            {d.name}: {d.value}
          </span>
        ))}
      </div>
    </Card>
  );
}
