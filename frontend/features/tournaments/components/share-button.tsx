'use client';

import { useState } from 'react';
import { Share2, Check, Copy, Globe, QrCode, X } from 'lucide-react';
import Link from 'next/link';
import QRCode from 'react-qr-code';

export default function ShareButton({ tournamentId }: { tournamentId: string }) {
  const [copied, setCopied] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [showQR, setShowQR] = useState(false);

  const origin = typeof window !== 'undefined' ? window.location.origin : '';
  const publicPath = `/t/${tournamentId}`;
  const fullUrl = `${origin}${publicPath}`;

  const handleCopy = () => {
    navigator.clipboard.writeText(fullUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="inline-flex items-center gap-2 px-3.5 py-2 rounded-lg bg-surface-elevated hover:bg-surface-base border border-border-default text-text-primary text-xs font-medium transition-colors shadow-xs"
        aria-label="Bagikan Turnamen"
      >
        <Share2 size={15} />
        <span>Bagikan</span>
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40 bg-black/40"
            onClick={() => setIsOpen(false)}
          />

          <div className="absolute top-full right-0 mt-2 w-80 p-5 rounded-xl bg-surface-raised border border-border-default shadow-xl z-50 animate-in fade-in zoom-in-95 duration-150">
            <div className="flex justify-between items-start mb-4">
              <div>
                <h4 className="text-sm font-bold text-text-primary flex items-center gap-2">
                  <Globe size={15} className="text-accent" /> Bagikan Turnamen
                </h4>
                <p className="text-xs text-text-muted mt-1">
                  Peserta & penonton dapat melihat bagan realtime via link ini.
                </p>
              </div>
              <button
                onClick={() => setIsOpen(false)}
                className="text-text-muted hover:text-text-primary transition-colors"
                aria-label="Tutup"
              >
                <X size={16} />
              </button>
            </div>

            <div className="flex bg-surface-base p-1 rounded-lg border border-border-subtle mb-4">
              <button
                onClick={() => setShowQR(false)}
                className={`flex-1 py-1.5 text-xs font-medium rounded-md transition-colors ${
                  !showQR
                    ? 'bg-surface-elevated text-text-primary shadow-xs'
                    : 'text-text-muted hover:text-text-primary'
                }`}
              >
                Copy Link
              </button>
              <button
                onClick={() => setShowQR(true)}
                className={`flex-1 py-1.5 text-xs font-medium rounded-md transition-colors ${
                  showQR
                    ? 'bg-surface-elevated text-text-primary shadow-xs'
                    : 'text-text-muted hover:text-text-primary'
                }`}
              >
                QR Code
              </button>
            </div>

            {!showQR ? (
              <div className="space-y-3">
                <div className="flex items-center gap-2 bg-surface-base p-2 rounded-lg border border-border-subtle">
                  <code className="text-xs text-text-secondary truncate flex-1 font-mono select-all">
                    {fullUrl}
                  </code>
                  <button
                    onClick={handleCopy}
                    className="p-1.5 rounded-md bg-surface-elevated hover:bg-surface-raised border border-border-default text-text-primary transition-colors shrink-0"
                    title="Salin Link"
                  >
                    {copied ? (
                      <Check size={14} className="text-emerald-400" />
                    ) : (
                      <Copy size={14} />
                    )}
                  </button>
                </div>
                <Link
                  href={publicPath}
                  target="_blank"
                  className="block w-full py-2 text-center rounded-lg bg-accent hover:bg-accent-hover text-white text-xs font-medium transition-colors shadow-xs"
                >
                  Buka Halaman Publik ↗
                </Link>
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center p-4 bg-white rounded-lg">
                <QRCode
                  size={140}
                  style={{ height: 'auto', maxWidth: '100%', width: '100%' }}
                  value={fullUrl}
                  viewBox="0 0 256 256"
                />
                <p className="text-[11px] text-zinc-900 font-medium mt-3">
                  Scan QR untuk membuka di smartphone
                </p>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}