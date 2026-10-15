import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import AppProviders from "@/components/providers/app-providers";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: "Gridify - Sistem Manajemen & Pembuat Bagan Turnamen Esports",
  description:
    "Platform operasi turnamen esports profesional: Single Elimination, Double Elimination, Round Robin, Hybrid UCL, dan Battle Royale dengan pembaruan realtime.",
  keywords: [
    "tournament bracket",
    "esport manager",
    "pembuat bagan turnamen",
    "bracket generator",
    "mlbb tournament",
    "pubg leaderboard",
  ],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" className={`${inter.variable} ${jetbrainsMono.variable} dark`}>
      <body className="bg-[#09090B] text-[#F4F4F5] antialiased">
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
