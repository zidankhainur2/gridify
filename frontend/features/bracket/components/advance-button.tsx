'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowRightCircle } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';

export default function AdvanceButton({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleAdvance = async () => {
    setLoading(true);

    try {
      await apiClient.bracket.advanceKnockout(tournamentId);
      toast.success('Babak Knockout Berhasil Dibuat!', {
        description: 'Pemenang fase grup telah dimajukan ke bracket playoff.',
      });

      setIsOpen(false);
      router.refresh();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal memproses stage berikutnya.';
      toast.error('Gagal Memajukan Stage', { description: msg });
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Button
        onClick={() => setIsOpen(true)}
        variant="primary"
        size="lg"
        className="w-full mt-4"
      >
        <span>Lanjut ke Babak Knockout Playoff</span>
        <ArrowRightCircle size={18} />
      </Button>

      <ConfirmDialog
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        onConfirm={handleAdvance}
        title="Lanjut ke Babak Knockout?"
        description="Pastikan seluruh pertandingan fase grup telah selesai dihitung. Peringkat 1 dan 2 dari masing-masing grup akan otomatis diundi ke bagan babak gugur (Knockout Playoff)."
        confirmText="Buat Babak Knockout"
        cancelText="Batal"
        variant="primary"
        isLoading={loading}
      />
    </>
  );
}
