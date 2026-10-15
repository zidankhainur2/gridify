'use client';

import * as React from 'react';
import Link from 'next/link';
import { Trophy, ArrowRight, Gamepad2, ChevronRight, Check } from 'lucide-react';
import { motion } from 'framer-motion';
import { BracketPreview } from './bracket-preview';
import {
  GameLogoBadge,
  SUPPORTED_GAMES_META,
  GameMeta,
} from '@/components/ui/game-logo';

const GENRE_FILTERS = ['Semua', 'MOBA', 'FPS', 'SPORTS', 'BATTLE_ROYALE'] as const;

export default function Hero() {
  const [selectedGenre, setSelectedGenre] = React.useState<string>('Semua');
  const [selectedGameSlug, setSelectedGameSlug] = React.useState<string>('valorant');

  const filteredGames = React.useMemo(() => {
    if (selectedGenre === 'Semua') return SUPPORTED_GAMES_META;
    return SUPPORTED_GAMES_META.filter((g) => g.genre === selectedGenre);
  }, [selectedGenre]);

  const activeGame = React.useMemo(() => {
    return (
      SUPPORTED_GAMES_META.find((g) => g.slug === selectedGameSlug) ||
      SUPPORTED_GAMES_META[0]
    );
  }, [selectedGameSlug]);

  return (
    <section className="relative pt-28 pb-16 md:pt-36 md:pb-24 px-4 md:px-8 overflow-hidden">
      <div className="max-w-6xl mx-auto text-center relative z-10">
        {/* Subtle Category Pill */}
        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
          className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-surface-raised border border-border-default text-text-secondary text-xs font-mono mb-6 select-none shadow-xs"
        >
          <Gamepad2 className="h-3.5 w-3.5 text-accent" />
          <span>Esports Tournament Operations Platform</span>
        </motion.div>

        {/* Headline */}
        <motion.h1
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.1 }}
          className="text-3xl sm:text-5xl md:text-6xl font-display font-black text-text-primary tracking-tight mb-5 leading-tight"
        >
          Sistem Operasi Turnamen Esports <br />
          <span className="text-text-muted font-medium text-2xl sm:text-4xl md:text-5xl">
            Cepat, Deterministik & Realtime.
          </span>
        </motion.h1>

        {/* Subhead */}
        <motion.p
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.2 }}
          className="text-sm sm:text-base md:text-lg text-text-secondary max-w-2xl mx-auto mb-8 leading-relaxed font-sans"
        >
          Generator bagan profesional untuk Single Elimination, Double Elimination,
          Liga Round Robin, Hybrid UCL, dan Battle Royale dengan integrasi live spectator tanpa hambatan.
        </motion.p>

        {/* CTA Buttons */}
        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.3 }}
          className="flex flex-col sm:flex-row items-center justify-center gap-3.5 mb-14"
        >
          <Link
            href="/dashboard/create"
            className="w-full sm:w-auto px-7 py-3.5 rounded-xl bg-accent text-accent-contrast font-bold text-sm hover:bg-accent-hover transition-all shadow-md shadow-accent/20 flex items-center justify-center gap-2 hover:scale-[1.02] active:scale-[0.98]"
          >
            <Trophy className="h-4 w-4" />
            Buat Turnamen Sekarang
          </Link>
          <Link
            href="/login"
            className="w-full sm:w-auto px-6 py-3.5 rounded-xl bg-surface-raised border border-border-default text-text-primary font-semibold text-sm hover:bg-surface-elevated hover:border-border-strong transition-all flex items-center justify-center gap-2 group"
          >
            Masuk ke Organizer
            <ArrowRight className="h-4 w-4 group-hover:translate-x-0.5 transition-transform text-accent" />
          </Link>
        </motion.div>

        {/* 10 Supported Games Showcase Roster */}
        <motion.div
          id="games"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.35 }}
          className="pt-6 pb-2"
        >
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-5 px-2">
            <div className="text-left">
              <h2 className="text-lg md:text-xl font-display font-bold text-text-primary flex items-center gap-2">
                <span>10 Cabang Esports Resmi</span>
                <span className="text-xs font-mono font-normal text-text-muted px-2 py-0.5 rounded bg-surface-elevated border border-border-subtle">
                  Verified Meta
                </span>
              </h2>
              <p className="text-xs text-text-muted mt-0.5">
                Pilih cabang game untuk melihat simulasi pratinjau bagan di bawah:
              </p>
            </div>

            {/* Genre Filter Tabs */}
            <div className="flex items-center gap-1.5 overflow-x-auto pb-1 self-start sm:self-auto">
              {GENRE_FILTERS.map((genre) => (
                <button
                  key={genre}
                  type="button"
                  onClick={() => setSelectedGenre(genre)}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium font-mono transition-all whitespace-nowrap ${
                    selectedGenre === genre
                      ? 'bg-accent text-accent-contrast font-bold shadow-xs'
                      : 'bg-surface-raised text-text-muted hover:text-text-primary hover:bg-surface-elevated border border-border-subtle'
                  }`}
                >
                  {genre === 'BATTLE_ROYALE' ? 'Battle Royale' : genre}
                </button>
              ))}
            </div>
          </div>

          {/* Games Card Grid */}
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3 mb-6">
            {filteredGames.map((game) => {
              const isSelected = activeGame.slug === game.slug;
              return (
                <button
                  key={game.slug}
                  type="button"
                  onClick={() => setSelectedGameSlug(game.slug)}
                  className={`flex items-center gap-3 p-3 rounded-xl text-left transition-all border ${
                    isSelected
                      ? 'bg-surface-elevated border-accent shadow-sm shadow-accent/10 ring-1 ring-accent'
                      : 'bg-surface-raised/70 border-border-default hover:border-border-strong hover:bg-surface-elevated/70'
                  }`}
                >
                  <GameLogoBadge
                    slug={game.slug}
                    name={game.name}
                    genre={game.genre}
                    size="md"
                    className="shrink-0"
                  />
                  <div className="min-w-0 flex-1">
                    <div className="font-semibold text-xs text-text-primary truncate">
                      {game.name}
                    </div>
                    <div className="text-[10px] font-mono text-text-muted truncate">
                      {game.genre} • {game.platform}
                    </div>
                  </div>
                  {isSelected && (
                    <Check className="h-3.5 w-3.5 text-accent shrink-0" />
                  )}
                </button>
              );
            })}
          </div>
        </motion.div>

        {/* Interactive Live Bracket Component connected to Selected Game */}
        <motion.div
          initial={{ opacity: 0, y: 25 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.4 }}
        >
          <BracketPreview selectedGame={activeGame} />
        </motion.div>
      </div>
    </section>
  );
}
