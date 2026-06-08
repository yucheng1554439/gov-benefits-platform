import { cn } from '@/lib/utils';

type BadgeVariant = 'default' | 'success' | 'warning' | 'danger' | 'info';

interface BadgeProps {
  children: React.ReactNode;
  variant?: BadgeVariant;
  className?: string;
}

const variants: Record<BadgeVariant, string> = {
  default: 'bg-slate-100 text-gov-slate',
  success: 'bg-green-100 text-gov-success',
  warning: 'bg-amber-100 text-gov-warning',
  danger: 'bg-red-100 text-gov-danger',
  info: 'bg-blue-100 text-gov-navy',
};

export function Badge({ children, variant = 'default', className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
        variants[variant],
        className,
      )}
    >
      {children}
    </span>
  );
}

export function statusBadgeVariant(status: string): BadgeVariant {
  if (['approved', 'appeal_approved', 'closed'].includes(status)) return 'success';
  if (['denied', 'appeal_denied'].includes(status)) return 'danger';
  if (['need_documents', 'supervisor_review', 'appeal_review'].includes(status)) return 'warning';
  if (['under_review', 'eligibility_review', 'submitted'].includes(status)) return 'info';
  return 'default';
}
