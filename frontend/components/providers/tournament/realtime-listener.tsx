'use client';

import { useTournamentSSE } from '@/hooks/use-tournament-sse';
import { WifiOff, Radio } from 'lucide-react';

export default function RealtimeListener({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const isConnected = useTournamentSSE(tournamentId);

  return (
    <div className="fixed bottom-3 right-3 z-30 pointer-events-none select-none">
      <div
        className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-mono font-medium backdrop-blur-md border shadow-sm transition-all duration-300 opacity-70 hover:opacity-100 ${
          isConnected
            ? 'bg-surface-elevated/90 border-emerald-500/30 text-emerald-400'
            : 'bg-surface-elevated/90 border-amber-500/30 text-amber-400'
        }`}
      >
        {isConnected ? (
          <>
            <span className="relative flex h-1.5 w-1.5">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500" />
            </span>
            <span className="tracking-wide">LIVE SSE</span>
          </>
        ) : (
          <>
            <WifiOff size={11} className="text-amber-400" />
            <span className="tracking-wide">CONNECTING...</span>
          </>
        )}
      </div>
    </div>
  );
}
