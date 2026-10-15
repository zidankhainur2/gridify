'use client';

import * as React from 'react';
import { Modal } from './modal';
import { Button } from './button';
import { AlertTriangle } from 'lucide-react';

export interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void | Promise<void>;
  title: string;
  message?: string;
  description?: string;
  confirmLabel?: string;
  confirmText?: string;
  cancelLabel?: string;
  cancelText?: string;
  variant?: 'primary' | 'secondary' | 'danger';
  isDanger?: boolean;
  isLoading?: boolean;
}

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  description,
  confirmLabel,
  confirmText,
  cancelLabel,
  cancelText,
  variant,
  isDanger,
  isLoading = false,
}: ConfirmDialogProps) {
  const displayMessage = description || message || '';
  const displayConfirm = confirmText || confirmLabel || 'Konfirmasi';
  const displayCancel = cancelText || cancelLabel || 'Batal';
  const isDestructive = isDanger || variant === 'danger';

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="sm">
      <div className="flex items-start gap-3.5">
        {isDestructive && (
          <div className="shrink-0 p-2 rounded-lg bg-danger/10 text-danger border border-danger/20">
            <AlertTriangle className="h-5 w-5" />
          </div>
        )}
        <div className="flex-1">
          <h4 className="text-sm font-semibold text-text-primary">{title}</h4>
          {displayMessage && (
            <p className="mt-1.5 text-xs text-text-muted leading-relaxed">
              {displayMessage}
            </p>
          )}
        </div>
      </div>

      <div className="mt-6 flex items-center justify-end gap-2.5 pt-3 border-t border-border-subtle">
        <Button
          type="button"
          variant="secondary"
          size="sm"
          onClick={onClose}
          disabled={isLoading}
        >
          {displayCancel}
        </Button>
        <Button
          type="button"
          variant={isDestructive ? 'danger' : 'primary'}
          size="sm"
          onClick={onConfirm}
          isLoading={isLoading}
        >
          {displayConfirm}
        </Button>
      </div>
    </Modal>
  );
}
