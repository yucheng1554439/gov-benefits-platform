'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { ApiClientError, api } from '@/lib/api/client';
import type { Agency, ApiListResponse, Case, Program } from '@/lib/api/types';
import { getSession, updateAgencyId } from '@/lib/auth/session';
import { Card } from '@/components/ui/Card';
import { Input, Select } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';

const STEPS = ['Program', 'Household', 'Income', 'Location', 'Review'];

const EMPLOYMENT_LABELS: Record<string, string> = {
  employed_full_time: 'Employed Full-Time',
  employed_part_time: 'Employed Part-Time',
  unemployed: 'Unemployed',
  retired: 'Retired',
};

function formatError(err: unknown, fallback: string): string {
  if (err instanceof ApiClientError) return err.message;
  if (err instanceof Error && err.message) return err.message;
  return fallback;
}

/** Accept flat Program[] or legacy AgencyProgram[] from the API. */
function normalizePrograms(data: unknown[]): Program[] {
  return data.flatMap((item) => {
    if (!item || typeof item !== 'object') return [];
    const row = item as Record<string, unknown>;
    const nested = row.program;
    if (nested && typeof nested === 'object') {
      const p = nested as Record<string, unknown>;
      const id = typeof p.id === 'string' ? p.id : typeof row.program_id === 'string' ? row.program_id : '';
      const name = typeof p.name === 'string' ? p.name : '';
      const code = typeof p.code === 'string' ? p.code : '';
      if (!id || !name) return [];
      return [{ id, code, name, description: typeof p.description === 'string' ? p.description : undefined }];
    }
    const id = typeof row.id === 'string' ? row.id : '';
    const name = typeof row.name === 'string' ? row.name : '';
    const code = typeof row.code === 'string' ? row.code : '';
    if (!id || !name) return [];
    return [{ id, code, name, description: typeof row.description === 'string' ? row.description : undefined }];
  });
}

