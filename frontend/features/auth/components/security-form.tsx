'use client';

import { useState } from 'react';
import { Lock, ShieldCheck, KeyRound, Eye, EyeOff } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { User } from '@/types/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function SecurityForm({ user }: { user: User }) {
  const [loading, setLoading] = useState(false);
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  const handleUpdatePassword = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!currentPassword || !newPassword) {
      toast.error('Harap isi kata sandi saat ini dan kata sandi baru');
      return;
    }

    if (newPassword !== confirmPassword) {
      toast.error('Konfirmasi kata sandi baru tidak cocok');
      return;
    }

    if (newPassword.length < 8) {
      toast.error('Kata sandi baru minimal 8 karakter');
      return;
    }

    setLoading(true);

    try {
      await apiClient.users.updatePassword(currentPassword, newPassword);
      toast.success('Kata Sandi Berhasil Diperbarui!');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal memperbarui kata sandi.';
      toast.error('Gagal Mengubah Password', { description: msg });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="rounded-xl bg-surface-raised border border-border-default overflow-hidden h-fit">
      {/* Header */}
      <div className="px-6 py-4 border-b border-border-subtle flex items-center gap-3 bg-surface-elevated/40">
        <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
          <ShieldCheck size={16} />
        </div>
        <div>
          <h3 className="font-semibold text-text-primary text-base">
            Keamanan Akun
          </h3>
          <p className="text-xs text-text-muted">
            Ubah kata sandi dan proteksi akun Anda
          </p>
        </div>
      </div>

      <div className="p-6 space-y-6">
        {/* Email Info */}
        <div className="space-y-1.5">
          <label className="text-xs font-medium text-text-secondary uppercase tracking-wider">
            Email Terdaftar
          </label>
          <div className="px-3.5 py-2.5 rounded-lg bg-surface-base border border-border-subtle text-sm text-text-muted font-mono truncate">
            {user.email}
          </div>
          <p className="text-[11px] text-text-muted">
            Email login tidak dapat diubah secara mandiri.
          </p>
        </div>

        <div className="h-px w-full bg-border-subtle" />

        {/* Change Password Form */}
        <form onSubmit={handleUpdatePassword} className="space-y-3.5">
          <div className="flex items-center justify-between">
            <label className="text-xs font-medium text-text-secondary uppercase tracking-wider flex items-center gap-1.5">
              <KeyRound size={13} /> Ganti Kata Sandi
            </label>
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="text-xs text-text-muted hover:text-text-primary transition-colors flex items-center gap-1"
            >
              {showPassword ? <EyeOff size={13} /> : <Eye size={13} />}
              <span>{showPassword ? 'Sembunyikan' : 'Lihat'}</span>
            </button>
          </div>

          <Input
            label="Kata Sandi Saat Ini"
            type={showPassword ? 'text' : 'password'}
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            placeholder="••••••••"
            required
            autoComplete="current-password"
          />

          <Input
            label="Kata Sandi Baru"
            type={showPassword ? 'text' : 'password'}
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            placeholder="Min. 8 karakter"
            required
            autoComplete="new-password"
          />

          <Input
            label="Konfirmasi Kata Sandi Baru"
            type={showPassword ? 'text' : 'password'}
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Ulangi kata sandi baru"
            required
            autoComplete="new-password"
          />

          <Button
            type="submit"
            variant="secondary"
            className="w-full mt-2"
            isLoading={loading}
          >
            <Lock size={15} />
            Perbarui Kata Sandi
          </Button>
        </form>
      </div>
    </div>
  );
}