'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { ApiListResponse, FeatureFlag } from '@/lib/api/types';

const flagCache = new Map<string, boolean>();

export function useFeatureFlag(flagKey: string, defaultValue = false): boolean {
  const [enabled, setEnabled] = useState(() => flagCache.get(flagKey) ?? defaultValue);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const res = await api.get<ApiListResponse<FeatureFlag>>('/feature-flags');
        res.data.forEach((f) => flagCache.set(f.flag_key, f.is_enabled));
        if (!cancelled) {
          setEnabled(flagCache.get(flagKey) ?? defaultValue);
        }
      } catch {
        if (!cancelled) setEnabled(defaultValue);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, [flagKey, defaultValue]);

  return enabled;
}
