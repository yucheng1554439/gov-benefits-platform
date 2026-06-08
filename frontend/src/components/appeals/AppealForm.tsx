'use client';

import { useState } from 'react';
import { api } from '@/lib/api/client';
import type { Appeal } from '@/lib/api/types';
import { Textarea } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';

interface AppealFormProps {
  caseId: string;
  onSubmitted?: (appeal: Appeal) => void;
}

export function AppealForm({ caseId, onSubmitted }: AppealFormProps) {
  const [grounds, setGrounds] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!grounds.trim()) {
      setError('Please describe your grounds for appeal.');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const appeal = await api.post<Appeal>('/appeals', { case_id: caseId, grounds });
      setGrounds('');
      onSubmitted?.(appeal);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to file appeal');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Textarea
        label="Grounds for Appeal"
        value={grounds}
        onChange={(e) => setGrounds(e.target.value)}
        rows={5}
        placeholder="Explain why you believe the decision should be reconsidered..."
        error={error}
        required
      />
      <Button type="submit" loading={loading}>
        File Appeal
      </Button>
    </form>
  );
}
