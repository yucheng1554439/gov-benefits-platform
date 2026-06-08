'use client';

import { useEffect, useRef } from 'react';
import { cn } from '@/lib/utils';
import { Button } from './Button';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
  size?: 'sm' | 'md' | 'lg';
}

const sizes = {
  sm: 'max-w-md',
  md: 'max-w-lg',
  lg: 'max-w-2xl',
};

export function Modal({ open, onClose, title, children, footer, size = 'md' }: ModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
    if (!open && dialog.open) dialog.close();
  }, [open]);

  return (
    <dialog
      ref={dialogRef}
      className={cn(
        'w-full rounded-lg border-0 bg-white p-0 shadow-xl backdrop:bg-black/50',
        sizes[size],
      )}
      onClose={onClose}
      aria-labelledby="modal-title"
    >
      <div className="border-b border-gov-border px-6 py-4">
        <h2 id="modal-title" className="text-lg font-semibold text-gov-navy">
          {title}
        </h2>
      </div>
      <div className="px-6 py-4">{children}</div>
      <div className="flex justify-end gap-2 border-t border-gov-border px-6 py-4">
        {footer ?? (
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        )}
      </div>
    </dialog>
  );
}
