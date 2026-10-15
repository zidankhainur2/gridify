'use client';

import { motion } from 'framer-motion';

const steps = [
  {
    num: '01',
    title: 'Buat Turnamen',
    desc: 'Pilih cabang game, atur format (Single/Double Elimination, Liga, UCL, atau Battle Royale), dan tentukan aturan skor.',
  },
  {
    num: '02',
    title: 'Daftarkan Peserta',
    desc: 'Masukkan daftar tim atau pemain. Sistem secara otomatis menyusun seeding deterministik atau pembagian grup yang adil.',
  },
  {
    num: '03',
    title: 'Jalankan & Siarkan',
    desc: 'Input skor pertandingan secara berkala. Bagan, klasemen, dan podium juara terupdate seketika di layar publik penonton.',
  },
];

export default function HowItWorks() {
  return (
    <section id="how-it-works" className="py-20 px-6">
      <div className="max-w-5xl mx-auto">
        <div className="text-center mb-14">
          <h2 className="text-2xl sm:text-3xl font-bold text-zinc-100 mb-3 tracking-tight">
            Tiga Langkah Menuju Turnamen Siap Gelar
          </h2>
          <p className="text-sm text-zinc-400">
            Alur kerja yang ringkas dan terstruktur untuk setiap penyelenggara acara.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 relative">
          {steps.map((step, idx) => (
            <motion.div
              key={idx}
              initial={{ opacity: 0, y: 15 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, delay: idx * 0.1 }}
              viewport={{ once: true }}
              className="p-6 rounded-xl bg-[#121215] border border-[#27272A] relative flex flex-col items-center text-center shadow-xs"
            >
              <div className="w-12 h-12 rounded-xl bg-[#18181B] border border-[#27272A] flex items-center justify-center text-base font-bold text-zinc-200 mb-4 font-mono">
                {step.num}
              </div>
              <h3 className="text-base font-semibold text-zinc-100 mb-2">
                {step.title}
              </h3>
              <p className="text-xs text-zinc-400 leading-relaxed">{step.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
