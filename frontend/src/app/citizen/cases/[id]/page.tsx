'use client';

import Link from 'next/link';
import { use } from 'react';
import { useCase } from '@/lib/hooks/useCases';
import { Card } from '@/components/ui/Card';
import { Badge, statusBadgeVariant } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { CaseTimeline } from '@/components/cases/CaseTimeline';
import { SLABadge } from '@/components/cases/SLABadge';
import { BenefitAmountCard } from '@/components/cases/BenefitAmountCard';
import { formatDate, formatStatus } from '@/lib/utils';

export default function CitizenCaseDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { caseData, loading, error } = useCase(id);

  if (loading) return <p className="text-gov-slate">Loading case...</p>;
  if (error || !caseData) return <p className="text-gov-danger">{error ?? 'Case not found'}</p>;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gov-navy">{caseData.case_number}</h1>
          <p className="text-gov-slate">{caseData.program?.name}</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant={statusBadgeVariant(caseData.status)}>{formatStatus(caseData.status)}</Badge>
          <SLABadge caseId={id} />
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <Card title="Case Details">
            <dl className="grid gap-3 sm:grid-cols-2">
              <div><dt className="text-xs text-gov-slate">Submitted</dt><dd>{formatDate(caseData.submitted_at)}</dd></div>
              <div><dt className="text-xs text-gov-slate">Priority</dt><dd className="capitalize">{caseData.priority}</dd></div>
              <div><dt className="text-xs text-gov-slate">ZIP Code</dt><dd>{caseData.zip_code ?? '—'}</dd></div>
              {caseData.application && (
                <>
                  <div><dt className="text-xs text-gov-slate">Household Size</dt><dd>{caseData.application.household_size}</dd></div>
                  <div><dt className="text-xs text-gov-slate">Annual Income</dt><dd>${caseData.application.annual_income.toLocaleString()}</dd></div>
                </>
              )}
            </dl>
          </Card>
          <Card title="Timeline">
            <CaseTimeline caseId={id} />
          </Card>
        </div>
        <div className="space-y-4">
          <BenefitAmountCard caseId={id} />
          <div className="flex flex-col gap-2">
            <Link href={`/citizen/cases/${id}/letters`}>
              <Button variant="outline" className="w-full">View Letters</Button>
            </Link>
            {caseData.status === 'denied' && (
              <Link href={`/citizen/cases/${id}/appeal`}>
                <Button className="w-full">File Appeal</Button>
              </Link>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
