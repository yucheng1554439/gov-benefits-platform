'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { formatDate } from '@/lib/utils';

interface BenefitResult {
  calculated_amount?: number;
  calculated_at?: string;
  rule_version?: number;
  calculation_trace?: Array<{ step: string; value: number }>;
}

interface BenefitAmountCardProps {
  caseId: string;
  canCalculate?: boolean;
}

export function BenefitAmountCard({ caseId, canCalculate }: BenefitAmountCardProps) {
  const [benefit, setBenefit] = useState<BenefitResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const loadBenefit = () => {
    api
      .get<BenefitResult>(`/cases/${caseId}/benefit`)
      .then(setBenefit)
      .catch(() => setBenefit(null));
  };

  useEffect(() => {
    loadBenefit();
  }, [caseId]);

  const calculate = async () => {
    setLoading(true);
    setMessage('');
    setError('');
    try {
      const result = await api.post<BenefitResult>(`/cases/${caseId}/benefit/calculate`);
      setBenefit(result);
      setMessage(result.calculated_amount != null && benefit?.calculated_amount != null ? 'Benefit amount recalculated.' : 'Benefit amount calculated.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to calculate benefit.');
    } finally {
      setLoading(false);
    }
  };

  const hasCalculation = benefit?.calculated_amount != null;

  return (
    <Card title="Benefit Amount">
      {hasCalculation ? (
        <div>
          <p className="text-3xl font-bold text-gov-navy">
            ${benefit.calculated_amount!.toLocaleString()}
            <span className="text-base font-normal text-gov-slate">/month</span>
          </p>
          {benefit.calculated_at && (
            <p className="mt-2 text-sm text-gov-slate">Calculated: {formatDate(benefit.calculated_at)}</p>
          )}
          {benefit.rule_version != null && (
            <p className="text-sm text-gov-slate">Rule Version: {benefit.rule_version}.0</p>
          )}
        </div>
      ) : (
        <p className="text-sm text-gov-slate">No benefit calculation on file.</p>
      )}
      {message && <p className="mt-3 text-sm text-green-700">{message}</p>}
      {error && (
        <p className="mt-3 text-sm text-gov-danger" role="alert">
          {error}
        </p>
      )}
      {canCalculate && (
        <Button className="mt-4" size="sm" onClick={calculate} loading={loading}>
          {hasCalculation ? 'Recalculate' : 'Calculate Benefit'}
        </Button>
      )}
    </Card>
  );
}
