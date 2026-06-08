import { cn } from '@/lib/utils';
import type { HTMLAttributes, ReactNode } from 'react';

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  title?: string;
  description?: string;
  actions?: ReactNode;
}

export function Card({ title, description, actions, className, children, ...props }: CardProps) {
  return (
    <div
      className={cn('rounded-lg border border-gov-border bg-white shadow-sm', className)}
      {...props}
    >
      {(title || actions) && (
        <div className="flex items-start justify-between border-b border-gov-border px-6 py-4">
          <div>
            {title && <h3 className="text-lg font-semibold text-gov-navy">{title}</h3>}
            {description && <p className="mt-1 text-sm text-gov-slate">{description}</p>}
          </div>
          {actions}
        </div>
      )}
      <div className="px-6 py-4">{children}</div>
    </div>
  );
}
