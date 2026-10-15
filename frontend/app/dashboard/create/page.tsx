import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import CreateTournamentForm from '@/features/tournaments/components/create-tournament-form';

export default async function CreateTournamentPage() {
  let games: any[] = [];
  try {
    games = await apiClient.games.list();
  } catch {
    games = [];
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6 animate-in fade-in duration-200 pb-12">
      <div>
        <Link
          href="/dashboard/tournaments"
          className="inline-flex items-center text-text-muted hover:text-text-primary transition-colors mb-3 gap-1.5 text-xs font-medium"
        >
          <ArrowLeft size={14} /> Kembali ke Turnamen Saya
        </Link>
        <h1 className="text-2xl md:text-3xl font-bold tracking-tight text-text-primary">
          Setup Turnamen Baru
        </h1>
        <p className="text-sm text-text-muted mt-1">
          Konfigurasi judul, game, format eliminasi/liga/battle royale, dan aturan pertandingan.
        </p>
      </div>

      <CreateTournamentForm games={games} />
    </div>
  );
}
