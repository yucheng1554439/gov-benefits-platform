'use client';

import { useNotifications } from '@/lib/hooks/useNotifications';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { formatDate } from '@/lib/utils';

export default function NotificationsPage() {
  const { notifications, loading, markRead } = useNotifications();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Notifications</h1>
        <p className="text-gov-slate">Updates about your cases and applications</p>
      </div>

      <Card>
        {loading ? (
          <p className="text-gov-slate">Loading...</p>
        ) : notifications.length === 0 ? (
          <p className="text-gov-slate">No notifications.</p>
        ) : (
          <ul className="divide-y divide-gov-border">
            {notifications.map((n) => (
              <li key={n.id} className="flex items-start justify-between gap-4 py-4">
                <div>
                  <div className="flex items-center gap-2">
                    <p className="font-medium text-gov-navy">{n.title}</p>
                    {!n.is_read && <Badge variant="info">New</Badge>}
                  </div>
                  <p className="mt-1 text-sm text-gov-slate">{n.body}</p>
                  <time className="mt-1 block text-xs text-gov-slate-light">{formatDate(n.created_at)}</time>
                </div>
                {!n.is_read && (
                  <Button size="sm" variant="outline" onClick={() => markRead(n.id)}>
                    Mark Read
                  </Button>
                )}
              </li>
            ))}
          </ul>
        )}
      </Card>
    </div>
  );
}
