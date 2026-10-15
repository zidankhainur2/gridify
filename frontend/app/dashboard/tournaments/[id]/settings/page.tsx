import { notFound } from 'next/navigation';
import { Settings, AlertTriangle } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Tournament } from '@/types/api';
import TournamentSettingsForm from '@/features/tournaments/components/tournament-settings-form';
import { DeleteTournamentButton } from '@/features/tournaments/components/delete-tournament-button';

export default async function SettingsPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let tournament: Tournament;
  try {
    tournament = await apiClient.tournaments.getById(id);
  } catch {
    notFound();
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6 pb-12">
      {/* Header Section */}
      <div className="pb-4 border-b border-border-default">
        <h2 className="text-xl font-bold tracking-tight text-text-primary">
          Pengaturan Turnamen
        </h2>
        <p className="text-xs text-text-muted mt-0.5">
          Kelola judul, deskripsi umum, dan status kompetisi Anda.
        </p>
      </div>

      {/* Main Settings Form */}
      <TournamentSettingsForm tournament={tournament} />

      {/* Danger Zone Section */}
      <div className="p-6 rounded-xl border border-danger/30 bg-danger/5 space-y-4">
        <div className="flex items-center gap-2 pb-2 border-b border-danger/20 text-danger font-semibold text-sm">
          <AlertTriangle size={16} />
          <span>Zona Berbahaya (Danger Zone)</span>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <p className="font-semibold text-sm text-text-primary">
              Hapus Turnamen Ini
            </p>
            <p className="text-xs text-text-muted mt-0.5 max-w-lg leading-relaxed">
              Tindakan ini akan menghapus turnamen secara permanen termasuk seluruh peserta, bagan bracket, dan skor pertandingan. Tindakan ini tidak dapat dibatalkan.
            </p>
          </div>

          <DeleteTournamentButton
            id={tournament.id}
            tournamentTitle={tournament.title}
            isIconOnly={false}
            redirectTo="/dashboard/tournaments"
            className="bg-danger/10 hover:bg-danger text-danger hover:text-white border border-danger/30 px-4 py-2 shrink-0 self-start sm:self-auto"
          />
        </div>
      </div>
    </div>
  );
}