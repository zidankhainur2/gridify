'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { LogIn, Eye, EyeOff, Gamepad2 } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function LoginPage() {
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!identifier || !password) {
      toast.error('Harap lengkapi semua kolom');
      return;
    }

    setLoading(true);

    try {
      await apiClient.auth.login(identifier.trim(), password);
      toast.success('Berhasil masuk', { description: 'Mengalihkan ke dashboard...' });
      router.push('/dashboard');
      router.refresh();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Login gagal. Periksa kredensial Anda.';
      toast.error('Gagal Masuk', { description: message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-surface-base px-4 py-12">
      <div className="w-full max-w-md p-8 rounded-xl bg-surface-raised border border-border-default shadow-xl">
        <div className="flex flex-col items-center text-center mb-8">
          <Link href="/" className="flex items-center gap-2 mb-4 group">
            <div className="w-10 h-10 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent group-hover:border-accent transition-colors">
              <Gamepad2 size={22} />
            </div>
            <span className="text-xl font-bold tracking-tight text-text-primary">Gridify</span>
          </Link>
          <h1 className="text-2xl font-bold tracking-tight text-text-primary">
            Masuk ke Akun
          </h1>
          <p className="text-sm text-text-muted mt-1">
            Kelola turnamen dan pantau bracket secara real-time
          </p>
        </div>

        <form onSubmit={handleLogin} className="space-y-4">
          <Input
            label="Email atau Username"
            type="text"
            value={identifier}
            onChange={(e) => setIdentifier(e.target.value)}
            placeholder="nama@email.com atau username"
            required
            autoComplete="username"
          />

          <div className="space-y-1.5">
            <div className="flex justify-between items-center">
              <label className="text-xs font-medium text-text-secondary uppercase tracking-wider">
                Password
              </label>
              <Link
                href="/forgot-password"
                className="text-xs text-text-muted hover:text-accent transition-colors"
              >
                Lupa Password?
              </Link>
            </div>
            <div className="relative">
              <Input
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
                autoComplete="current-password"
                className="pr-10"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-3 text-text-muted hover:text-text-primary transition-colors"
                aria-label={showPassword ? 'Sembunyikan password' : 'Tampilkan password'}
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </div>

          <Button
            type="submit"
            variant="primary"
            className="w-full mt-2"
            isLoading={loading}
          >
            <LogIn size={18} />
            Masuk Sekarang
          </Button>
        </form>

        <p className="text-center text-sm text-text-muted mt-6">
          Belum punya akun?{' '}
          <Link
            href="/register"
            className="text-accent hover:text-accent-hover font-medium transition-colors"
          >
            Daftar Gratis
          </Link>
        </p>
      </div>
    </div>
  );
}