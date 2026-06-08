'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { ApiListResponse, Letter } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { formatDate } from '@/lib/utils';

interface LetterActionsCardProps {
  caseId: string;
  caseStatus: string;
}

export function LetterActionsCard({ caseId, caseStatus }: LetterActionsCardProps) {
  const [letters, setLetters] = useState<Letter[]>([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const canGenerate = ['approved', 'denied', 'appeal_approved', 'appeal_denied'].includes(caseStatus);

  const loadLetters = () => {
    api
      .get<ApiListResponse<Letter>>(`/cases/${caseId}/letters`)
      .then((res) => setLetters(res.data ?? []))
      .catch(() => setLetters([]));
  };

  useEffect(() => {
    loadLetters();
  }, [caseId]);

  const generate = async (letterType: 'approval_notice' | 'denial_notice') => {
    setLoading(true);
    setMessage('');
    setError('');
    try {
      await api.post(`/cases/${caseId}/letters`, { letter_type: letterType });
      setMessage('Letter generated successfully.');
      loadLetters();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to generate letter.');
    } finally {
      setLoading(false);
    }
  };

  if (!canGenerate && letters.length === 0) {
    return null;
  }

  return (
    <Card title="Correspondence">
      {letters.length > 0 && (
        <ul className="mb-4 space-y-2 text-sm text-gov-slate">
          {letters.map((letter) => (
            <li key={letter.id}>
              {letter.letter_type.replace(/_/g, ' ')} — {formatDate(letter.generated_at)}
            </li>
          ))}
        </ul>
      )}
      {canGenerate && (
        <div className="flex flex-wrap gap-2">
          {(caseStatus === 'approved' || caseStatus === 'appeal_approved') && (
            <Button size="sm" loading={loading} onClick={() => generate('approval_notice')}>
              Generate Approval Letter
            </Button>
          )}
          {(caseStatus === 'denied' || caseStatus === 'appeal_denied') && (
            <Button size="sm" loading={loading} onClick={() => generate('denial_notice')}>
              Generate Denial Letter
            </Button>
          )}
        </div>
      )}
      {message && <p className="mt-3 text-sm text-green-700">{message}</p>}
      {error && (
        <p className="mt-3 text-sm text-gov-danger" role="alert">
          {error}
        </p>
      )}
    </Card>
  );
}
