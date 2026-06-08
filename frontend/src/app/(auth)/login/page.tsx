'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/hooks/useAuth';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';

const DEMO_ACCOUNTS = [
  { email: 'citizen1@example.com', role: 'Citizen' },
  { email: 'worker1@dpss.lacounty.gov', role: 'Case Worker' },
  { email: 'supervisor1@dpss.lacounty.gov', role: 'Supervisor' },
  { email: 'admin@dpss.lacounty.gov', role: 'Admin' },
];

export default function LoginPage() {
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('Password123!');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await login(email, password);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  const fillDemo = (demoEmail: string) => {
    setEmail(demoEmail);
    setPassword('Password123!');
  };

  return (
    <Card title="Sign In" description="Government Benefits Portal">
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          autoComplete="email"
        />
        <Input
          label="Password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          autoComplete="current-password"
        />
        {error && (
          <p className="text-sm text-gov-danger" role="alert">
            {error}
          </p>
        )}
        <Button type="submit" className="w-full" loading={loading}>
          Sign In
        </Button>
      </form>

      <div className="mt-6 border-t border-gov-border pt-4">
        <p className="mb-2 text-xs font-medium uppercase text-gov-slate">Demo Accounts</p>
        <div className="space-y-1">
          {DEMO_ACCOUNTS.map((acc) => (
            <button
              key={acc.email}
              type="button"
              onClick={() => fillDemo(acc.email)}
              className="block w-full rounded px-2 py-1 text-left text-sm text-gov-navy hover:bg-gov-surface"
            >
              {acc.role}: {acc.email}
            </button>
          ))}
        </div>
        <p className="mt-2 text-xs text-gov-slate">Password: Password123!</p>
      </div>

      <p className="mt-4 text-center text-sm text-gov-slate">
        New applicant?{' '}
        <Link href="/register" className="font-medium text-gov-navy underline">
          Register
        </Link>
      </p>
    </Card>
  );
}
