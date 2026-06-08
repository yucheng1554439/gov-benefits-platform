'use client';

import Link from 'next/link';
import { use, useEffect, useState } from 'react';
import { api, API_BASE } from '@/lib/api/client';
import type { ApiListResponse, Letter } from '@/lib/api/types';
import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Button } from '@/components/ui/Button';
import { formatDate } from '@/lib/utils';

export default function CitizenLettersPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const [letters, setLetters] = useState<Letter[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .get<ApiListResponse<Letter>>(`/cases/${id}/letters`)
      .then((res) => setLetters(res.data ?? []))
      .catch(() => setLetters([]))
      .finally(() => setLoading(false));
  }, [id]);

  const downloadLetter = async (letterId: string) => {
    const token = localStorage.getItem('access_token');
    const agency = localStorage.getItem('agency_id');
    const response = await fetch(`${API_BASE}/letters/${letterId}/download`, {
      headers: {
        Authorization: `Bearer ${token}`,
        ...(agency ? { 'X-Agency-ID': agency } : {}),
      },
    });
    if (!response.ok) return;
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `letter-${letterId}.pdf`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const columns = [
    { key: 'type', header: 'Type', render: (row: Letter) => row.letter_type.replace(/_/g, ' ') },
    { key: 'date', header: 'Generated', render: (row: Letter) => formatDate(row.generated_at) },
    {
      key: 'actions',
      header: '',
      render: (row: Letter) => (
        <Button size="sm" variant="outline" onClick={() => downloadLetter(row.id)}>
          Download PDF
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <Link href={`/citizen/cases/${id}`} className="text-sm text-gov-navy underline">
          ← Back to Case
        </Link>
        <h1 className="mt-2 text-2xl font-bold text-gov-navy">Case Letters</h1>
      </div>
      <Card>
        {loading ? <p className="text-gov-slate">Loading...</p> : <Table columns={columns} data={letters} keyField="id" />}
      </Card>
    </div>
  );
}
