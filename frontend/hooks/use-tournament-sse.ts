'use client';

import { useEffect, useState, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';

export function useTournamentSSE(
  tournamentId: string,
  onUpdate?: (event: { type: string; data?: any }) => void
) {
  const router = useRouter();
  const [isConnected, setIsConnected] = useState(false);
  const onUpdateRef = useRef(onUpdate);
  onUpdateRef.current = onUpdate;

  useEffect(() => {
    if (!tournamentId) return;

    const streamUrl = apiClient.public.getStreamUrl(tournamentId);
    let eventSource: EventSource | null = null;
    let retryTimeout: NodeJS.Timeout;

    const connect = () => {
      try {
        eventSource = new EventSource(streamUrl);

        eventSource.onopen = () => {
          setIsConnected(true);
        };

        eventSource.onerror = () => {
          setIsConnected(false);
          if (eventSource) {
            eventSource.close();
          }
          // Coba reconnect setelah 5 detik
          retryTimeout = setTimeout(connect, 5000);
        };

        // Standard message handler
        eventSource.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === 'ping') return;

            onUpdateRef.current?.({ type: 'message', data });
            router.refresh();
          } catch {
            // Ignore parse errors on raw messages
          }
        };

        // Specific SSE event listeners
        eventSource.addEventListener('match_updated', (event: MessageEvent) => {
          try {
            const data = JSON.parse(event.data);
            onUpdateRef.current?.({ type: 'match_updated', data });
          } catch {}
          router.refresh();
        });

        eventSource.addEventListener('tournament_updated', (event: MessageEvent) => {
          try {
            const data = JSON.parse(event.data);
            onUpdateRef.current?.({ type: 'tournament_updated', data });
          } catch {}
          router.refresh();
        });

        eventSource.addEventListener('participant_updated', (event: MessageEvent) => {
          try {
            const data = JSON.parse(event.data);
            onUpdateRef.current?.({ type: 'participant_updated', data });
          } catch {}
          router.refresh();
        });
      } catch (err) {
        console.error('Failed to init SSE EventSource:', err);
        setIsConnected(false);
      }
    };

    connect();

    return () => {
      clearTimeout(retryTimeout);
      if (eventSource) {
        eventSource.close();
      }
    };
  }, [tournamentId, router]);

  return isConnected;
}
