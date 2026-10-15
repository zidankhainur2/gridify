'use client';

import * as React from 'react';
import Link from 'next/link';
import { Gamepad2, LayoutDashboard, Menu, X } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { User } from '@/types/api';

export default function Navbar({ user }: { user: User | null }) {
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);

  const navLinks = [
    { href: '#games', label: 'Cabang Game' },
    { href: '#features', label: 'Fitur' },
    { href: '#how-it-works', label: 'Cara Kerja' },
  ];

  return (
    <header className="fixed top-0 w-full z-50 border-b border-border-default bg-surface-raised/90 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 md:px-8 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2.5 group">
          <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent group-hover:border-accent/40 transition-colors shadow-xs">
            <Gamepad2 className="h-4 w-4" />
          </div>
          <span className="text-lg font-display font-bold tracking-tight text-text-primary">
            Gridify
          </span>
        </Link>

        {/* Desktop Menu */}
        <nav className="hidden md:flex items-center gap-8 text-sm font-medium text-text-muted">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="hover:text-text-primary transition-colors py-1"
            >
              {link.label}
            </Link>
          ))}
        </nav>

        {/* Auth Buttons & Mobile Toggle */}
        <div className="flex items-center gap-3">
          {user ? (
            <div className="flex items-center gap-4">
              <div className="hidden md:flex flex-col items-end text-right">
                <span className="text-[11px] text-text-muted font-medium">Masuk sebagai</span>
                <span className="text-xs font-semibold text-text-primary max-w-[140px] truncate">
                  {user.username || user.email}
                </span>
              </div>

              <Link
                href="/dashboard"
                className="flex items-center gap-2 px-3.5 py-1.5 rounded-lg bg-accent text-accent-contrast text-xs font-bold hover:bg-accent-hover transition-colors shadow-xs"
              >
                <LayoutDashboard className="h-3.5 w-3.5" />
                <span>Dashboard</span>
              </Link>
            </div>
          ) : (
            <div className="hidden sm:flex items-center gap-2.5">
              <Link
                href="/login"
                className="text-xs font-medium text-text-secondary hover:text-text-primary px-3 py-1.5 transition-colors"
              >
                Masuk Organizer
              </Link>
              <Link
                href="/register"
                className="px-3.5 py-1.5 rounded-lg bg-accent text-accent-contrast text-xs font-bold hover:bg-accent-hover transition-colors shadow-xs"
              >
                Daftar Gratis
              </Link>
            </div>
          )}

          {/* Mobile Menu Hamburger Button */}
          <button
            type="button"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden p-2 rounded-lg text-text-muted hover:text-text-primary hover:bg-surface-elevated transition-colors border border-border-subtle"
            aria-label="Toggle navigation menu"
          >
            {mobileMenuOpen ? <X size={18} /> : <Menu size={18} />}
          </button>
        </div>
      </div>

      {/* Mobile Drawer */}
      <AnimatePresence>
        {mobileMenuOpen && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.2 }}
            className="md:hidden border-b border-border-default bg-surface-raised px-4 py-4 space-y-3"
          >
            <div className="flex flex-col space-y-2">
              {navLinks.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  onClick={() => setMobileMenuOpen(false)}
                  className="px-3 py-2 rounded-lg text-sm font-medium text-text-secondary hover:text-text-primary hover:bg-surface-elevated transition-colors"
                >
                  {link.label}
                </Link>
              ))}
            </div>

            {!user && (
              <div className="pt-3 border-t border-border-subtle flex flex-col gap-2">
                <Link
                  href="/login"
                  onClick={() => setMobileMenuOpen(false)}
                  className="w-full text-center py-2 text-xs font-medium text-text-secondary hover:text-text-primary rounded-lg border border-border-default bg-surface-elevated"
                >
                  Masuk Organizer
                </Link>
                <Link
                  href="/register"
                  onClick={() => setMobileMenuOpen(false)}
                  className="w-full text-center py-2 text-xs font-bold text-accent-contrast bg-accent hover:bg-accent-hover rounded-lg shadow-xs"
                >
                  Daftar Gratis
                </Link>
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </header>
  );
}
