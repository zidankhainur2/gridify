'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Trophy, Settings2, Info, Gamepad2, ArrowRight, Check, GitBranch, GitFork, Table2, Crosshair } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Game, TournamentFormat } from '@/types/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/select';
import { GameLogoBadge } from '@/components/ui/game-logo';

interface FormatOption {
  id: TournamentFormat;
  label: string;
  description: string;
  validGenres: string[];
  icon: typeof GitBranch;
}

const FORMAT_OPTIONS: FormatOption[] = [
  {
    id: 'SINGLE_ELIMINATION',
    label: 'Single Elimination',
    description: 'Sistem gugur murni: sekali kalah langsung tereliminasi. Cepat & dramatis.',
    validGenres: ['MOBA', 'FPS', 'SPORTS'],
    icon: GitBranch,
  },
  {
    id: 'DOUBLE_ELIMINATION',
    label: 'Double Elimination',
    description: 'Upper & Lower Bracket: peserta baru gugur setelah dua kali kekalahan.',
    validGenres: ['MOBA', 'FPS', 'SPORTS'],
    icon: GitFork,
  },
  {
    id: 'ROUND_ROBIN',
    label: 'Round Robin / Liga',
    description: 'Setiap peserta bertanding melawan seluruh tim lain. Klasemen poin menentukan juara.',
    validGenres: ['MOBA', 'FPS', 'SPORTS'],
    icon: Table2,
  },
  {
    id: 'HYBRID_UCL',
    label: 'Hybrid / UCL (Grup + Gugur)',
    description: 'Babak 1 Fase Grup (Round Robin) dilanjutkan ke Babak Knockout Playoff.',
    validGenres: ['MOBA', 'FPS', 'SPORTS'],
    icon: Trophy,
  },
  {
    id: 'BATTLE_ROYALE',
    label: 'Battle Royale',
    description: 'Pertandingan multi-tim serentak dalam beberapa putaran/match dengan akumulasi poin.',
    validGenres: ['BATTLE_ROYALE'],
    icon: Crosshair,
  },
];

