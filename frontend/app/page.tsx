import Navbar from "@/components/landing/navbar";
import Hero from "@/components/landing/hero";
import Features from "@/components/landing/features";
import HowItWorks from "@/components/landing/how-it-works";
import Footer from "@/components/landing/footer";
import Link from "next/link";
import { ArrowRight, Trophy } from "lucide-react";
import { apiClient } from "@/lib/api-client";
import { User } from "@/types/api";

export default async function Home() {
  let user: User | null = null;
  try {
    user = await apiClient.users.getMe();
  } catch {
    user = null;
  }

  return (
    <div className="min-h-screen bg-surface-base text-text-primary selection:bg-accent/20 font-sans overflow-x-hidden">
      <Navbar user={user} />

      <main>
        <Hero />
        <Features />
        <HowItWorks />

        {/* Final CTA Section */}
        <section className="py-24 px-6 border-t border-border-subtle bg-surface-base relative">
          <div className="max-w-4xl mx-auto text-center relative z-10">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-surface-raised border border-border-default text-xs font-mono text-accent mb-6">
              <Trophy size={14} />
              <span>SIAP MEMULAI KOMPETISI</span>
            </div>
            
            <h2 className="text-3xl md:text-5xl font-black text-text-primary tracking-tight mb-4">
              Mulai Gelar Turnamen Esport Kelas Dunia
            </h2>
            <p className="text-base md:text-lg text-text-muted mb-8 max-w-2xl mx-auto">
              Sistem bracket otomatis berkecepatan tinggi, integrasi scoring real-time, dan spectator view instan tanpa ribet.
            </p>

            <div className="flex flex-col sm:flex-row justify-center gap-4">
              {user ? (
                <Link
                  href="/dashboard/create"
                  className="px-8 py-3.5 rounded-lg bg-accent text-white font-medium hover:bg-accent-hover transition-colors shadow-sm flex items-center justify-center gap-2"
                >
                  Buat Turnamen Sekarang <ArrowRight size={18} />
                </Link>
              ) : (
                <>
                  <Link
                    href="/register"
                    className="px-8 py-3.5 rounded-lg bg-accent text-white font-medium hover:bg-accent-hover transition-colors shadow-sm"
                  >
                    Daftar Gratis
                  </Link>
                  <Link
                    href="/login"
                    className="px-8 py-3.5 rounded-lg bg-surface-raised text-text-primary font-medium hover:bg-surface-elevated transition-colors border border-border-default flex items-center justify-center gap-2"
                  >
                    Masuk ke Dashboard
                    <ArrowRight size={16} />
                  </Link>
                </>
              )}
            </div>
          </div>
        </section>
      </main>

      <Footer />
    </div>
  );
}
