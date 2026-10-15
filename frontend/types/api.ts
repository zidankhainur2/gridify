export interface User {
  id: string;
  email: string;
  username: string;
  first_name: string;
  last_name: string;
  phone?: string;
  created_at: string;
  updated_at: string;
}

export interface Game {
  id: string;
  slug: string;
  name: string;
  genre: 'MOBA' | 'FPS' | 'SPORTS' | 'BATTLE_ROYALE';
  platform: string;
  created_at: string;
}

export type TournamentFormat =
  | 'SINGLE_ELIMINATION'
  | 'DOUBLE_ELIMINATION'
  | 'ROUND_ROBIN'
  | 'HYBRID_UCL'
  | 'BATTLE_ROYALE';

export type TournamentStatus = 'DRAFT' | 'ONGOING' | 'COMPLETED';

export interface Tournament {
  id: string;
  owner_id: string;
  game_id: string;
  game?: Game;
  games?: Game; // alias for backward-compatibility
  title: string;
  description: string;
  format_type: TournamentFormat;
  status: TournamentStatus;
  best_of: number;
  home_away: boolean;
  points_per_win: number;
  points_per_draw: number;
  has_third_place: boolean;
  br_total_matches: number;
  br_max_teams_per_match: number;
  br_point_system?: Record<string, any>;
  match_rules?: Record<string, any>;
  stages?: Stage[];
  participants_count: number;
  upcoming_matches_count?: number;
  created_at: string;
  updated_at: string;
}

export type StageType =
  | 'SINGLE_ELIMINATION'
  | 'DOUBLE_ELIMINATION'
  | 'ROUND_ROBIN'
  | 'LEADERBOARD'
  | 'KNOCKOUT';

export type StageStatus = 'PENDING' | 'ONGOING' | 'COMPLETED';

export interface Stage {
  id: string;
  tournament_id: string;
  name: string;
  type: StageType;
  sequence_order: number;
  status: StageStatus;
  created_at: string;
}

export interface Participant {
  id: string;
  tournament_id: string;
  name: string;
  contact_info?: string;
  seed_number: number;
  group_name?: string;
  is_verified: boolean;
  created_at?: string;
  updated_at?: string;
}

export type MatchBracketType =
  | 'UPPER'
  | 'LOWER'
  | 'GRAND_FINAL'
  | 'THIRD_PLACE'
  | 'ROUND_ROBIN'
  | 'BATTLE_ROYALE';

export type MatchStatus = 'SCHEDULED' | 'ONGOING' | 'COMPLETED';

export interface Match {
  id: string;
  tournament_id: string;
  stage_id: string;
  bracket_type: MatchBracketType;
  round_number: number;
  match_number: number;
  group_label?: string;
  participant_a_id?: string;
  participant_a?: Participant;
  participant_b_id?: string;
  participant_b?: Participant;
  score_a?: number;
  score_b?: number;
  participant_a_score?: number;
  participant_b_score?: number;
  scores?: {
    a?: number;
    b?: number;
    groups?: string[];
    results?: any[];
    [key: string]: any;
  };
  winner_id?: string;
  is_bye: boolean;
  status: MatchStatus;
  next_match_id?: string;
  next_match_slot?: 'A' | 'B';
  loser_next_match_id?: string;
  loser_next_match_slot?: 'A' | 'B';
  playing_groups?: string[];
  map_name?: string;
  br_results?: any[];
}

export interface StandingRow {
  participant?: Participant;
  participant_id?: string;
  participant_name?: string;
  group_name?: string;
  played: number;
  won: number;
  drawn: number;
  lost: number;
  goals_for: number;
  goals_against: number;
  goal_difference: number;
  points: number;
  rank: number;
  total_kills?: number;
  placement_points?: number;
  kill_points?: number;
  total_points?: number;
}

export interface GroupStanding {
  group_label: string;
  rows: StandingRow[];
}

export interface BRLeaderboardRow {
  participant?: Participant;
  participant_id?: string;
  participant_name?: string;
  matches_played: number;
  wins: number;
  total_kills: number;
  placement_points: number;
  kill_points: number;
  total_points: number;
  rank: number;
}

export interface ChampionResult {
  rank: number;
  participant?: Participant;
  title?: string;
}

export interface Champions {
  first?: Participant;
  second?: Participant;
  third?: Participant;
  results?: ChampionResult[];
}

export interface StageViewData {
  stage: Stage;
  tournament?: Tournament;
  stages?: Stage[];
  matches: Match[];
  standings?: GroupStanding[];
  br_leaderboard?: BRLeaderboardRow[];
  champions?: Champions | ChampionResult[];
}

export interface DashboardSummary {
  total_tournaments: number;
  active_tournaments: number;
  upcoming_matches_count: number;
  recent_tournaments: Tournament[];
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  meta?: {
    page?: number;
    per_page?: number;
    total?: number;
  };
}

export interface ApiError {
  code: string;
  message: string;
  details?: any;
  request_id?: string;
}

export interface ApiErrorResponse {
  success: false;
  error: ApiError;
}