export default function ApplyPage() {
  const router = useRouter();
  const [step, setStep] = useState(0);
  const [agencies, setAgencies] = useState<Agency[]>([]);
  const [programs, setPrograms] = useState<Program[]>([]);
  const [loadingAgencies, setLoadingAgencies] = useState(true);
  const [loadingPrograms, setLoadingPrograms] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [loadError, setLoadError] = useState('');
  const [programsError, setProgramsError] = useState('');
  const [error, setError] = useState('');
  const [form, setForm] = useState({
    agency_id: '',
    program_id: '',
    household_size: '1',
    annual_income: '',
    employment_status: 'employed_full_time',
    zip_code: '',
    census_tract: '',
    priority: 'normal',
  });

  useEffect(() => {
    const session = getSession();
    setLoadingAgencies(true);
    setLoadError('');

    api
      .get<ApiListResponse<Agency>>('/agencies', { skipAuth: true })
      .then((res) => {
        setAgencies(res.data ?? []);
        const agencyId = session?.agency_id ?? res.data?.[0]?.id ?? '';
        if (!agencyId) {
          setLoadError('No agencies are available. Contact your administrator.');
          return;
        }
        setForm((f) => ({ ...f, agency_id: agencyId }));
        return loadPrograms(agencyId);
      })
      .catch((err) => setLoadError(formatError(err, 'Unable to load agencies.')))
      .finally(() => setLoadingAgencies(false));
  }, []);

  const loadPrograms = async (agencyId: string) => {
    setLoadingPrograms(true);
    setProgramsError('');
    try {
      const res = await api.get<ApiListResponse<Program>>(`/agencies/${agencyId}/programs`, {
        skipAuth: true,
      });
      const list = normalizePrograms(res.data ?? []);
      setPrograms(list);
      setForm((f) => ({
        ...f,
        agency_id: agencyId,
        program_id: list.some((p) => p.id === f.program_id) ? f.program_id : list[0]?.id ?? '',
      }));
      if (list.length === 0) {
        setProgramsError('No programs are enabled for this agency.');
      }
    } catch (err) {
      setPrograms([]);
      setProgramsError(formatError(err, 'Unable to load programs for the selected agency.'));
    } finally {
      setLoadingPrograms(false);
    }
  };

  const update = (field: string, value: string) => setForm((f) => ({ ...f, [field]: value }));

  const validateStep = (currentStep: number): string | null => {
    if (currentStep === 0) {
      if (!form.agency_id) return 'Select an agency to continue.';
      if (programs.length === 0) return 'No programs are available for the selected agency.';
      if (!form.program_id) return 'Select a program to continue.';
      if (!programs.some((p) => p.id === form.program_id)) {
        return 'Select a valid program to continue.';
      }
    }
    if (currentStep === 1) {
      const size = parseInt(form.household_size, 10);
      if (!Number.isFinite(size) || size < 1) return 'Household size must be at least 1.';
    }
    if (currentStep === 2) {
      const income = parseFloat(form.annual_income);
      if (form.annual_income !== '' && (!Number.isFinite(income) || income < 0)) {
        return 'Annual income must be zero or greater.';
      }
    }
    return null;
  };

  const validateSubmit = (): string | null => {
    for (let i = 0; i < STEPS.length - 1; i += 1) {
      const message = validateStep(i);
      if (message) return message;
    }
    if (!form.program_id || !programs.some((p) => p.id === form.program_id)) {
      return 'Select a valid program before submitting.';
    }
    if (!form.zip_code.trim()) return 'ZIP code is required before submitting.';
    return null;
  };

  const next = () => {
    const message = validateStep(step);
    if (message) {
      setError(message);
      return;
    }
    setError('');
    setStep((s) => Math.min(s + 1, STEPS.length - 1));
  };

  const back = () => {
    setError('');
    setStep((s) => Math.max(s - 1, 0));
  };

  const submit = async () => {
    const validationMessage = validateSubmit();
    if (validationMessage) {
      setError(validationMessage);
      return;
    }

    setSubmitting(true);
    setError('');
    try {
      updateAgencyId(form.agency_id);
      const created = await api.post<Case>(
        '/applications',
        {
          agency_id: form.agency_id,
          program_id: form.program_id,
          household_size: parseInt(form.household_size, 10),
          annual_income: parseFloat(form.annual_income) || 0,
          employment_status: form.employment_status,
          zip_code: form.zip_code.trim(),
          census_tract: form.census_tract.trim(),
          priority: form.priority,
          form_data: {},
        },
        { agencyId: form.agency_id },
      );
      router.push(`/citizen/cases/${created.id}`);
    } catch (err) {
      setError(formatError(err, 'Unable to submit your application. Please try again.'));
    } finally {
      setSubmitting(false);
    }
  };

  const selectedAgency = agencies.find((a) => a.id === form.agency_id);
  const selectedProgram = programs.find((p) => p.id === form.program_id);

  if (loadingAgencies) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-gov-navy border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Apply for Benefits</h1>
        <p className="text-gov-slate">Complete all steps to submit your application</p>
      </div>

      {loadError && (
        <div className="rounded-md border border-gov-danger/30 bg-red-50 px-4 py-3 text-sm text-gov-danger" role="alert">
          {loadError}
        </div>
      )}

      <nav aria-label="Application steps" className="flex gap-2">
        {STEPS.map((label, i) => (
          <div
            key={label}
            className={`flex-1 rounded-md px-2 py-2 text-center text-xs font-medium ${
              i === step ? 'bg-gov-navy text-white' : i < step ? 'bg-gov-gold/20 text-gov-navy' : 'bg-gov-surface text-gov-slate'
            }`}
          >
            {i + 1}. {label}
          </div>
        ))}
      </nav>

      <Card title={`Step ${step + 1}: ${STEPS[step]}`}>
        {step === 0 && (
          <div className="space-y-4">
            <Select
              label="Agency"
              value={form.agency_id}
              onChange={(e) => {
                update('agency_id', e.target.value);
                updateAgencyId(e.target.value);
                void loadPrograms(e.target.value);
              }}
              options={agencies.map((a) => ({ value: a.id, label: a.name }))}
            />
            {loadingPrograms ? (
              <p className="text-sm text-gov-slate">Loading programs…</p>
            ) : (
              <Select
                label="Program"
                value={form.program_id}
                onChange={(e) => update('program_id', e.target.value)}
                disabled={programs.length === 0}
                options={programs.map((p) => ({ value: p.id, label: p.name }))}
              />
            )}
            {programsError && <p className="text-sm text-gov-danger">{programsError}</p>}
          </div>
        )}
        {step === 1 && (
          <Input
            label="Household Size"
            type="number"
            min={1}
            value={form.household_size}
            onChange={(e) => update('household_size', e.target.value)}
            required
          />
        )}
        {step === 2 && (
          <div className="space-y-4">
            <Input
              label="Annual Income ($)"
              type="number"
              min={0}
              value={form.annual_income}
              onChange={(e) => update('annual_income', e.target.value)}
            />
            <Select
              label="Employment Status"
              value={form.employment_status}
              onChange={(e) => update('employment_status', e.target.value)}
              options={[
                { value: 'employed_full_time', label: 'Employed Full-Time' },
                { value: 'employed_part_time', label: 'Employed Part-Time' },
                { value: 'unemployed', label: 'Unemployed' },
                { value: 'retired', label: 'Retired' },
              ]}
            />
          </div>
        )}
        {step === 3 && (
          <div className="space-y-4">
            <Input
              label="ZIP Code"
              value={form.zip_code}
              onChange={(e) => update('zip_code', e.target.value)}
              required
            />
            <Input label="Census Tract" value={form.census_tract} onChange={(e) => update('census_tract', e.target.value)} />
            <Select
              label="Priority"
              value={form.priority}
              onChange={(e) => update('priority', e.target.value)}
              options={[
                { value: 'normal', label: 'Normal' },
                { value: 'high', label: 'High' },
                { value: 'urgent', label: 'Urgent' },
              ]}
            />
          </div>
        )}
        {step === 4 && (
          <dl className="space-y-2 text-sm">
            <div className="flex justify-between gap-4">
              <dt className="text-gov-slate">Agency</dt>
              <dd className="text-right">{selectedAgency?.name ?? '—'}</dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-gov-slate">Program</dt>
              <dd className="text-right">{selectedProgram?.name ?? '—'}</dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-gov-slate">Household</dt>
              <dd>{form.household_size}</dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-gov-slate">Income</dt>
              <dd>${form.annual_income || '0'}</dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-gov-slate">Employment</dt>
              <dd>{EMPLOYMENT_LABELS[form.employment_status] ?? form.employment_status}</dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-gov-slate">ZIP</dt>
              <dd>{form.zip_code || '—'}</dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-gov-slate">Priority</dt>
              <dd className="capitalize">{form.priority}</dd>
            </div>
          </dl>
        )}
        {error && (
          <p className="mt-4 text-sm text-gov-danger" role="alert">
            {error}
          </p>
        )}
      </Card>

      <div className="flex justify-between">
        <Button variant="outline" onClick={back} disabled={step === 0 || submitting}>
          Back
        </Button>
        {step < STEPS.length - 1 ? (
          <Button onClick={next} disabled={!!loadError || (step === 0 && (loadingPrograms || programs.length === 0))}>
            Continue
          </Button>
        ) : (
          <Button
            onClick={submit}
            loading={submitting}
            disabled={!!loadError || !form.program_id || programs.length === 0}
          >
            Submit Application
          </Button>
        )}
      </div>
    </div>
  );
}
