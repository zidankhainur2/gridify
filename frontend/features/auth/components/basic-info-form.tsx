'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { User as UserIcon, Phone, AtSign, Save, Contact } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { User } from '@/types/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function BasicInfoForm({ user }: { user: User }) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  const [firstName, setFirstName] = useState(user.first_name || '');
  const [lastName, setLastName] = useState(user.last_name || '');
  const [username, setUsername] = useState(user.username || '');
  const [phone, setPhone] = useState(user.phone || '');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!username.trim()) {
      toast.error('Username wajib diisi');
      return;
    }

    setLoading(true);

    try {
      await apiClient.users.updateMe({
        first_name: firstName.trim() || undefined,
        last_name: lastName.trim() || undefined,
        username: username.trim(),
        phone: phone.trim() || undefined,
      });

      toast.success('Profil Berhasil Diperbarui!');
      router.refresh();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal memperbarui profil.';
      toast.error('Gagal Menyimpan', { description: msg });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="rounded-xl bg-surface-raised border border-border-default overflow-hidden">
      {/* Header Form */}
      <div className="px-6 py-4 border-b border-border-subtle flex items-center gap-3 bg-surface-elevated/40">
        <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
          <Contact size={16} />
        </div>
        <div>
          <h3 className="font-semibold text-text-primary text-base">
            Informasi Pribadi & Penyelenggara
          </h3>
          <p className="text-xs text-text-muted">
            Identitas yang ditampilkan pada profil turnamen esport Anda.
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="p-6 space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Input
            label="Nama Depan"
            type="text"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            placeholder="Alex"
          />
          <Input
            label="Nama Belakang"
            type="text"
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
            placeholder="Rivaldi"
          />
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Input
            label="Username Akun"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="alexrivaldi"
            required
            autoComplete="username"
          />
          <Input
            label="Nomor WhatsApp / Kontak"
            type="tel"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            placeholder="08123456789"
          />
        </div>

        <div className="pt-3 flex justify-end border-t border-border-subtle">
          <Button type="submit" variant="primary" isLoading={loading}>
            <Save size={16} />
            Simpan Perubahan
          </Button>
        </div>
      </form>
    </div>
  );
}