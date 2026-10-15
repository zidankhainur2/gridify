'use client';

import Link from 'next/link';
import { Trophy, LayoutDashboard } from 'lucide-react';
import { motion } from 'framer-motion';
import { User } from '@/types/api';

export default function Navbar({ user }: { user: User | null }) {
  return (
    <motion.nav
      initial={{ y: -60 }}
      animate={{ y: 0 }}
      transition={{ duration: 0.3 }}
      className="fixed top-0 w-full z-50 border-b border-[#27272A] bg-[#09090B]/90 backdrop-blur-md"
    >
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2.5 group">
          <div className="w-8 h-8 rounded-lg bg-zinc-100 flex items-center justify-center text-zinc-950 transition-transform group-hover:scale-105 shadow-sm">
            <Trophy className="h-4 w-4" />
          </div>
          <span className="text-lg font-bold tracking-tight text-zinc-100">
            Gridify
          </span>
        </Link>

        {/* Desktop Menu */}
        <div className="hidden md:flex items-center gap-8 text-sm font-medium text-zinc-400">
          <Link href="#features" className="hover:text-zinc-100 transition-colors">
            Fitur
          </Link>
          <Link href="#how-it-works" className="hover:text-zinc-100 transition-colors">
            Cara Kerja
          </Link>
          <Link href="/login" className="hover:text-zinc-100 transition-colors">
            Bagan Publik
          </Link>
        </div>

        {/* Auth Buttons */}
        <div className="flex items-center gap-3">
          {user ? (
            <div className="flex items-center gap-4">
              <div className="hidden md:flex flex-col items-end text-right">
                <span className="text-[11px] text-zinc-500 font-medium">Masuk sebagai</span>
                <span className="text-xs font-semibold text-zinc-200 max-w-[140px] truncate">
                  {user.username || user.email}
                </span>
              </div>

              <Link
                href="/dashboard"
                className="flex items-center gap-2 px-3.5 py-1.5 rounded-lg bg-zinc-100 text-zinc-950 text-xs font-semibold hover:bg-zinc-200 transition-colors shadow-sm"
              >
                <LayoutDashboard className="h-3.5 w-3.5" />
                <span>Dashboard</span>
              </Link>
            </div>
          ) : (
            <>
              <Link
                href="/login"
                className="text-xs font-medium text-zinc-300 hover:text-white px-3 py-1.5 transition-colors"
              >
                Masuk
              </Link>
              <Link
                href="/register"
                className="px-3.5 py-1.5 rounded-lg bg-zinc-100 text-zinc-950 text-xs font-semibold hover:bg-zinc-200 transition-colors shadow-sm"
              >
                Daftar Gratis
              </Link>
            </>
          )}
        </div>
      </div>
    </motion.nav>
  );
}
