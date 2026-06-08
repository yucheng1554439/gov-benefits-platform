'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { Agency, ApiListResponse } from '@/lib/api/types';
import { updateAgencyId } from '@/lib/auth/session';

interface AgencySwitcherProps {
  currentAgencyId: string;
}

export function AgencySwitcher({ currentAgencyId }: AgencySwitcherProps) {
  const [agencies, setAgencies] = useState<Agency[]>([]);
  const [selected, setSelected] = useState(currentAgencyId);

  useEffect(() => {
    api
      .get<ApiListResponse<Agency>>('/agencies', { skipAuth: true })
      .then((res) => setAgencies(res.data ?? []))
      .catch(() => setAgencies([]));
  }, []);

  useEffect(() => {
    setSelected(currentAgencyId);
  }, [currentAgencyId]);

  const handleChange = (agencyId: string) => {
    setSelected(agencyId);
    updateAgencyId(agencyId);
    window.location.reload();
  };

  if (agencies.length === 0) return null;

  return (
    <label className="flex items-center gap-2 text-sm">
      <span className="text-gov-slate">Agency</span>
      <select
        value={selected}
        onChange={(e) => handleChange(e.target.value)}
        className="rounded-md border border-gov-border px-2 py-1 text-gov-navy focus:outline-none focus:ring-2 focus:ring-gov-gold/40"
        aria-label="Select agency"
      >
        {agencies.map((a) => (
          <option key={a.id} value={a.id}>
            {a.name}
          </option>
        ))}
      </select>
    </label>
  );
}
