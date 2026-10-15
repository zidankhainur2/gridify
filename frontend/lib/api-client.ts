import {
  ApiErrorResponse,
  User,
  Game,
  Tournament,
  TournamentFormat,
  Participant,
  Match,
  StageViewData,
  DashboardSummary,
} from '@/types/api';

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export class ApiClientError extends Error {
  status: number;
  code: string;
  details?: any;
  requestId?: string;

  constructor(status: number, code: string, message: string, details?: any, requestId?: string) {
    super(message);
    this.name = 'ApiClientError';
    this.status = status;
    this.code = code;
    this.details = details;
    this.requestId = requestId;
  }
}

interface FetchOptions extends RequestInit {
  params?: Record<string, string | number | boolean | undefined>;
}

async function request<T>(endpoint: string, options: FetchOptions = {}): Promise<T> {
  const { params, headers = {}, ...rest } = options;

  let url = `${API_BASE_URL}${endpoint}`;
  if (params) {
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) {
        searchParams.append(key, String(value));
      }
    });
    const queryString = searchParams.toString();
    if (queryString) {
      url += `?${queryString}`;
    }
  }

  const reqHeaders: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(headers as Record<string, string>),
  };

  // If executing in SSR (Node.js runtime), forward incoming cookies from next/headers
  if (typeof window === 'undefined') {
    try {
      const { cookies } = await import('next/headers');
      const cookieStore = await cookies();
      const cookieHeader = cookieStore.toString();
      if (cookieHeader) {
        reqHeaders['Cookie'] = cookieHeader;
      }
    } catch {
      // In non-request context (e.g. build time), cookies() might throw, ignore safely
    }
  }

  const res = await fetch(url, {
    ...rest,
    headers: reqHeaders,
    credentials: 'include', // for client-side cookies
  });

  const isJson = res.headers.get('content-type')?.includes('application/json');
  const json = isJson ? await res.json() : null;

  if (!res.ok) {
    const errorResponse = json as ApiErrorResponse | null;
    const code = errorResponse?.error?.code || `HTTP_${res.status}`;
    const message = errorResponse?.error?.message || res.statusText || 'Terjadi kesalahan sistem';
    throw new ApiClientError(
      res.status,
      code,
      message,
      errorResponse?.error?.details,
      errorResponse?.error?.request_id
    );
  }

  // Unwrap backend Envelope: { success: true, data: T } -> T
  if (json && typeof json === 'object' && 'data' in json && (json as any).success === true) {
    return (json as any).data as T;
  }

  return json as T;
}

