'use client';

import {
  Zap,
  Swords,
  Users,
  ShieldCheck,
  Share2,
  BarChart3,
} from 'lucide-react';
import { motion } from 'framer-motion';

const features = [
  {
    icon: Zap,
    title: 'Pembaruan Realtime SSE',
    desc: 'Skor dan progresi bagan diperbarui seketika di layar penonton tanpa refresh melalui koneksi Server-Sent Events berkecepatan tinggi.',
  },
  {
    icon: Swords,
    title: '5 Format Turnamen Lengkap',
    desc: 'Mendukung Single Elimination, Double Elimination, Liga Round Robin (1 & 2-Leg), Hybrid UCL, dan Battle Royale multi-ronde.',
  },
  {
    icon: Users,
    title: 'Distribusi & Seeding Otomatis',
    desc: 'Algoritma seeding deterministik pangkat-dua, auto-advance BYE, dan pembagian grup UCL yang teratur dan adil.',
  },
  {
    icon: BarChart3,
    title: 'Klasemen & Leaderboard Sentral',
    desc: 'Perhitungan otomatis poin kemenangan, selisih gol, head-to-head liga, serta kalkulasi WWCD dan kill points Battle Royale.',
  },
  {
    icon: ShieldCheck,
    title: 'Keamanan Data & Integritas Skor',
    desc: 'Seluruh riwayat skor dan bracket tersimpan dengan transaksi database atomik dan proteksi otorisasi penuh penyelenggara.',
  },
  {
    icon: Share2,
    title: 'Halaman Publik Penonton',
    desc: 'Tautan shareable publik untuk penonton dengan bagan visual SVG presisi tinggi yang responsif di smartphone maupun desktop.',
  },
];

export default function Features() {
  return (
    <section
      id="features"
      className="py-20 px-4 md:px-8 relative bg-surface-raised/40 border-y border-border-default"
    >
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-14">
          <h2 className="text-2xl sm:text-3xl font-display font-bold text-text-primary mb-3 tracking-tight">
            Fitur Operasi Turnamen Kelas Kompetitif
          </h2>
          <p className="text-sm text-text-secondary max-w-2xl mx-auto leading-relaxed">
            Dirancang khusus untuk mempermudah operasional komunitas gamer, esports arena, dan organizer kompetisi profesional.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {features.map((feature, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 15 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, delay: idx * 0.05 }}
              viewport={{ once: true }}
              className="p-6 rounded-xl bg-surface-raised border border-border-default hover:border-border-strong hover:bg-surface-elevated/70 transition-all group shadow-xs"
            >
              <div className="w-10 h-10 rounded-lg bg-surface-elevated border border-border-subtle flex items-center justify-center mb-4 text-accent group-hover:scale-105 transition-transform">
                <feature.icon className="h-5 w-5" />
              </div>
              <h3 className="text-base font-display font-semibold text-text-primary mb-2">
                {feature.title}
              </h3>
              <p className="text-xs text-text-muted leading-relaxed">
                {feature.desc}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
