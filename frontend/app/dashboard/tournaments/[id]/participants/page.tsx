import { notFound } from 'next/navigation';
import { Users, Info, ShieldAlert } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Tournament, Participant } from '@/types/api';
import AddParticipantForm from '@/features/participants/components/add-participant-form';
import { EditParticipantModal } from '@/features/participants/components/edit-participant-modal';
import { DeleteParticipantButton } from '@/features/participants/components/delete-button';
import { EmptyState } from '@/components/ui/empty-state';

export default async function ParticipantsPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let tournament: Tournament;
  let participants: Participant[] = [];

  try {
    tournament = await apiClient.tournaments.getById(id);
    participants = await apiClient.participants.list(id);
  } catch {
    notFound();
  }

  const isLocked = tournament.status !== 'DRAFT';

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-10">
      {/* Header Info */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-border-default">
        <div>
          <h2 className="text-xl font-bold tracking-tight text-text-primary">
            Daftar Peserta Kompetisi
          </h2>
          <p className="text-xs text-text-muted mt-0.5">
            Kelola tim atau pemain yang berkompetisi di turnamen ini.
          </p>
        </div>
        <div className="text-xs font-mono text-text-muted bg-surface-raised px-3 py-1.5 rounded-lg border border-border-default self-start sm:self-auto">
          Total Peserta: <span className="font-bold text-text-primary">{participants.length}</span>
        </div>
      </div>

      {/* Lock Status Alert */}
      {isLocked && (
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-300 text-xs flex items-center gap-3">
          <ShieldAlert size={18} className="shrink-0 text-amber-400" />
          <span>
            Turnamen berstatus <strong>{tournament.status}</strong>. Perubahan peserta dikunci untuk menjaga integritas bracket dan riwayat skor.
          </span>
        </div>
      )}

      {/* Main Grid: Form Tambah & List Peserta */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Kolom Form Tambah */}
        <div className="lg:col-span-1 sticky top-20">
          <AddParticipantForm tournamentId={id} disabled={isLocked} />
        </div>

        {/* Kolom List Peserta */}
        <div className="lg:col-span-2 space-y-2.5">
          {participants.length === 0 ? (
            <div className="p-8 rounded-xl bg-surface-raised border border-border-dashed border-border-default text-center">
              <Users size={24} className="mx-auto text-text-muted mb-2 opacity-50" />
              <p className="text-sm font-medium text-text-secondary">
                Belum ada peserta terdaftar
              </p>
              <p className="text-xs text-text-muted mt-1">
                Gunakan form di samping untuk mendaftarkan tim atau pemain ke turnamen ini.
              </p>
            </div>
          ) : (
            participants.map((p, idx) => (
              <div
                key={p.id}
                className="group flex items-center justify-between p-4 rounded-xl bg-surface-raised border border-border-default hover:border-accent/40 transition-colors"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-xs font-mono font-bold text-text-muted shrink-0">
                    {idx + 1}
                  </div>
                  <div className="min-w-0">
                    <p className="font-semibold text-sm text-text-primary truncate">
                      {p.name}
                    </p>
                    {p.contact_info ? (
                      <p className="text-xs text-text-muted truncate">
                        {p.contact_info}
                      </p>
                    ) : (
                      <p className="text-xs text-text-muted/60 italic">
                        Tidak ada kontak
                      </p>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-1 shrink-0 ml-3">
                  <EditParticipantModal
                    tournamentId={id}
                    participant={p}
                    disabled={isLocked}
                  />
                  <DeleteParticipantButton
                    tournamentId={id}
                    participantId={p.id}
                    participantName={p.name}
                    disabled={isLocked}
                  />
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}