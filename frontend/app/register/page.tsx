'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { UserPlus, Eye, EyeOff, Gamepad2, ArrowLeft } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function RegisterPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const [formData, setFormData] = useState({
    email: '',
    username: '',
    firstName: '',
    lastName: '',
    phone: '',
    password: '',
    confirmPassword: '',
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const handleSignUp = async (e: React.FormEvent) => {
    e.preventDefault();

    if (formData.password !== formData.confirmPassword) {
      toast.error('Validasi Gagal', {
        description: 'Password dan Konfirmasi Password tidak cocok.',
      });
      return;
    }

    if (formData.password.length < 8) {
      toast.error('Password Lemah', {
        description: 'Password minimal 8 karakter.',
      });
      return;
    }

    setLoading(true);

    try {
      await apiClient.auth.register({
        email: formData.email.trim(),
        username: formData.username.trim(),
        password: formData.password,
        first_name: formData.firstName.trim(),
        last_name: formData.lastName.trim(),
        phone: formData.phone.trim() || undefined,
      });

      toast.success('Pendaftaran Berhasil!', {
        description: 'Akun Anda telah dibuat. Mengalihkan ke dashboard...',
      });
      router.push('/dashboard');
      router.refresh();
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : 'Registrasi gagal. Coba lagi.';
      toast.error('Gagal Mendaftar', { description: message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-surface-base px-4 py-12">
      <div className="w-full max-w-xl p-8 rounded-xl bg-surface-raised border border-border-default shadow-xl">
        <div className="flex flex-col items-center text-center mb-8">
          <Link href="/" className="flex items-center gap-2 mb-4 group">
            <div className="w-10 h-10 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent group-hover:border-accent transition-colors">
              <Gamepad2 size={22} />
            </div>
            <span className="text-xl font-bold tracking-tight text-text-primary">Gridify</span>
          </Link>
          <h1 className="text-2xl font-bold tracking-tight text-text-primary">
            Buat Akun Organizer
          </h1>
          <p className="text-sm text-text-muted mt-1">
            Gelar turnamen esport dengan sistem bracket otomatis dan realtime
          </p>
        </div>

        <form onSubmit={handleSignUp} className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input
              label="Nama Depan"
              name="firstName"
              type="text"
              value={formData.firstName}
              onChange={handleChange}
              placeholder="Alex"
              required
            />
            <Input
              label="Nama Belakang"
              name="lastName"
              type="text"
              value={formData.lastName}
              onChange={handleChange}
              placeholder="Rivaldi"
              required
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input
              label="Username"
              name="username"
              type="text"
              value={formData.username}
              onChange={handleChange}
              placeholder="alexrivaldi"
              required
              autoComplete="username"
            />
            <Input
              label="Nomor WhatsApp"
              name="phone"
              type="tel"
              value={formData.phone}
              onChange={handleChange}
              placeholder="08123456789"
            />
          </div>

          <Input
            label="Alamat Email"
            name="email"
            type="email"
            value={formData.email}
            onChange={handleChange}
            placeholder="alex@esport.com"
            required
            autoComplete="email"
          />

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-secondary uppercase tracking-wider">
                Password
              </label>
              <div className="relative">
                <Input
                  name="password"
                  type={showPassword ? 'text' : 'password'}
                  value={formData.password}
                  onChange={handleChange}
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
                  name="confirmPassword"
                  type={showConfirmPassword ? 'text' : 'password'}
                  value={formData.confirmPassword}
                  onChange={handleChange}
                  placeholder="Ulangi password"
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
          </div>

          <Button
            type="submit"
            variant="primary"
            className="w-full mt-4"
            isLoading={loading}
          >
            <UserPlus size={18} />
            Daftar Akun Baru
          </Button>
        </form>

        <p className="text-center text-sm text-text-muted mt-6">
          Sudah punya akun?{' '}
          <Link
            href="/login"
            className="text-accent hover:text-accent-hover font-medium transition-colors"
          >
            Masuk di sini
          </Link>
        </p>
      </div>
    </div>
  );
}