export default function CreateTournamentForm({ games }: { games: Game[] }) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  const [title, setTitle] = useState('');
  const [gameId, setGameId] = useState('');
  const [description, setDescription] = useState('');
  const [selectedFormat, setSelectedFormat] = useState<TournamentFormat | ''>('');

  // Rules / Config state
  const [bestOf, setBestOf] = useState('1');
  const [homeAway, setHomeAway] = useState(false);
  const [hasThirdPlace, setHasThirdPlace] = useState(false);
  const [totalMatches, setTotalMatches] = useState(5);

  const selectedGame = games.find((g) => g.id === gameId);

  const availableFormats = selectedGame
    ? FORMAT_OPTIONS.filter((f) => f.validGenres.includes(selectedGame.genre))
    : [];

  const handleGameSelect = (nextGameId: string) => {
    setGameId(nextGameId);
    const game = games.find((g) => g.id === nextGameId);

    if (game?.genre === 'BATTLE_ROYALE') {
      setSelectedFormat('BATTLE_ROYALE');
    } else if (selectedFormat === 'BATTLE_ROYALE') {
      setSelectedFormat('');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!title.trim() || !gameId || !selectedFormat) {
      toast.error('Harap lengkapi judul, game, dan format turnamen');
      return;
    }

    setLoading(true);

    try {
      const created = await apiClient.tournaments.create({
        title: title.trim(),
        game_id: gameId,
        format_type: selectedFormat as TournamentFormat,
        description: description.trim() || undefined,
        best_of: parseInt(bestOf, 10) || 1,
        home_away: homeAway,
        has_third_place: hasThirdPlace,
        br_total_matches: selectedFormat === 'BATTLE_ROYALE' ? totalMatches : undefined,
      });

      toast.success('Turnamen Berhasil Dibuat!', {
        description: `Turnamen "${created.title}" siap dikelola.`,
      });

      router.push(`/dashboard/tournaments/${created.id}`);
      router.refresh();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal membuat turnamen.';
      toast.error('Gagal Membuat Turnamen', { description: msg });
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* 1. Pilih Cabang Game Resmi */}
      <div className="p-6 rounded-xl bg-surface-raised border border-border-default space-y-4">
        <div className="flex items-center gap-2.5 pb-3 border-b border-border-subtle">
          <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
            <Gamepad2 size={16} />
          </div>
          <div>
            <h2 className="text-base font-semibold text-text-primary">
              1. Pilih Cabang Game Esports
            </h2>
            <p className="text-xs text-text-muted">
              Pilih cabang game resmi yang akan dikompetisikan
            </p>
          </div>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3" role="radiogroup" aria-label="Pilihan Game">
          {games.map((g) => {
            const isSelected = gameId === g.id;
            return (
              <button
                key={g.id}
                type="button"
                role="radio"
                aria-checked={isSelected}
                onClick={() => handleGameSelect(g.id)}
                className={`flex flex-col items-center text-center p-3 rounded-xl border transition-all cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent ${
                  isSelected
                    ? 'border-2 border-accent bg-accent/12 shadow-xs'
                    : 'border-border-default bg-surface-base hover:border-border-strong hover:bg-surface-elevated/70'
                }`}
              >
                <div className="relative mb-2">
                  <GameLogoBadge slug={g.slug} name={g.name} genre={g.genre} size="md" />
                  {isSelected && (
                    <div className="absolute -top-1.5 -right-1.5 w-4 h-4 rounded-full bg-accent text-white flex items-center justify-center shadow-xs">
                      <Check size={10} strokeWidth={3} />
                    </div>
                  )}
                </div>
                <span className="font-semibold text-xs text-text-primary line-clamp-1">
                  {g.name}
                </span>
                <span className="text-[10px] font-mono text-text-muted uppercase mt-0.5">
                  {g.genre} • {g.platform}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {/* 2. Informasi Dasar */}
      <div className="p-6 rounded-xl bg-surface-raised border border-border-default space-y-4">
        <div className="flex items-center gap-2.5 pb-3 border-b border-border-subtle">
          <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
            <Info size={16} />
          </div>
          <div>
            <h2 className="text-base font-semibold text-text-primary">
              2. Informasi Turnamen
            </h2>
            <p className="text-xs text-text-muted">
              Judul kompetisi dan deskripsi panduan peserta
            </p>
          </div>
        </div>

        <div className="space-y-4">
          <Input
            label="Nama Turnamen"
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Contoh: Valorant Champions Community Season 1"
            required
          />

          <Textarea
            label="Deskripsi & Aturan Singkat (Opsional)"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Tuliskan jadwal babak, hadiah turnamen, atau aturan kualifikasi..."
            rows={3}
          />
        </div>
      </div>

      {/* 3. Format Turnamen */}
      <div className="p-6 rounded-xl bg-surface-raised border border-border-default space-y-4">
        <div className="flex items-center gap-2.5 pb-3 border-b border-border-subtle">
          <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
            <Trophy size={16} />
          </div>
          <div>
            <h2 className="text-base font-semibold text-text-primary">
              3. Format Kompetisi
            </h2>
            <p className="text-xs text-text-muted">
              Bagan pertandingan yang kompatibel dengan cabang game terpilih
            </p>
          </div>
        </div>

        {!selectedGame ? (
          <div className="text-center py-8 rounded-lg border border-dashed border-border-default bg-surface-base text-text-muted text-sm">
            <Gamepad2 size={24} className="mx-auto mb-2 opacity-50" />
            Silakan pilih Game di atas terlebih dahulu.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3" role="radiogroup" aria-label="Format Kompetisi">
            {availableFormats.map((format) => {
              const isSelected = selectedFormat === format.id;
              const IconComp = format.icon;

              return (
                <button
                  key={format.id}
                  type="button"
                  role="radio"
                  aria-checked={isSelected}
                  onClick={() => setSelectedFormat(format.id)}
                  className={`p-4 rounded-xl text-left cursor-pointer transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent ${
                    isSelected
                      ? 'border-2 border-accent bg-accent/12 shadow-sm ring-1 ring-accent/30'
                      : 'border border-border-default bg-surface-base hover:border-border-strong hover:bg-surface-elevated/70'
                  }`}
                >
                  <div className="flex items-start justify-between gap-3 mb-2">
                    <div className="flex items-center gap-2.5">
                      <div className={`p-2 rounded-lg border ${isSelected ? 'bg-accent/20 border-accent/40 text-accent' : 'bg-surface-elevated border-border-default text-text-muted'}`}>
                        <IconComp size={18} />
                      </div>
                      <span className="font-bold text-sm text-text-primary">
                        {format.label}
                      </span>
                    </div>

                    {isSelected ? (
                      <span className="text-[10px] font-mono font-bold uppercase tracking-wider bg-accent text-white px-2 py-0.5 rounded shadow-xs flex items-center gap-1">
                        <Check size={11} strokeWidth={3} />
                        AKTIF
                      </span>
                    ) : (
                      <div className="w-4 h-4 rounded-full border border-border-default shrink-0 mt-1" />
                    )}
                  </div>
                  <p className="text-xs text-text-secondary leading-relaxed">
                    {format.description}
                  </p>
                </button>
              );
            })}
          </div>
        )}
      </div>

      {/* 4. Konfigurasi Match Rules */}
      {selectedGame && selectedFormat && (
        <div className="p-6 rounded-xl bg-surface-raised border border-border-default space-y-4">
          <div className="flex items-center gap-2.5 pb-3 border-b border-border-subtle">
            <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
              <Settings2 size={16} />
            </div>
            <div>
              <h2 className="text-base font-semibold text-text-primary">
                4. Aturan Pertandingan
              </h2>
              <p className="text-xs text-text-muted">
                Parameter seri Best-Of, perebutan juara 3, atau konfigurasi match
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Opsi Best-Of */}
            {selectedFormat !== 'BATTLE_ROYALE' && (
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-secondary uppercase tracking-wider">
                  Format Seri (Best Of)
                </label>
                <div className="flex gap-2">
                  {[
                    { val: '1', label: 'BO1 (1 Game)' },
                    { val: '3', label: 'BO3 (Best of 3)' },
                    { val: '5', label: 'BO5 (Best of 5)' },
                  ].map((bo) => {
                    const isSelected = bestOf === bo.val;
                    return (
                      <button
                        key={bo.val}
                        type="button"
                        onClick={() => setBestOf(bo.val)}
                        className={`flex-1 py-2.5 px-3 rounded-lg border text-xs font-mono font-bold transition-all cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent ${
                          isSelected
                            ? 'border-2 border-accent bg-accent/15 text-text-primary shadow-xs'
                            : 'border-border-default bg-surface-base text-text-muted hover:border-border-strong hover:text-text-primary'
                        }`}
                      >
                        {bo.label}
                      </button>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Bronze Match Juara 3 */}
            {selectedFormat === 'SINGLE_ELIMINATION' && (
              <label className="flex items-center justify-between p-3.5 rounded-lg border border-border-default bg-surface-base cursor-pointer hover:border-border-strong transition-colors">
                <div>
                  <span className="font-semibold text-sm text-text-primary block">
                    Bronze Match (Perebutan Juara 3)
                  </span>
                  <span className="text-xs text-text-muted">
                    Adakan pertandingan antara kedua tim yang gugur di Semifinal
                  </span>
                </div>
                <input
                  type="checkbox"
                  checked={hasThirdPlace}
                  onChange={(e) => setHasThirdPlace(e.target.checked)}
                  className="w-4 h-4 rounded border-border-default text-accent focus:ring-accent bg-surface-raised ml-3"
                />
              </label>
            )}

            {/* Home & Away */}
            {(selectedFormat === 'ROUND_ROBIN' ||
              selectedFormat === 'HYBRID_UCL') && (
              <label className="flex items-center justify-between p-3.5 rounded-lg border border-border-default bg-surface-base cursor-pointer hover:border-border-strong transition-colors">
                <div>
                  <span className="font-semibold text-sm text-text-primary block">
                    Format Home & Away (2 Leg)
                  </span>
                  <span className="text-xs text-text-muted">
                    Setiap tim bertanding dua kali (kandang & tandang)
                  </span>
                </div>
                <input
                  type="checkbox"
                  checked={homeAway}
                  onChange={(e) => setHomeAway(e.target.checked)}
                  className="w-4 h-4 rounded border-border-default text-accent focus:ring-accent bg-surface-raised ml-3"
                />
              </label>
            )}

            {/* Battle Royale Match Config */}
            {selectedFormat === 'BATTLE_ROYALE' && (
              <div className="space-y-1.5 md:col-span-2">
                <label className="text-xs font-medium text-text-secondary uppercase tracking-wider">
                  Total Putaran Match Battle Royale
                </label>
                <div className="flex gap-2">
                  {[3, 4, 5, 6, 8].map((num) => {
                    const isSelected = totalMatches === num;
                    return (
                      <button
                        key={num}
                        type="button"
                        onClick={() => setTotalMatches(num)}
                        className={`flex-1 py-2.5 px-3 rounded-lg border text-xs font-mono font-bold transition-all cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent ${
                          isSelected
                            ? 'border-2 border-accent bg-accent/15 text-text-primary shadow-xs'
                            : 'border-border-default bg-surface-base text-text-muted hover:border-border-strong hover:text-text-primary'
                        }`}
                      >
                        {num} Round
                      </button>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Submit Button */}
      <div className="flex justify-end pt-2">
        <Button
          type="submit"
          variant="primary"
          size="lg"
          isLoading={loading}
          disabled={!selectedFormat}
        >
          <span>Buat Turnamen Sekarang</span>
          <ArrowRight size={18} />
        </Button>
      </div>
    </form>
  );
}
