'use client';

import { useCallback, useEffect, useState } from 'react';
import { api } from '@/lib/api/client';
import type { ApiListResponse, Notification } from '@/lib/api/types';

export function useNotifications(unreadOnly = false) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchNotifications = useCallback(async () => {
    setLoading(true);
    try {
      const query = unreadOnly ? '?unread=true' : '';
      const res = await api.get<ApiListResponse<Notification>>(`/notifications${query}`);
      setNotifications(res.data ?? []);
    } catch {
      setNotifications([]);
    } finally {
      setLoading(false);
    }
  }, [unreadOnly]);

  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  const markRead = useCallback(
    async (id: string) => {
      await api.patch(`/notifications/${id}/read`);
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, is_read: true } : n)),
      );
    },
    [],
  );

  return { notifications, loading, markRead, refetch: fetchNotifications };
}
