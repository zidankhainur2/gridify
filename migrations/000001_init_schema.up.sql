CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Fungsi trigger universal untuk memperbarui kolom updated_at secara otomatis
CREATE OR REPLACE FUNCTION set_updated_at_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 1. USERS & AUTHENTICATION
-- ============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash TEXT NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL DEFAULT '',
    phone VARCHAR(30),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_users_email_nonempty CHECK (LENGTH(TRIM(email)) >= 5),
    CONSTRAINT chk_users_username_format CHECK (username ~ '^[a-zA-Z0-9_.-]{3,50}$')
);

-- Case-insensitive unique indexes untuk email dan username
CREATE UNIQUE INDEX uq_users_email_lower ON users (LOWER(email));
CREATE UNIQUE INDEX uq_users_username_lower ON users (LOWER(username));

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address VARCHAR(45),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);

CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_reset_user_id ON password_reset_tokens(user_id);

-- ============================================================================
-- 2. MASTER GAMES
-- ============================================================================
CREATE TABLE games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(120) NOT NULL,
    genre VARCHAR(32) NOT NULL CHECK (
        genre IN ('MOBA', 'FPS', 'SPORTS', 'BATTLE_ROYALE')
    ),
    platform VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 3. TOURNAMENTS
-- ============================================================================
CREATE TABLE tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE RESTRICT,
    title VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    format_type VARCHAR(32) NOT NULL CHECK (
        format_type IN (
            'SINGLE_ELIMINATION',
            'DOUBLE_ELIMINATION',
            'ROUND_ROBIN',
            'HYBRID_UCL',
            'BATTLE_ROYALE'
        )
    ),
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (
        status IN ('DRAFT', 'ONGOING', 'COMPLETED')
    ),
    best_of SMALLINT NOT NULL DEFAULT 1 CHECK (best_of IN (1, 3, 5, 7)),
    home_away BOOLEAN NOT NULL DEFAULT FALSE,
    points_per_win SMALLINT NOT NULL DEFAULT 3 CHECK (points_per_win >= 1),
    points_per_draw SMALLINT NOT NULL DEFAULT 1 CHECK (points_per_draw >= 0),
    has_third_place BOOLEAN NOT NULL DEFAULT FALSE,
    br_total_matches SMALLINT NOT NULL DEFAULT 6 CHECK (br_total_matches BETWEEN 1 AND 24),
    br_max_teams_per_match SMALLINT NOT NULL DEFAULT 16 CHECK (br_max_teams_per_match BETWEEN 4 AND 64),
    br_point_system JSONB NOT NULL DEFAULT '{"killPoint": 1, "placementPoints": {"1": 10, "2": 6, "3": 5, "4": 4, "5": 3, "6": 2, "7": 1, "8": 1}}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_tournaments_title_len CHECK (LENGTH(TRIM(title)) >= 3)
);

CREATE INDEX idx_tournaments_owner_status ON tournaments(owner_id, status, created_at DESC);
CREATE INDEX idx_tournaments_game_id ON tournaments(game_id);

CREATE TRIGGER trg_tournaments_updated_at
BEFORE UPDATE ON tournaments
FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();

-- ============================================================================
-- 4. STAGES
-- ============================================================================
CREATE TABLE stages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(32) NOT NULL CHECK (
        type IN (
            'SINGLE_ELIMINATION',
            'DOUBLE_ELIMINATION',
            'ROUND_ROBIN',
            'LEADERBOARD'
        )
    ),
    sequence_order SMALLINT NOT NULL CHECK (sequence_order >= 1),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (
        status IN ('PENDING', 'ONGOING', 'COMPLETED')
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tournament_id, sequence_order)
);

CREATE INDEX idx_stages_tournament_order ON stages(tournament_id, sequence_order);

-- ============================================================================
-- 5. PARTICIPANTS
-- ============================================================================
CREATE TABLE participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    contact_info VARCHAR(160),
    seed_number INT NOT NULL DEFAULT 1 CHECK (seed_number >= 1),
    group_name VARCHAR(32),
    is_verified BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_participants_name_nonempty CHECK (LENGTH(TRIM(name)) >= 1)
);

