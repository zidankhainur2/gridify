'use client';

import { Suspense, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { Lock, Eye, EyeOff, ArrowLeft } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

function UpdatePasswordContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get('token') || '';
  const router = useRouter();

  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();

    if (password !== confirmPassword) {
      toast.error('Password Tidak Cocok', {
        description: 'Pastikan password baru dan konfirmasi sama persis.',
      });
      return;
    }

    if (password.length < 8) {
      toast.error('Password Lemah', {
        description: 'Minimal 8 karakter.',
      });
      return;
    }

    setLoading(true);

    try {
      if (token) {
        await apiClient.auth.resetPassword(token, password);
        toast.success('Password Berhasil Diubah!', {
          description: 'Silakan masuk menggunakan password baru Anda.',
        });
        router.push('/login');
      } else {
        toast.error('Token Tidak Valid', {
          description: 'Token reset password tidak ditemukan di tautan Anda.',
        });
      }
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : 'Gagal mengubah password.';
      toast.error('Gagal', { description: message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="w-full max-w-md p-8 rounded-xl bg-surface-raised border border-border-default shadow-xl">
      <Link
        href="/login"
        className="inline-flex items-center gap-2 text-text-muted hover:text-text-primary text-sm mb-6 transition-colors"
      >
        <ArrowLeft size={16} /> Kembali ke Login
      </Link>

      <div className="text-center mb-8">
        <div className="w-12 h-12 rounded-xl bg-surface-elevated border border-border-default flex items-center justify-center mx-auto mb-4 text-accent">
          <Lock size={22} />
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-text-primary mb-1">
          Atur Password Baru
        </h1>
        <p className="text-sm text-text-muted">
          Buat password baru yang kuat untuk akun Gridify Anda.
        </p>
      </div>

      <form onSubmit={handleUpdate} className="space-y-4">
        <div className="space-y-1.5">
          <label className="text-xs font-medium text-text-secondary uppercase tracking-wider">
            Password Baru
          </label>
          <div className="relative">
            <Input
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Min. 8 karakter"
              required
              autoComplete="new-password"
              className="pr-10"
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-3 top-3 text-text-muted hover:text-text-primary transition-colors"
              aria-label={showPassword ? 'Sembunyikan' : 'Lihat'}
            >
              {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
        </div>

        <div className="space-y-1.5">
          <label className="text-xs font-medium text-text-secondary uppercase tracking-wider">
            Konfirmasi Password
          </label>
          <div className="relative">
            <Input
              type={showConfirmPassword ? 'text' : 'password'}
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Ulangi password baru"
              required
              autoComplete="new-password"
              className="pr-10"
            />
            <button
              type="button"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              className="absolute right-3 top-3 text-text-muted hover:text-text-primary transition-colors"
              aria-label={showConfirmPassword ? 'Sembunyikan' : 'Lihat'}
            >
              {showConfirmPassword ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
        </div>

        <Button
          type="submit"
          variant="primary"
          className="w-full mt-4"
          isLoading={loading}
        >
          Simpan Password Baru
        </Button>
      </form>
    </div>
  );
}

export default function UpdatePasswordPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-surface-base px-4 py-12">
      <Suspense
        fallback={
          <div className="w-full max-w-md p-8 rounded-xl bg-surface-raised border border-border-default text-center text-text-muted">
            Memuat...
          </div>
        }
      >
        <UpdatePasswordContent />
      </Suspense>
    </div>
  );
}