'use client';

import { WorkflowDesigner } from '@/components/workflow/WorkflowDesigner';
import { useFeatureFlag } from '@/lib/feature-flags/useFeatureFlag';

export default function AdminWorkflowPage() {
  const enabled = useFeatureFlag('new_workflow_engine');

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gov-navy">Workflow Designer</h1>
        <p className="text-gov-slate">
          Visual case status transitions {enabled ? '(engine enabled)' : '(engine disabled)'}
        </p>
      </div>
      <WorkflowDesigner />
    </div>
  );
}
