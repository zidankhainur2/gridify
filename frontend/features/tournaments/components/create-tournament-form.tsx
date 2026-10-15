'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Trophy, Settings2, Info, Gamepad2, ArrowRight } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Game, TournamentFormat } from '@/types/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, Textarea } from '@/components/ui/select';

interface FormatOption {
  id: TournamentFormat;
  label: string;
  description: string;
  validGenres: string[];
}

const FORMAT_OPTIONS: FormatOption[] = [
  {
    id: 'SINGLE_ELIMINATION',
    label: 'Single Elimination',
    description: 'Sistem gugur sekali kalah langsung tereliminasi. Cepat & dramatis.',
    validGenres: ['MOBA', 'FPS', 'SPORTS', 'FIGHTING', 'BATTLE_ROYALE'],
  },
  {
    id: 'DOUBLE_ELIMINATION',
    label: 'Double Elimination',
    description: 'Upper & Lower Bracket. Peserta gugur setelah dua kali kekalahan.',
    validGenres: ['MOBA', 'FPS', 'SPORTS', 'FIGHTING'],
  },
  {
    id: 'ROUND_ROBIN',
    label: 'Round Robin / Liga',
    description: 'Semua peserta bertanding satu sama lain. Klasemen poin menentukan juara.',
    validGenres: ['MOBA', 'FPS', 'SPORTS', 'FIGHTING'],
  },
  {
    id: 'HYBRID_UCL',
    label: 'Hybrid / UCL (Grup + Gugur)',
    description: 'Fase Grup Klasemen (Round Robin) dilanjutkan ke Babak Knockout.',
    validGenres: ['MOBA', 'FPS', 'SPORTS'],
  },
  {
    id: 'BATTLE_ROYALE',
    label: 'Battle Royale',
    description: 'Pertandingan multi-tim serentak dalam beberapa putaran/match akumulasi poin.',
    validGenres: ['BATTLE_ROYALE'],
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

  const handleGameChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const nextGameId = e.target.value;
    setGameId(nextGameId);
    const game = games.find((g) => g.id === nextGameId);

    if (game?.genre === 'BATTLE_ROYALE') {
      setSelectedFormat('BATTLE_ROYALE');
    } else {
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

    const matchRules: Record<string, any> = {
      best_of: parseInt(bestOf, 10) || 1,
      home_away: homeAway,
      has_third_place: hasThirdPlace,
    };

    if (selectedFormat === 'BATTLE_ROYALE') {
      matchRules.total_matches = totalMatches;
    }

    try {
      const created = await apiClient.tournaments.create({
        title: title.trim(),
        game_id: gameId,
        format_type: selectedFormat as TournamentFormat,
        description: description.trim() || undefined,
        match_rules: matchRules,
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
      {/* 1. Informasi Dasar */}
      <div className="p-6 rounded-xl bg-surface-raised border border-border-default space-y-4">
        <div className="flex items-center gap-2.5 pb-3 border-b border-border-subtle">
          <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
            <Info size={16} />
          </div>
          <div>
            <h2 className="text-base font-semibold text-text-primary">
              Informasi Dasar
            </h2>
            <p className="text-xs text-text-muted">
              Judul kompetisi dan pilihan game esports
            </p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            label="Nama Turnamen"
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Contoh: Valorant National Cup Season 1"
            required
          />

          <Select
            label="Pilih Game"
            value={gameId}
            onChange={handleGameChange}
            options={[
              { value: '', label: '-- Pilih Game Esports --' },
              ...games.map((g) => ({
                value: g.id,
                label: `${g.name} (${g.genre} • ${g.platform})`,
              })),
            ]}
            required
          />
        </div>

        <Textarea
          label="Deskripsi & Aturan Singkat (Opsional)"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Tuliskan jadwal babak, hadiah turnamen, atau aturan kualifikasi..."
          rows={3}
        />
      </div>

      {/* 2. Format Turnamen */}
      <div className="p-6 rounded-xl bg-surface-raised border border-border-default space-y-4">
        <div className="flex items-center gap-2.5 pb-3 border-b border-border-subtle">
          <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
            <Trophy size={16} />
          </div>
          <div>
            <h2 className="text-base font-semibold text-text-primary">
              Format Kompetisi
            </h2>
            <p className="text-xs text-text-muted">
              Pilih bagan pertandingan yang sesuai dengan game
            </p>
          </div>
        </div>

        {!selectedGame ? (
          <div className="text-center py-8 rounded-lg border border-dashed border-border-default bg-surface-base text-text-muted text-sm">
            <Gamepad2 size={24} className="mx-auto mb-2 opacity-50" />
            Silakan pilih Game di bagian Informasi Dasar terlebih dahulu.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {availableFormats.map((format) => {
              const isSelected = selectedFormat === format.id;
              return (
                <div
                  key={format.id}
                  onClick={() => setSelectedFormat(format.id)}
                  className={`p-4 rounded-xl border cursor-pointer transition-colors ${
                    isSelected
                      ? 'bg-accent/10 border-accent'
                      : 'bg-surface-base border-border-subtle hover:border-border-default'
                  }`}
                >
                  <div className="flex items-center gap-3 mb-1.5">
                    <div
                      className={`w-4 h-4 rounded-full border flex items-center justify-center shrink-0 ${
                        isSelected ? 'border-accent' : 'border-border-default'
                      }`}
                    >
                      {isSelected && (
                        <div className="w-2 h-2 rounded-full bg-accent" />
                      )}
                    </div>
                    <span className="font-semibold text-sm text-text-primary">
                      {format.label}
                    </span>
                  </div>
                  <p className="text-xs text-text-muted ml-7 leading-relaxed">
                    {format.description}
                  </p>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* 3. Konfigurasi Match Rules */}
      {selectedGame && selectedFormat && (
        <div className="p-6 rounded-xl bg-surface-raised border border-border-default space-y-4">
          <div className="flex items-center gap-2.5 pb-3 border-b border-border-subtle">
            <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
              <Settings2 size={16} />
            </div>
            <div>
              <h2 className="text-base font-semibold text-text-primary">
                Aturan Pertandingan
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
                  ].map((bo) => (
                    <button
                      key={bo.val}
                      type="button"
                      onClick={() => setBestOf(bo.val)}
                      className={`flex-1 py-2 px-3 rounded-lg border text-xs font-mono font-medium transition-colors ${
                        bestOf === bo.val
                          ? 'bg-accent/10 border-accent text-accent'
                          : 'bg-surface-base border-border-subtle text-text-muted hover:text-text-primary'
                      }`}
                    >
                      {bo.label}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Bronze Match Juara 3 */}
            {selectedFormat === 'SINGLE_ELIMINATION' && (
              <label className="flex items-center justify-between p-3.5 rounded-lg border border-border-subtle bg-surface-base cursor-pointer hover:border-border-default transition-colors">
                <div>
                  <span className="font-medium text-sm text-text-primary block">
                    Bronze Match (Perebutan Juara 3)
                  </span>
                  <span className="text-xs text-text-muted">
                    Adakan pertandingan antara kedua tim yang kalah di Semifinal
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
              <label className="flex items-center justify-between p-3.5 rounded-lg border border-border-subtle bg-surface-base cursor-pointer hover:border-border-default transition-colors">
                <div>
                  <span className="font-medium text-sm text-text-primary block">
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
                  {[3, 4, 5, 6, 8].map((num) => (
                    <button
                      key={num}
                      type="button"
                      onClick={() => setTotalMatches(num)}
                      className={`flex-1 py-2 px-3 rounded-lg border text-xs font-mono font-medium transition-colors ${
                        totalMatches === num
                          ? 'bg-accent/10 border-accent text-accent'
                          : 'bg-surface-base border-border-subtle text-text-muted hover:text-text-primary'
                      }`}
                    >
                      {num} Round
                    </button>
                  ))}
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
