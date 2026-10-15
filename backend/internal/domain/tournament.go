package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Game struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Genre     string    `json:"genre"`
	Platform  string    `json:"platform"`
	CreatedAt time.Time `json:"created_at"`
}

type Tournament struct {
	ID                  uuid.UUID       `json:"id"`
	OwnerID             uuid.UUID       `json:"owner_id"`
	GameID              uuid.UUID       `json:"game_id"`
	Game                *Game           `json:"game,omitempty"`
	Title               string          `json:"title"`
	Description         string          `json:"description"`
	FormatType          string          `json:"format_type"`
	Status              string          `json:"status"`
	BestOf              int             `json:"best_of"`
	HomeAway            bool            `json:"home_away"`
	PointsPerWin        int             `json:"points_per_win"`
	PointsPerDraw       int             `json:"points_per_draw"`
	HasThirdPlace       bool            `json:"has_third_place"`
	BRTotalMatches      int             `json:"br_total_matches"`
	BRMaxTeamsPerMatch  int             `json:"br_max_teams_per_match"`
	BRPointSystem       json.RawMessage `json:"br_point_system"`
	Stages              []Stage         `json:"stages,omitempty"`
	ParticipantsCount   int             `json:"participants_count"`
	UpcomingMatchesCount int            `json:"upcoming_matches_count,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type Stage struct {
	ID            uuid.UUID `json:"id"`
	TournamentID  uuid.UUID `json:"tournament_id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	SequenceOrder int       `json:"sequence_order"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type Participant struct {
	ID           uuid.UUID `json:"id"`
	TournamentID uuid.UUID `json:"tournament_id"`
	Name         string    `json:"name"`
	ContactInfo  *string   `json:"contact_info,omitempty"`
	SeedNumber   int       `json:"seed_number"`
	GroupName    *string   `json:"group_name,omitempty"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Match struct {
	ID                uuid.UUID    `json:"id"`
	TournamentID      uuid.UUID    `json:"tournament_id"`
	StageID           uuid.UUID    `json:"stage_id"`
	BracketType       string       `json:"bracket_type"`
	RoundNumber       int          `json:"round_number"`
	MatchNumber       int          `json:"match_number"`
	GroupLabel        *string      `json:"group_label,omitempty"`
	ParticipantAID    *uuid.UUID   `json:"participant_a_id,omitempty"`
	ParticipantA      *Participant `json:"participant_a,omitempty"`
	ParticipantBID    *uuid.UUID   `json:"participant_b_id,omitempty"`
	ParticipantB      *Participant `json:"participant_b,omitempty"`
	ScoreA            int          `json:"score_a"`
	ScoreB            int          `json:"score_b"`
	WinnerID          *uuid.UUID   `json:"winner_id,omitempty"`
	IsBye             bool         `json:"is_bye"`
	Status            string       `json:"status"`
	NextMatchID       *uuid.UUID   `json:"next_match_id,omitempty"`
	NextMatchSlot     *string      `json:"next_match_slot,omitempty"`
	LoserNextMatchID  *uuid.UUID   `json:"loser_next_match_id,omitempty"`
	LoserNextMatchSlot *string     `json:"loser_next_match_slot,omitempty"`
	PlayingGroups     []string     `json:"playing_groups"`
	MapName           *string      `json:"map_name,omitempty"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}

type BattleRoyaleMatchResult struct {
	ID              uuid.UUID    `json:"id"`
	MatchID         uuid.UUID    `json:"match_id"`
	ParticipantID   uuid.UUID    `json:"participant_id"`
	Participant     *Participant `json:"participant,omitempty"`
	PlacementRank   int          `json:"placement_rank"`
	Kills           int          `json:"kills"`
	PlacementPoints int          `json:"placement_points"`
	KillPoints      int          `json:"kill_points"`
	TotalPoints     int          `json:"total_points"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type DashboardSummary struct {
	TotalTournaments     int          `json:"total_tournaments"`
	ActiveTournaments    int          `json:"active_tournaments"`
	UpcomingMatchesCount int          `json:"upcoming_matches_count"`
	RecentTournaments    []Tournament `json:"recent_tournaments"`
}
