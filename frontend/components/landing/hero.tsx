'use client';

import Link from 'next/link';
import { Trophy, ArrowRight, Activity } from 'lucide-react';
import { motion } from 'framer-motion';
import { BracketPreview } from './bracket-preview';

export default function Hero() {
  return (
    <section className="relative pt-28 pb-16 md:pt-36 md:pb-24 px-6 overflow-hidden">
      <div className="max-w-5xl mx-auto text-center relative z-10">
        {/* Badge */}
        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
          className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-[#18181B] border border-[#27272A] text-zinc-300 text-xs font-medium mb-6 shadow-xs select-none"
        >
          <Activity className="h-3.5 w-3.5 text-emerald-400" />
          <span>Arsitektur Baru: Next.js 16 + Go + Realtime SSE</span>
        </motion.div>

        {/* Headline */}
        <motion.h1
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.1 }}
          className="text-4xl sm:text-5xl md:text-6xl font-bold text-zinc-100 tracking-tight mb-5 leading-tight"
        >
          Sistem Operasi Turnamen Esports <br />
          <span className="text-zinc-400 font-normal">
            Cepat, Deterministik & Tanpa Hambatan.
          </span>
        </motion.h1>

        {/* Subhead */}
        <motion.p
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.2 }}
          className="text-base sm:text-lg text-zinc-400 max-w-2xl mx-auto mb-8 leading-relaxed"
        >
          Generator bagan profesional untuk Single Elimination, Double Elimination,
          Liga Round Robin, Hybrid UCL, dan Battle Royale dengan integrasi streaming langsung.
        </motion.p>

        {/* CTA Buttons */}
        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.3 }}
          className="flex flex-col sm:flex-row items-center justify-center gap-3.5"
        >
          <Link
            href="/dashboard/create"
            className="w-full sm:w-auto px-6 py-3 rounded-lg bg-zinc-100 text-zinc-950 font-semibold text-sm hover:bg-zinc-200 transition-colors shadow-sm flex items-center justify-center gap-2"
          >
            <Trophy className="h-4 w-4" />
            Buat Turnamen Sekarang
          </Link>
          <Link
            href="/login"
            className="w-full sm:w-auto px-6 py-3 rounded-lg bg-[#18181B] border border-[#27272A] text-zinc-200 font-semibold text-sm hover:bg-[#27272A] transition-colors flex items-center justify-center gap-2 group"
          >
            Masuk ke Dashboard
            <ArrowRight className="h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
          </Link>
        </motion.div>

        {/* Interactive Live Bracket Component */}
        <motion.div
          initial={{ opacity: 0, y: 25 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.4 }}
        >
          <BracketPreview />
        </motion.div>
      </div>
    </section>
  );
}
