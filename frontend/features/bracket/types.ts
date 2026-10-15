import { Match, Participant } from '@/types/api';

export type MatchWithParticipants = Match;

export type BattleRoyaleResult = {
  teamId: string;
  rank: number | '';
  kills: number | '';
  total: number;
};

export type GroupData = {
  name: string;
  teams: { id: string; name: string }[];
};