export const apiClient = {
  auth: {
    async register(payload: {
      email: string;
      username: string;
      password: string;
      full_name?: string;
      first_name?: string;
      last_name?: string;
      phone?: string;
    }): Promise<User> {
      const firstName = payload.first_name || (payload.full_name ? payload.full_name.split(' ')[0] : payload.username);
      const lastName = payload.last_name || (payload.full_name ? payload.full_name.split(' ').slice(1).join(' ') : '');
      const body: Record<string, any> = {
        email: payload.email,
        username: payload.username,
        password: payload.password,
        first_name: firstName || payload.username,
        last_name: lastName,
      };
      if (payload.phone) {
        body.phone = payload.phone;
      }
      return request<User>('/auth/register', {
        method: 'POST',
        body: JSON.stringify(body),
      });
    },

    async login(
      payloadOrIdentifier: { identifier: string; password: string } | string,
      maybePassword?: string
    ): Promise<User> {
      const payload =
        typeof payloadOrIdentifier === 'string'
          ? { identifier: payloadOrIdentifier, password: maybePassword || '' }
          : payloadOrIdentifier;

      return request<User>('/auth/login', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    },

    async logout(): Promise<{ message: string }> {
      return request<{ message: string }>('/auth/logout', {
        method: 'POST',
      });
    },

    async forgotPassword(payloadOrEmail: { email: string } | string): Promise<{ message: string }> {
      const email = typeof payloadOrEmail === 'string' ? payloadOrEmail : payloadOrEmail.email;
      return request<{ message: string }>('/auth/forgot-password', {
        method: 'POST',
        body: JSON.stringify({ email }),
      });
    },

    async resetPassword(
      tokenOrPayload: { token: string; new_password: string } | string,
      maybePassword?: string
    ): Promise<{ message: string }> {
      const payload =
        typeof tokenOrPayload === 'string'
          ? { token: tokenOrPayload, new_password: maybePassword || '' }
          : tokenOrPayload;

      return request<{ message: string }>('/auth/reset-password', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    },
  },

  users: {
    async getMe(): Promise<User> {
      return request<User>('/users/me');
    },

    async updateMe(payload: {
      username?: string;
      first_name?: string;
      last_name?: string;
      phone?: string;
    }): Promise<User> {
      return request<User>('/users/me', {
        method: 'PATCH',
        body: JSON.stringify(payload),
      });
    },

    async changePassword(payload: {
      current_password: string;
      new_password: string;
    }): Promise<{ message: string }> {
      return request<{ message: string }>('/users/me/password', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    },

    async updatePassword(
      currentPasswordOrPayload: string | { current_password: string; new_password: string },
      maybeNewPassword?: string
    ): Promise<{ message: string }> {
      const payload =
        typeof currentPasswordOrPayload === 'string'
          ? { current_password: currentPasswordOrPayload, new_password: maybeNewPassword || '' }
          : currentPasswordOrPayload;
      return this.changePassword(payload);
    },
  },

  dashboard: {
    async getSummary(): Promise<DashboardSummary> {
      return request<DashboardSummary>('/dashboard/summary');
    },
  },

  games: {
    async list(): Promise<Game[]> {
      return request<Game[]>('/games');
    },
  },

  tournaments: {
    async list(params?: {
      search?: string;
      q?: string;
      status?: string;
      page?: number;
      per_page?: number;
    }): Promise<Tournament[]> {
      const queryParams: Record<string, string | number | boolean | undefined> = { ...params };
      if (params?.search && !queryParams.q) {
        queryParams.q = params.search;
      }
      return request<Tournament[]>('/tournaments', { params: queryParams });
    },

    async create(payload: {
      game_id: string;
      title: string;
      description?: string;
      format_type: TournamentFormat;
      best_of?: number;
      home_away?: boolean;
      points_per_win?: number;
      points_per_draw?: number;
      has_third_place?: boolean;
      br_total_matches?: number;
      br_max_teams_per_match?: number;
      match_rules?: Record<string, any>;
    }): Promise<Tournament> {
      return request<Tournament>('/tournaments', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    },

    async get(id: string): Promise<Tournament> {
      return request<Tournament>(`/tournaments/${id}`);
    },

    async getById(id: string): Promise<Tournament> {
      return this.get(id);
    },

    async update(
      id: string,
      payload: {
        title?: string;
        description?: string;
        status?: 'DRAFT' | 'ONGOING' | 'COMPLETED';
        best_of?: number;
        home_away?: boolean;
        points_per_win?: number;
        points_per_draw?: number;
        has_third_place?: boolean;
      }
    ): Promise<Tournament> {
      return request<Tournament>(`/tournaments/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(payload),
      });
    },

    async delete(id: string): Promise<{ message: string }> {
      return request<{ message: string }>(`/tournaments/${id}`, {
        method: 'DELETE',
      });
    },
  },

  participants: {
    async list(tournamentId: string): Promise<Participant[]> {
      return request<Participant[]>(`/tournaments/${tournamentId}/participants`);
    },

    async add(
      tournamentId: string,
      payload: { name: string; contact_info?: string; seed_number?: number }
    ): Promise<Participant> {
      return request<Participant>(`/tournaments/${tournamentId}/participants`, {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    },

    async create(
      tournamentId: string,
      payload: { name: string; contact_info?: string; seed_number?: number }
    ): Promise<Participant> {
      return this.add(tournamentId, payload);
    },

    async update(
      tournamentId: string,
      participantId: string,
      payload: {
        name?: string;
        contact_info?: string;
        seed_number?: number;
        group_name?: string;
        is_verified?: boolean;
      }
    ): Promise<Participant> {
      return request<Participant>(
        `/tournaments/${tournamentId}/participants/${participantId}`,
        {
          method: 'PATCH',
          body: JSON.stringify(payload),
        }
      );
    },

    async delete(tournamentId: string, participantId: string): Promise<{ message: string }> {
      return request<{ message: string }>(
        `/tournaments/${tournamentId}/participants/${participantId}`,
        {
          method: 'DELETE',
        }
      );
    },
  },

  bracket: {
    async generate(tournamentId: string, stageId?: string): Promise<StageViewData> {
      let targetStageId = stageId;
      if (!targetStageId) {
        const t = await apiClient.tournaments.getById(tournamentId);
        targetStageId = t.stages?.[0]?.id;
        if (!targetStageId) {
          throw new Error('Tahap (stage) turnamen belum siap untuk di-generate');
        }
      }
      return request<StageViewData>(
        `/tournaments/${tournamentId}/stages/${targetStageId}/generate`,
        {
          method: 'POST',
        }
      );
    },

    async advanceKnockout(tournamentId: string, stageId?: string): Promise<StageViewData> {
      let targetStageId = stageId;
      if (!targetStageId) {
        const t = await apiClient.tournaments.getById(tournamentId);
        targetStageId = t.stages?.[0]?.id;
        if (!targetStageId) {
          throw new Error('Babak turnamen tidak ditemukan');
        }
      }
      return request<StageViewData>(
        `/tournaments/${tournamentId}/stages/${targetStageId}/advance`,
        {
          method: 'POST',
        }
      );
    },

    async getView(tournamentId: string, stageId?: string): Promise<StageViewData> {
      let targetStageId = stageId;
      if (!targetStageId) {
        const t = await apiClient.tournaments.getById(tournamentId);
        targetStageId = t.stages?.[0]?.id;
        if (!targetStageId) {
          throw new Error('Babak turnamen tidak ditemukan');
        }
      }
      return request<StageViewData>(
        `/tournaments/${tournamentId}/stages/${targetStageId}/view`
      );
    },

    async getStageView(tournamentId: string, stageId?: string): Promise<StageViewData> {
      return this.getView(tournamentId, stageId);
    },
  },

  matches: {
    async updateScore(
      tournamentId: string,
      matchId: string,
      payload: {
        score_a?: number;
        score_b?: number;
        participant_a_score?: number;
        participant_b_score?: number;
        status?: 'ONGOING' | 'COMPLETED';
      }
    ): Promise<Match> {
      const body: Record<string, any> = {
        score_a: payload.score_a ?? payload.participant_a_score ?? 0,
        score_b: payload.score_b ?? payload.participant_b_score ?? 0,
      };
      if (payload.status) {
        body.status = payload.status;
      }
      return request<Match>(
        `/tournaments/${tournamentId}/matches/${matchId}/score`,
        {
          method: 'PATCH',
          body: JSON.stringify(body),
        }
      );
    },

    async updateBattleRoyaleMatch(
      tournamentId: string,
      matchId: string,
      resultsOrPayload: any
    ): Promise<{ message: string }> {
      const rawList = Array.isArray(resultsOrPayload)
        ? resultsOrPayload
        : resultsOrPayload?.results || [];

      const payload = {
        results: rawList.map((r: any) => {
          const pId = r.participant_id || r.team_id || r.teamId;
          const rank = Number(r.placement_rank || r.rank || 0);
          const kills = Number(r.kills || 0);
          const placementPts = Number(r.placement_points || (rank === 1 ? 12 : rank === 2 ? 9 : rank === 3 ? 7 : 0));
          const killPts = Number(r.kill_points || kills);
          return {
            participant_id: pId,
            placement_rank: rank,
            kills: kills,
            placement_points: placementPts,
            kill_points: killPts,
          };
        }),
      };

      return request<{ message: string }>(
        `/tournaments/${tournamentId}/matches/${matchId}/br-results`,
        {
          method: 'PUT',
          body: JSON.stringify(payload),
        }
      );
    },

    async updateBRResults(
      tournamentId: string,
      matchId: string,
      payload: any
    ): Promise<{ message: string }> {
      return this.updateBattleRoyaleMatch(tournamentId, matchId, payload);
    },

    async updateSlot(
      tournamentId: string,
      matchId: string,
      payload: { slot: 'A' | 'B'; participant_id?: string }
    ): Promise<Match> {
      return request<Match>(
        `/tournaments/${tournamentId}/matches/${matchId}/slot`,
        {
          method: 'PATCH',
          body: JSON.stringify(payload),
        }
      );
    },
  },

  public: {
    async getTournament(tournamentId: string, stageId?: string): Promise<StageViewData> {
      return request<StageViewData>(`/public/tournaments/${tournamentId}`, {
        params: { stage_id: stageId },
      });
    },

    async getStageView(tournamentId: string, stageId?: string): Promise<StageViewData> {
      return this.getTournament(tournamentId, stageId);
    },

    getStreamUrl(tournamentId: string): string {
      return `${API_BASE_URL}/public/tournaments/${tournamentId}/stream`;
    },
  },
};
