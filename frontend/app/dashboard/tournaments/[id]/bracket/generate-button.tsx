'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Shuffle } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';

interface GenerateBracketButtonProps {
  tournamentId: string;
  participantCount: number;
  label?: string;
  variant?: 'primary' | 'secondary' | 'danger';
}

export default function GenerateBracketButton({
  tournamentId,
  participantCount,
  label = 'Generate Bagan Turnamen',
  variant = 'primary',
}: GenerateBracketButtonProps) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleClick = () => {
    if (participantCount < 2) {
      toast.error('Minimal butuh 2 peserta untuk mengenerate bagan turnamen.');
      return;
    }
    setIsOpen(true);
  };

  const handleGenerate = async () => {
    setLoading(true);

    try {
      await apiClient.bracket.generate(tournamentId);
      toast.success('Bagan Pertandingan Berhasil Digenerate!', {
        description: 'Jadwal pertandingan dan stage telah disusun rapi.',
      });

      setIsOpen(false);
      router.refresh();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal membuat bagan.';
      toast.error('Gagal Mengenerate Bagan', { description: msg });
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Button
        onClick={handleClick}
        variant={variant === 'primary' ? 'primary' : 'secondary'}
        className="shadow-sm"
      >
        <Shuffle size={16} />
        <span>{label}</span>
      </Button>

      <ConfirmDialog
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        onConfirm={handleGenerate}
        title="Generate / Reset Bagan Pertandingan?"
        description={`Anda akan menyusun bagan pertandingan baru untuk ${participantCount} peserta terdaftar. Jika bagan sudah ada sebelumnya, seluruh skor dan progresi akan diatur ulang.`}
        confirmText="Generate Bagan Sekarang"
        cancelText="Batal"
        variant="primary"
        isLoading={loading}
      />
    </>
  );
}
