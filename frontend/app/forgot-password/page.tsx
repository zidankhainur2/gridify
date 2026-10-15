'use client';

import { useState } from 'react';
import Link from 'next/link';
import { ArrowLeft, Mail, CheckCircle2 } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [isSent, setIsSent] = useState(false);

  const handleReset = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email) {
      toast.error('Masukkan alamat email terdaftar');
      return;
    }

    setLoading(true);

    try {
      await apiClient.auth.forgotPassword(email.trim());
      setIsSent(true);
      toast.success('Permintaan Terkirim', {
        description: 'Jika email terdaftar, instruksi reset akan dikirimkan.',
      });
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : 'Gagal mengirim instruksi reset password.';
      toast.error('Gagal', { description: message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-surface-base px-4 py-12">
      <div className="w-full max-w-md p-8 rounded-xl bg-surface-raised border border-border-default shadow-xl">
        <Link
          href="/login"
          className="inline-flex items-center gap-2 text-text-muted hover:text-text-primary text-sm mb-6 transition-colors"
        >
          <ArrowLeft size={16} /> Kembali ke Login
        </Link>

        <div className="text-center mb-8">
          <div className="w-12 h-12 rounded-xl bg-surface-elevated border border-border-default flex items-center justify-center mx-auto mb-4 text-accent">
            <Mail size={22} />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-text-primary mb-1">
            Lupa Password
          </h1>
          <p className="text-sm text-text-muted">
            {isSent
              ? `Instruksi pemulihan telah dikirim ke ${email}.`
              : 'Masukkan email akun Anda untuk menerima tautan pemulihan.'}
          </p>
        </div>

        {!isSent ? (
          <form onSubmit={handleReset} className="space-y-4">
            <Input
              label="Email Terdaftar"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="nama@email.com"
              required
              autoComplete="email"
            />

            <Button
              type="submit"
              variant="primary"
              className="w-full mt-2"
              isLoading={loading}
            >
              Kirim Link Pemulihan
            </Button>
          </form>
        ) : (
          <div className="space-y-4">
            <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-emerald-400 text-sm flex items-start gap-3">
              <CheckCircle2 size={18} className="shrink-0 mt-0.5" />
              <span>
                Periksa kotak masuk atau spam email Anda. Ikuti petunjuk untuk mengatur ulang password.
              </span>
            </div>
            <Button
              onClick={() => setIsSent(false)}
              variant="secondary"
              className="w-full"
            >
              Kirim Ulang
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}