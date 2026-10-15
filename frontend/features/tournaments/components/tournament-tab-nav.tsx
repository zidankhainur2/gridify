'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { LayoutGrid, Trophy, Users, Settings } from 'lucide-react';

interface TabNavProps {
  tournamentId: string;
}

export function TournamentTabNav({ tournamentId }: TabNavProps) {
  const pathname = usePathname();
  const basePath = `/dashboard/tournaments/${tournamentId}`;

  const navItems = [
    {
      label: 'Ringkasan',
      href: basePath,
      icon: LayoutGrid,
      exact: true,
    },
    {
      label: 'Bagan & Pertandingan',
      href: `${basePath}/bracket`,
      icon: Trophy,
      exact: false,
    },
    {
      label: 'Peserta',
      href: `${basePath}/participants`,
      icon: Users,
      exact: false,
    },
    {
      label: 'Pengaturan',
      href: `${basePath}/settings`,
      icon: Settings,
      exact: false,
    },
  ];

  return (
    <div className="border-b border-border-default bg-surface-raised sticky top-0 z-30">
      <div className="max-w-7xl mx-auto px-4 md:px-8 overflow-x-auto">
        <nav className="flex space-x-1 py-1">
          {navItems.map((item) => {
            const isActive = item.exact
              ? pathname === item.href
              : pathname?.startsWith(item.href);

            return (
              <Link
                key={item.href}
                href={item.href}
                className={`relative px-4 py-3 flex items-center gap-2 text-sm font-medium transition-colors whitespace-nowrap rounded-md ${
                  isActive
                    ? 'text-accent bg-surface-elevated font-semibold'
                    : 'text-text-muted hover:text-text-primary hover:bg-surface-base'
                }`}
              >
                <item.icon
                  size={16}
                  className={isActive ? 'text-accent' : 'text-text-muted'}
                />
                <span>{item.label}</span>
                {isActive && (
                  <span className="absolute bottom-0 left-2 right-2 h-0.5 bg-accent rounded-full" />
                )}
              </Link>
            );
          })}
        </nav>
      </div>
    </div>
  );
}
