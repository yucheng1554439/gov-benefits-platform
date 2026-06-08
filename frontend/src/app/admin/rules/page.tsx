'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { formatDate } from '@/lib/utils';

interface EligibilityRule {
  id: string;
  name: string;
  program_name: string;
  program_code: string;
  is_active: boolean;
  version: number;
  effective_from: string;
  effective_to?: string;
  conditions?: Record<string, unknown>;
}

interface SimulateResult {
  is_eligible: boolean;
  rule_version: number;
  evaluation_trace?: Array<Record<string, unknown>>;
}

export default function AdminRulesPage() {
  const [rules, setRules] = useState<EligibilityRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [simulatingId, setSimulatingId] = useState<string | null>(null);
  const [simulateResult, setSimulateResult] = useState<SimulateResult | null>(null);

  useEffect(() => {
    api
      .get<{ data: EligibilityRule[] }>('/admin/eligibility-rules')
      .then((res) => setRules(res.data ?? []))
      .catch((err) => setError(err instanceof Error ? err.message : 'Unable to load rules.'))
      .finally(() => setLoading(false));
  }, []);

  const simulate = async (ruleId: string) => {
    setSimulatingId(ruleId);
    setSimulateResult(null);
    try {
      const result = await api.post<SimulateResult>(`/admin/eligibility-rules/${ruleId}/simulate`, {
        annual_income: 28000,
        household_size: 3,
      });
      setSimulateResult(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Simulation failed.');
    } finally {
      setSimulatingId(null);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Eligibility Rules</h1>
        <p className="text-gov-slate">Active program eligibility criteria and versions</p>
      </div>

      {loading && <p className="text-gov-slate">Loading rules...</p>}
      {error && (
        <p className="text-gov-danger" role="alert">
          {error}
        </p>
      )}

      <div className="grid gap-4">
        {rules.map((rule) => (
          <Card key={rule.id} title={rule.name}>
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant="info">{rule.program_name}</Badge>
              {rule.is_active && <Badge variant="success">Active</Badge>}
              <Badge variant="default">Version {rule.version || 1}</Badge>
            </div>
            <dl className="mt-3 grid gap-1 text-sm text-gov-slate sm:grid-cols-2">
              <div>
                <dt className="font-medium text-gov-navy">Effective from</dt>
                <dd>{rule.effective_from ? formatDate(rule.effective_from) : '—'}</dd>
              </div>
              <div>
                <dt className="font-medium text-gov-navy">Effective to</dt>
                <dd>{rule.effective_to ? formatDate(rule.effective_to) : 'Open-ended'}</dd>
              </div>
              <div>
                <dt className="font-medium text-gov-navy">Last modified</dt>
                <dd>{rule.effective_from ? formatDate(rule.effective_from) : '—'}</dd>
              </div>
              <div>
                <dt className="font-medium text-gov-navy">Program code</dt>
                <dd>{rule.program_code}</dd>
              </div>
            </dl>
            <div className="mt-4 flex flex-wrap gap-2">
              <Button size="sm" variant="secondary" disabled title="Rule editing is view-only in this release">
                Edit
              </Button>
              <Button
                size="sm"
                onClick={() => simulate(rule.id)}
                loading={simulatingId === rule.id}
              >
                Simulate
              </Button>
            </div>
          </Card>
        ))}
      </div>

      {simulateResult && (
        <Card title="Simulation Result">
          <p className="text-sm text-gov-slate">
            Rule version {simulateResult.rule_version} — sample household (3 members, $28,000 income)
          </p>
          <p className="mt-2 font-medium text-gov-navy">
            Result: {simulateResult.is_eligible ? 'Eligible' : 'Not eligible'}
          </p>
        </Card>
      )}
    </div>
  );
}
