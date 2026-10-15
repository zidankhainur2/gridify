'use client';

import * as React from 'react';
import Image from 'next/image';
import { Swords, Crosshair, Trophy, Flame } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export interface GameMeta {
  slug: string;
  name: string;
  genre: 'MOBA' | 'FPS' | 'SPORTS' | 'BATTLE_ROYALE';
  platform: 'PC' | 'Mobile' | 'Multiplatform' | 'PC/Console';
  color: string;
}

export const SUPPORTED_GAMES_META: GameMeta[] = [
  {
    slug: 'mobile-legends',
    name: 'Mobile Legends: Bang Bang',
    genre: 'MOBA',
    platform: 'Mobile',
    color: '#F59E0B',
  },
  {
    slug: 'dota-2',
    name: 'Dota 2',
    genre: 'MOBA',
    platform: 'PC',
    color: '#EF4444',
  },
  {
    slug: 'league-of-legends',
    name: 'League of Legends',
    genre: 'MOBA',
    platform: 'PC',
    color: '#CA8A04',
  },
  {
    slug: 'valorant',
    name: 'Valorant',
    genre: 'FPS',
    platform: 'PC',
    color: '#FF4655',
  },
  {
    slug: 'counter-strike-2',
    name: 'Counter-Strike 2',
    genre: 'FPS',
    platform: 'PC',
    color: '#F59E0B',
  },
  {
    slug: 'ea-sports-fc',
    name: 'EA Sports FC 24',
    genre: 'SPORTS',
    platform: 'Multiplatform',
    color: '#10B981',
  },
  {
    slug: 'efootball',
    name: 'eFootball 2024',
    genre: 'SPORTS',
    platform: 'Multiplatform',
    color: '#38BDF8',
  },
  {
    slug: 'pubg-mobile',
    name: 'PUBG Mobile',
    genre: 'BATTLE_ROYALE',
    platform: 'Mobile',
    color: '#FBBF24',
  },
  {
    slug: 'free-fire',
    name: 'Free Fire',
    genre: 'BATTLE_ROYALE',
    platform: 'Mobile',
    color: '#F97316',
  },
  {
    slug: 'apex-legends',
    name: 'Apex Legends',
    genre: 'BATTLE_ROYALE',
    platform: 'PC/Console',
    color: '#DC2626',
  },
];

export interface GameLogoBadgeProps {
  slug?: string;
  name?: string;
  gameName?: string;
  genre?: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  className?: string;
  showFallbackOnly?: boolean;
}

export function GameLogoBadge({
  slug,
  name,
  gameName,
  genre,
  size = 'md',
  className,
}: GameLogoBadgeProps) {
  const [hasError, setHasError] = React.useState(false);
  const effectiveName = name || gameName;

  // Normalize slug if provided or infer from name
  const effectiveSlug = React.useMemo(() => {
    if (slug) return slug;
    if (!effectiveName) return '';
    const found = SUPPORTED_GAMES_META.find(
      (g) => g.name.toLowerCase() === effectiveName.toLowerCase()
    );
    if (found) return found.slug;
    return effectiveName
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '');
  }, [slug, effectiveName]);

  const meta = SUPPORTED_GAMES_META.find((g) => g.slug === effectiveSlug);
  const effectiveGenre = genre || meta?.genre || 'MOBA';

  const sizeDimensions = {
    xs: { box: 'h-6 w-6 rounded-md', img: 16, iconSize: 12, text: 'text-[9px]' },
    sm: { box: 'h-8 w-8 rounded-lg', img: 20, iconSize: 14, text: 'text-[10px]' },
    md: { box: 'h-10 w-10 rounded-xl', img: 26, iconSize: 18, text: 'text-xs' },
    lg: { box: 'h-12 w-12 rounded-xl', img: 32, iconSize: 22, text: 'text-sm' },
    xl: { box: 'h-16 w-16 rounded-2xl', img: 44, iconSize: 28, text: 'text-base' },
  }[size];

  // Fallback emblem component based on genre
  const renderFallback = () => {
    const initials = (name || meta?.name || 'G')
      .split(' ')
      .map((w) => w[0])
      .slice(0, 2)
      .join('')
      .toUpperCase();

    const genreIcons = {
      MOBA: Swords,
      FPS: Crosshair,
      SPORTS: Trophy,
      BATTLE_ROYALE: Flame,
    };
    const IconComp = genreIcons[effectiveGenre as keyof typeof genreIcons] || Swords;

    return (
      <div
        className={twMerge(
          clsx(
            'flex items-center justify-center shrink-0 border border-border-default bg-surface-elevated text-text-primary font-bold shadow-xs select-none',
            sizeDimensions.box,
            className
          )
        )}
        title={name || meta?.name || 'Game'}
      >
        {size === 'xs' || size === 'sm' ? (
          <span className={clsx('font-mono font-bold leading-none', sizeDimensions.text)}>
            {initials}
          </span>
        ) : (
          <div className="flex flex-col items-center justify-center">
            <IconComp size={sizeDimensions.iconSize} className="text-accent" />
          </div>
        )}
      </div>
    );
  };

  if (!effectiveSlug || hasError) {
    return renderFallback();
  }

  return (
    <div
      className={twMerge(
        clsx(
          'relative flex items-center justify-center shrink-0 overflow-hidden border border-border-default bg-surface-raised shadow-xs',
          sizeDimensions.box,
          className
        )
      )}
      title={name || meta?.name || effectiveSlug}
    >
      <Image
        src={`/games/${effectiveSlug}.svg`}
        alt={name || meta?.name || effectiveSlug}
        width={sizeDimensions.img}
        height={sizeDimensions.img}
        className="object-contain"
        onError={() => setHasError(true)}
      />
    </div>
  );
}

export default GameLogoBadge;
