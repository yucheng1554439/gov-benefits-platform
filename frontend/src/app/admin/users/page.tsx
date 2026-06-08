'use client';

import { Card } from '@/components/ui/Card';
import { Table } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';

const DEMO_USERS = [
  { id: '1', email: 'citizen1@example.com', name: 'Maria Garcia', role: 'citizen', status: 'active' },
  { id: '2', email: 'citizen2@example.com', name: 'James Wilson', role: 'citizen', status: 'active' },
  { id: '3', email: 'worker1@dpss.lacounty.gov', name: 'Sarah Chen', role: 'case_worker', status: 'active' },
  { id: '4', email: 'worker2@dpss.lacounty.gov', name: 'David Martinez', role: 'case_worker', status: 'active' },
  { id: '5', email: 'supervisor1@dpss.lacounty.gov', name: 'Patricia Johnson', role: 'supervisor', status: 'active' },
  { id: '6', email: 'admin@dpss.lacounty.gov', name: 'Robert Admin', role: 'admin', status: 'active' },
];

export default function AdminUsersPage() {
  const columns = [
    { key: 'name', header: 'Name', render: (row: typeof DEMO_USERS[0]) => row.name },
    { key: 'email', header: 'Email', render: (row: typeof DEMO_USERS[0]) => row.email },
    { key: 'role', header: 'Role', render: (row: typeof DEMO_USERS[0]) => <Badge>{row.role}</Badge> },
    { key: 'status', header: 'Status', render: (row: typeof DEMO_USERS[0]) => <Badge variant="success">{row.status}</Badge> },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">User Management</h1>
        <p className="text-gov-slate">Manage portal users and role assignments</p>
      </div>
      <Card title="Users" description="Seed data users from LA County DPSS">
        <Table columns={columns} data={DEMO_USERS} keyField="id" />
      </Card>
    </div>
  );
}