CREATE INDEX idx_participants_tournament_seed ON participants(tournament_id, seed_number, created_at);
CREATE INDEX idx_participants_tournament_group ON participants(tournament_id, group_name);

CREATE TRIGGER trg_participants_updated_at
BEFORE UPDATE ON participants
FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();

-- ============================================================================
-- 6. MATCHES
-- ============================================================================
CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    stage_id UUID NOT NULL REFERENCES stages(id) ON DELETE CASCADE,
    bracket_type VARCHAR(20) NOT NULL DEFAULT 'UPPER' CHECK (
        bracket_type IN (
            'UPPER',
            'LOWER',
            'GRAND_FINAL',
            'THIRD_PLACE',
            'ROUND_ROBIN',
            'BATTLE_ROYALE'
        )
    ),
    round_number INT NOT NULL CHECK (round_number >= 1),
    match_number INT NOT NULL CHECK (match_number >= 1),
    group_label VARCHAR(32),
    participant_a_id UUID REFERENCES participants(id) ON DELETE SET NULL,
    participant_b_id UUID REFERENCES participants(id) ON DELETE SET NULL,
    score_a INT NOT NULL DEFAULT 0 CHECK (score_a >= 0),
    score_b INT NOT NULL DEFAULT 0 CHECK (score_b >= 0),
    winner_id UUID REFERENCES participants(id) ON DELETE SET NULL,
    is_bye BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED' CHECK (
        status IN ('SCHEDULED', 'ONGOING', 'COMPLETED')
    ),
    next_match_id UUID REFERENCES matches(id) ON DELETE SET NULL,
    next_match_slot CHAR(1) CHECK (next_match_slot IN ('A', 'B')),
    loser_next_match_id UUID REFERENCES matches(id) ON DELETE SET NULL,
    loser_next_match_slot CHAR(1) CHECK (loser_next_match_slot IN ('A', 'B')),
    playing_groups TEXT[] NOT NULL DEFAULT '{}',
    map_name VARCHAR(80),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_matches_stage_bracket_round_num UNIQUE (stage_id, bracket_type, round_number, match_number),
    CONSTRAINT chk_matches_distinct_participants CHECK (
        participant_a_id IS NULL OR participant_b_id IS NULL OR participant_a_id <> participant_b_id
    ),
    CONSTRAINT chk_matches_valid_winner CHECK (
        winner_id IS NULL OR winner_id = participant_a_id OR winner_id = participant_b_id
    )
);

CREATE INDEX idx_matches_stage_order ON matches(stage_id, bracket_type, round_number, match_number);
CREATE INDEX idx_matches_tournament_upcoming ON matches(tournament_id, status, round_number, match_number)
    WHERE status <> 'COMPLETED';
CREATE INDEX idx_matches_next_match_id ON matches(next_match_id) WHERE next_match_id IS NOT NULL;
CREATE INDEX idx_matches_loser_next_match_id ON matches(loser_next_match_id) WHERE loser_next_match_id IS NOT NULL;

CREATE TRIGGER trg_matches_updated_at
BEFORE UPDATE ON matches
FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();

-- ============================================================================
-- 7. BATTLE ROYALE MATCH RESULTS
-- ============================================================================
CREATE TABLE battle_royale_match_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    placement_rank SMALLINT NOT NULL CHECK (placement_rank >= 1),
    kills SMALLINT NOT NULL DEFAULT 0 CHECK (kills >= 0),
    placement_points INT NOT NULL DEFAULT 0,
    kill_points INT NOT NULL DEFAULT 0,
    total_points INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (match_id, participant_id)
);

CREATE INDEX idx_br_results_match_rank ON battle_royale_match_results(match_id, placement_rank);
CREATE INDEX idx_br_results_participant_id ON battle_royale_match_results(participant_id);

CREATE TRIGGER trg_br_results_updated_at
BEFORE UPDATE ON battle_royale_match_results
FOR EACH ROW EXECUTE FUNCTION set_updated_at_timestamp();
