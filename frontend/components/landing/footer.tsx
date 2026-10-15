import Link from "next/link";
import { Gamepad2 } from "lucide-react";

export default function Footer() {
  return (
    <footer className="py-12 border-t border-border-subtle bg-surface-base">
      <div className="max-w-7xl mx-auto px-6 flex flex-col md:flex-row justify-between items-center gap-6">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg bg-surface-raised border border-border-default flex items-center justify-center text-accent">
            <Gamepad2 size={18} />
          </div>
          <span className="text-lg font-bold tracking-tight text-text-primary">Gridify</span>
        </div>

        <div className="text-text-muted text-xs font-mono">
          © {new Date().getFullYear()} Gridify. Modern Esports Tournament Engine.
        </div>

        <div className="flex gap-6 text-sm text-text-muted">
          <Link href="/login" className="hover:text-text-primary transition-colors">
            Masuk
          </Link>
          <Link href="/register" className="hover:text-text-primary transition-colors">
            Daftar
          </Link>
          <a
            href="https://github.com"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-text-primary transition-colors"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  );
}
