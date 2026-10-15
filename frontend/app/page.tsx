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

        {/* Final CTA Section (AI-slop pill removed) */}
        <section className="py-24 px-4 md:px-8 border-t border-border-default bg-surface-raised/30 relative">
          <div className="max-w-4xl mx-auto text-center relative z-10">
            <h2 className="text-3xl md:text-5xl font-display font-black text-text-primary tracking-tight mb-4">
              Mulai Gelar Turnamen Esports Anda
            </h2>
            <p className="text-sm md:text-base text-text-secondary mb-8 max-w-2xl mx-auto leading-relaxed">
              Sistem bracket otomatis berkecepatan tinggi, integrasi scoring real-time, dan spectator view instan tanpa konfigurasi rumit.
            </p>

            <div className="flex flex-col sm:flex-row justify-center gap-3.5">
              {user ? (
                <Link
                  href="/dashboard/create"
                  className="px-8 py-3.5 rounded-xl bg-accent text-accent-contrast font-bold hover:bg-accent-hover transition-all shadow-md shadow-accent/20 flex items-center justify-center gap-2 hover:scale-[1.02] active:scale-[0.98]"
                >
                  <Trophy size={18} />
                  <span>Buat Turnamen Sekarang</span>
                  <ArrowRight size={18} />
                </Link>
              ) : (
                <>
                  <Link
                    href="/register"
                    className="px-8 py-3.5 rounded-xl bg-accent text-accent-contrast font-bold hover:bg-accent-hover transition-all shadow-md shadow-accent/20 flex items-center justify-center hover:scale-[1.02] active:scale-[0.98]"
                  >
                    Daftar Gratis
                  </Link>
                  <Link
                    href="/login"
                    className="px-8 py-3.5 rounded-xl bg-surface-elevated border border-border-default text-text-primary font-semibold hover:bg-surface-overlay hover:border-border-strong transition-all flex items-center justify-center"
                  >
                    Masuk Organizer
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
