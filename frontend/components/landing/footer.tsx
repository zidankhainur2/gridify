import Link from 'next/link';
import { Gamepad2 } from 'lucide-react';

export default function Footer() {
  return (
    <footer className="border-t border-border-default bg-surface-raised/40 py-14 px-4 md:px-8">
      <div className="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-12 gap-10">
        {/* Column 1: Brand & Overview (5 cols) */}
        <div className="md:col-span-5 space-y-4">
          <Link href="/" className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
              <Gamepad2 size={18} />
            </div>
            <span className="text-lg font-display font-bold tracking-tight text-text-primary">
              Gridify
            </span>
          </Link>
          <p className="text-xs text-text-muted leading-relaxed max-w-sm">
            Platform operasional turnamen esports realtime dengan generator bagan otomatis, kalkulasi klasemen deterministik, dan siaran langsung spectator view.
          </p>
          <div className="pt-2 text-[11px] font-mono text-text-muted">
            © {new Date().getFullYear()} Gridify. All rights reserved.
          </div>
        </div>

        {/* Column 2: Supported Formats (4 cols) */}
        <div className="md:col-span-4 space-y-3">
          <h4 className="text-xs font-mono font-bold uppercase tracking-wider text-text-primary">
            Format Kompetisi
          </h4>
          <ul className="space-y-2 text-xs text-text-secondary">
            <li>Single Elimination (Knockout Murni & Juara 3)</li>
            <li>Double Elimination (Upper & Lower Bracket)</li>
            <li>Liga Round Robin (1-Leg & 2-Leg Home-Away)</li>
            <li>Hybrid UCL (Babak Grup + Knockout Fase Akhir)</li>
            <li>Battle Royale Leaderboard (Placement & Kill Points)</li>
          </ul>
        </div>

        {/* Column 3: Internal Navigation (3 cols) */}
        <div className="md:col-span-3 space-y-3">
          <h4 className="text-xs font-mono font-bold uppercase tracking-wider text-text-primary">
            Navigasi
          </h4>
          <ul className="space-y-2 text-xs text-text-muted">
            <li>
              <Link href="#games" className="hover:text-text-primary transition-colors">
                Cabang Game
              </Link>
            </li>
            <li>
              <Link href="#features" className="hover:text-text-primary transition-colors">
                Fitur Unggulan
              </Link>
            </li>
            <li>
              <Link href="#how-it-works" className="hover:text-text-primary transition-colors">
                Cara Kerja
              </Link>
            </li>
            <li>
              <Link href="/login" className="hover:text-text-primary transition-colors">
                Masuk Organizer
              </Link>
            </li>
            <li>
              <Link href="/register" className="hover:text-text-primary transition-colors">
                Daftar Akun Baru
              </Link>
            </li>
          </ul>
        </div>
      </div>
    </footer>
  );
}
