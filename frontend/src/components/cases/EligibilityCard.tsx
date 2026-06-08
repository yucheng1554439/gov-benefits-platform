'use client';

import { useEffect, useState } from 'react';
import { ApiClientError, api } from '@/lib/api/client';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';

interface EligibilityResult {
  is_eligible?: boolean;
  evaluation_trace?: Array<{
    field: string;
    op: string;
    pass: boolean;
    actual?: number;
    value?: number;
  }>;
  evaluated_at?: string;
}

interface EligibilityCardProps {
  caseId: string;
  canEvaluate?: boolean;
}

export function EligibilityCard({ caseId, canEvaluate }: EligibilityCardProps) {
  const [result, setResult] = useState<EligibilityResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const load = () => {
    api
      .get<EligibilityResult>(`/cases/${caseId}/eligibility`)
      .then((data) => setResult(data))
      .catch(() => setResult(null));
  };

  useEffect(() => {
    load();
  }, [caseId]);

  const evaluate = async () => {
    setLoading(true);
    setMessage('');
    setError('');
    try {
      const data = await api.post<EligibilityResult>(`/cases/${caseId}/eligibility/evaluate`);
      setResult(data);
      setMessage(data.is_eligible ? 'Applicant is eligible.' : 'Applicant is not eligible.');
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : 'Unable to evaluate eligibility.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card title="Eligibility">
      {result?.is_eligible != null ? (
        <div className="space-y-3">
          <Badge variant={result.is_eligible ? 'success' : 'danger'}>
            {result.is_eligible ? 'Eligible' : 'Not Eligible'}
          </Badge>
          {result.evaluation_trace && result.evaluation_trace.length > 0 && (
            <ul className="space-y-1 text-sm text-gov-slate">
              {result.evaluation_trace.map((step) => (
                <li key={`${step.field}-${step.op}`}>
                  {step.field} {step.op} {step.value ?? ''} — {step.pass ? 'pass' : 'fail'}
                </li>
              ))}
            </ul>
          )}
        </div>
      ) : (
        <p className="text-sm text-gov-slate">No eligibility evaluation on file.</p>
      )}
      {message && <p className="mt-3 text-sm text-green-700">{message}</p>}
      {error && <p className="mt-3 text-sm text-gov-danger">{error}</p>}
      {canEvaluate && (
        <Button className="mt-4" size="sm" onClick={evaluate} loading={loading}>
          Evaluate Eligibility
        </Button>
      )}
    </Card>
  );
}
