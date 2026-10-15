import { ReactNode } from 'react';
import Link from 'next/link';
import { notFound, redirect } from 'next/navigation';
import { Trophy, ChevronLeft, Share2, Gamepad2, Radio } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Tournament } from '@/types/api';
import { TournamentStatusBadge } from '@/components/ui/badge';
import { GameLogoBadge } from '@/components/ui/game-logo';
import { TournamentTabNav } from '@/features/tournaments/components/tournament-tab-nav';

type Props = {
  children: ReactNode;
  params: Promise<{ id: string }>;
};

export default async function TournamentLayout({ children, params }: Props) {
  const { id } = await params;

  let tournament: Tournament;
  try {
    tournament = await apiClient.tournaments.getById(id);
  } catch {
    notFound();
  }

  const game = tournament.game || tournament.games;
  const gameName = game?.name || 'Esports Tournament';
  const formatLabel = tournament.format_type.replace(/_/g, ' ');

  return (
    <div className="flex flex-col min-h-screen bg-surface-base text-text-primary font-sans">
      {/* Top Header Bar */}
      <header className="bg-surface-raised border-b border-border-default px-4 md:px-8 py-5">
        <div className="max-w-7xl mx-auto w-full">
          {/* Breadcrumb Navigation */}
          <Link
            href="/dashboard/tournaments"
            className="inline-flex items-center text-xs font-medium text-text-muted hover:text-text-primary mb-3 transition-colors gap-1"
          >
            <ChevronLeft size={14} /> Kembali ke Turnamen Saya
          </Link>

          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="flex items-center gap-3.5">
              <GameLogoBadge slug={game?.slug} name={game?.name} genre={game?.genre} size="lg" />
              <div>
                <div className="flex items-center gap-3">
                  <h1 className="text-xl md:text-2xl font-bold tracking-tight text-text-primary">
                    {tournament.title}
                  </h1>
                  <TournamentStatusBadge status={tournament.status} size="md" />
                </div>

                <div className="flex flex-wrap items-center gap-3 text-xs text-text-muted mt-1.5">
                  <span className="font-medium text-text-secondary">
                    {gameName}
                  </span>
                  <span>•</span>
                  <span className="font-mono uppercase tracking-wider text-[11px]">
                    {formatLabel}
                  </span>
                </div>
              </div>
            </div>

            {/* Public Link Action */}
            <div className="flex items-center gap-2 self-start md:self-auto">
              <Link
                href={`/t/${id}`}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2 px-3.5 py-2 rounded-lg bg-surface-elevated hover:bg-surface-base text-text-primary text-xs font-medium border border-border-default transition-colors shadow-xs"
              >
                <Radio size={14} className="text-emerald-400 animate-pulse" />
                <span>Halaman Publik Live</span>
                <Share2 size={13} className="text-text-muted" />
              </Link>
            </div>
          </div>
        </div>
      </header>

      {/* Navigation Tabs (Sticky) */}
      <TournamentTabNav tournamentId={id} />

      {/* Main Workspace Body */}
      <main className="flex-1 max-w-7xl mx-auto w-full p-4 md:p-8">
        {children}
      </main>
    </div>
  );
}
