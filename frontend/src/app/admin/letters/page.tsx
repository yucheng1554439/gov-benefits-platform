'use client';

import { Card } from '@/components/ui/Card';

const TEMPLATES = [
  { type: 'approval_notice', name: 'Approval Notice', fields: ['CitizenName', 'ProgramName', 'BenefitAmount', 'CaseNumber'] },
  { type: 'denial_notice', name: 'Denial Notice', fields: ['CitizenName', 'ProgramName', 'DenialReason', 'CaseNumber'] },
];

export default function AdminLettersPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Letter Templates</h1>
        <p className="text-gov-slate">Manage correspondence templates</p>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        {TEMPLATES.map((t) => (
          <Card key={t.type} title={t.name}>
            <p className="text-sm text-gov-slate">Type: {t.type}</p>
            <p className="mt-2 text-xs text-gov-slate">Merge fields: {t.fields.join(', ')}</p>
          </Card>
        ))}
      </div>
    </div>
  );
}
