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
    desc: 'Skor dan progresi bagan diperbarui instan di layar penonton tanpa refresh melalui Server-Sent Events dari Go engine.',
  },
  {
    icon: Swords,
    title: '5 Format Turnamen Lengkap',
    desc: 'Mendukung Single Elimination, Double Elimination, Liga Round Robin (1 & 2-Leg), Hybrid UCL, dan Battle Royale.',
  },
  {
    icon: Users,
    title: 'Distribusi & Seeding Otomatis',
    desc: 'Algoritma seeding deterministik pangkat-dua, auto-advance BYE, dan pembagian grup UCL yang teratur.',
  },
  {
    icon: BarChart3,
    title: 'Klasemen & Leaderboard Sentral',
    desc: 'Perhitungan otomatis poin kemenangan, selisih gol, head-to-head liga, serta kalkulasi WWCD dan kill points BR.',
  },
  {
    icon: ShieldCheck,
    title: 'Keamanan ACID & Otorisasi Ketat',
    desc: 'Seluruh mutasi dijalankan dalam transaksi PostgreSQL 17 dengan verifikasi hak milik turnamen (owner_id).',
  },
  {
    icon: Share2,
    title: 'Halaman Publik Penonton',
    desc: 'Tautan shareable publik untuk penonton dengan bagan visual SVG presisi tinggi yang responsif di segala perangkat.',
  },
];

export default function Features() {
  return (
    <section
      id="features"
      className="py-20 px-6 relative bg-[#121215]/40 border-y border-[#27272A]"
    >
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-14">
          <h2 className="text-2xl sm:text-3xl font-bold text-zinc-100 mb-3 tracking-tight">
            Fitur Operasi Turnamen Kelas Kompetitif
          </h2>
          <p className="text-sm text-zinc-400 max-w-2xl mx-auto leading-relaxed">
            Dirancang khusus untuk mempermudah operasional komunitas gamer, esports arena, dan liga kompetitif.
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
              className="p-6 rounded-xl bg-[#121215] border border-[#27272A] hover:border-zinc-700 transition-colors group"
            >
              <div className="w-10 h-10 rounded-lg bg-[#18181B] border border-[#27272A] flex items-center justify-center mb-4 text-zinc-200 group-hover:text-white transition-colors">
                <feature.icon className="h-5 w-5" />
              </div>
              <h3 className="text-base font-semibold text-zinc-100 mb-2">
                {feature.title}
              </h3>
              <p className="text-xs text-zinc-400 leading-relaxed">
                {feature.desc}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
