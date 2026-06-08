'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/hooks/useAuth';
import { api } from '@/lib/api/client';
import type { Agency, ApiListResponse } from '@/lib/api/types';
import { Input, Select } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';

export default function RegisterPage() {
  const { register } = useAuth();
  const [agencies, setAgencies] = useState<Agency[]>([]);
  const [form, setForm] = useState({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    phone: '',
    agency_id: '',
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    api
      .get<ApiListResponse<Agency>>('/agencies', { skipAuth: true })
      .then((res) => {
        setAgencies(res.data ?? []);
        if (res.data?.[0]) setForm((f) => ({ ...f, agency_id: res.data[0].id }));
      })
      .catch(() => setAgencies([]));
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await register(form);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  const update = (field: string, value: string) => setForm((f) => ({ ...f, [field]: value }));

  return (
    <Card title="Create Account" description="Register as a new citizen applicant">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <Input label="First Name" value={form.first_name} onChange={(e) => update('first_name', e.target.value)} required />
          <Input label="Last Name" value={form.last_name} onChange={(e) => update('last_name', e.target.value)} required />
        </div>
        <Input label="Email" type="email" value={form.email} onChange={(e) => update('email', e.target.value)} required />
        <Input
          label="Password"
          type="password"
          value={form.password}
          onChange={(e) => update('password', e.target.value)}
          required
          hint="Minimum 8 characters"
        />
        <Input label="Phone" type="tel" value={form.phone} onChange={(e) => update('phone', e.target.value)} />
        {agencies.length > 0 && (
          <Select
            label="Agency"
            value={form.agency_id}
            onChange={(e) => update('agency_id', e.target.value)}
            options={agencies.map((a) => ({ value: a.id, label: a.name }))}
          />
        )}
        {error && <p className="text-sm text-gov-danger">{error}</p>}
        <Button type="submit" className="w-full" loading={loading}>
          Register
        </Button>
      </form>
      <p className="mt-4 text-center text-sm text-gov-slate">
        Already have an account?{' '}
        <Link href="/login" className="font-medium text-gov-navy underline">
          Sign In
        </Link>
      </p>
    </Card>
  );
}
