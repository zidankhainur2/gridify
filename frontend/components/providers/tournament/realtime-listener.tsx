'use client';

import { useTournamentSSE } from '@/hooks/use-tournament-sse';
import { WifiOff } from 'lucide-react';

export default function RealtimeListener({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const isConnected = useTournamentSSE(tournamentId);

  return (
    <div className="fixed bottom-4 right-4 z-50 pointer-events-none">
      <div
        className={`flex items-center gap-2 px-3 py-1.5 rounded-full text-[11px] font-mono font-medium backdrop-blur-md border shadow-lg transition-all duration-300 ${
          isConnected
            ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400'
            : 'bg-zinc-800/80 border-border-default text-text-muted'
        }`}
      >
        {isConnected ? (
          <>
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
            </span>
            <span>SSE REALTIME</span>
          </>
        ) : (
          <>
            <WifiOff size={12} />
            <span>CONNECTING SSE...</span>
          </>
        )}
      </div>
    </div>
  );
}
