'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';

interface DeleteButtonProps {
  id: string;
  tournamentTitle?: string;
  className?: string;
  isIconOnly?: boolean;
  redirectTo?: string;
}

export function DeleteTournamentButton({
  id,
  tournamentTitle = 'turnamen ini',
  className = '',
  isIconOnly = true,
  redirectTo,
}: DeleteButtonProps) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    setIsDeleting(true);

    try {
      await apiClient.tournaments.delete(id);
      toast.success('Turnamen Berhasil Dihapus', {
        description: `"${tournamentTitle}" telah dihapus secara permanen.`,
      });

      setIsOpen(false);
      if (redirectTo) {
        router.push(redirectTo);
      } else {
        router.refresh();
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal menghapus turnamen.';
      toast.error('Gagal Menghapus Turnamen', { description: msg });
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <>
      <button
        type="button"
        onClick={(e) => {
          e.preventDefault();
          e.stopPropagation();
          setIsOpen(true);
        }}
        className={`inline-flex items-center gap-2 rounded-lg text-text-muted hover:text-danger hover:bg-danger/10 transition-colors p-2 text-sm font-medium ${className}`}
        title="Hapus Turnamen"
        aria-label="Hapus Turnamen"
      >
        <Trash2 size={16} />
        {!isIconOnly && <span>Hapus Turnamen</span>}
      </button>

      <ConfirmDialog
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        onConfirm={handleDelete}
        title="Hapus Turnamen?"
        description={`Apakah Anda yakin ingin menghapus "${tournamentTitle}"? Seluruh data bagan, stage, peserta, dan skor pertandingan akan dihapus secara permanen.`}
        confirmText="Hapus Permanen"
        cancelText="Batal"
        variant="danger"
        isLoading={isDeleting}
      />
    </>
  );
}