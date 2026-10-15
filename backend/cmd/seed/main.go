package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gridify/backend/internal/config"
	"github.com/gridify/backend/internal/database"
	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/repository"
	"github.com/gridify/backend/internal/service"
	"github.com/gridify/backend/internal/service/bracket"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Gagal memuat konfigurasi: %v", err)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	defer pool.Close()

	fmt.Println("🚀 Memulai Gridify Database Seeder...")

	userRepo := repository.NewUserRepository(pool)
	gameRepo := repository.NewGameRepository(pool)
	tournamentRepo := repository.NewTournamentRepository(pool)
	stageRepo := repository.NewStageRepository(pool)
	participantRepo := repository.NewParticipantRepository(pool)
	matchRepo := repository.NewMatchRepository(pool)
	brRepo := repository.NewBattleRoyaleRepository(pool)
	progression := bracket.NewProgressionEngine(matchRepo, stageRepo, tournamentRepo, participantRepo)

	// 1. Seed or get Default Organizer Account
	organizer, err := seedOrganizer(ctx, userRepo)
	if err != nil {
		log.Fatalf("Gagal menyiapkan akun organizer: %v", err)
	}
	fmt.Printf("✅ Organizer Account Ready: %s %s (%s) [ID: %s]\n", organizer.FirstName, organizer.LastName, organizer.Email, organizer.ID)

	// 2. Fetch Master Games
	games, err := gameRepo.GetAllGames(ctx)
	if err != nil || len(games) == 0 {
		log.Fatalf("Gagal memuat master games: %v", err)
	}
	gameMap := make(map[string]domain.Game)
	for _, g := range games {
		gameMap[g.Slug] = g
	}

	// 3. Clean up previous seed tournaments
	cleanPreviousSeeds(ctx, pool, organizer.ID)

	// 4. Seed Tournament 1: Single Elimination (Valorant)
	seedSingleElimination(ctx, pool, tournamentRepo, stageRepo, participantRepo, matchRepo, progression, organizer.ID, gameMap["valorant"])

	// 5. Seed Tournament 2: Double Elimination (Dota 2)
	seedDoubleElimination(ctx, pool, tournamentRepo, stageRepo, participantRepo, matchRepo, progression, organizer.ID, gameMap["dota-2"])

	// 6. Seed Tournament 3: Round Robin (EA FC 24)
	seedRoundRobin(ctx, pool, tournamentRepo, stageRepo, participantRepo, matchRepo, progression, organizer.ID, gameMap["ea-sports-fc"])

	// 7. Seed Tournament 4: Hybrid UCL (Mobile Legends)
	seedHybridUCL(ctx, pool, tournamentRepo, stageRepo, participantRepo, matchRepo, progression, organizer.ID, gameMap["mobile-legends"])

	// 8. Seed Tournament 5: Battle Royale (PUBG Mobile)
	seedBattleRoyale(ctx, pool, tournamentRepo, stageRepo, participantRepo, matchRepo, brRepo, organizer.ID, gameMap["pubg-mobile"])

	fmt.Println("\n🎉 Database Seeding Berhasil 100%! Seluruh 5 format turnamen siap digunakan.")
	fmt.Println("👉 Kredensial Login Organizer:")
	fmt.Println("   Email   : organizer@gridify.test (atau username: organizer)")
	fmt.Println("   Password: Password123!")
}

func seedOrganizer(ctx context.Context, userRepo *repository.UserRepository) (*domain.User, error) {
	email := "organizer@gridify.test"
	username := "organizer"

	user, err := userRepo.GetUserByEmailOrUsername(ctx, email)
	if err == nil && user != nil {
		return user, nil
	}

	hash, err := service.HashPassword("Password123!")
	if err != nil {
		return nil, err
	}

	newUser := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		Username:     username,
		FirstName:    "Gridify",
		LastName:     "Organizer",
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := userRepo.CreateUser(ctx, newUser); err != nil {
		return nil, err
	}
	return newUser, nil
}

func cleanPreviousSeeds(ctx context.Context, pool *pgxpool.Pool, ownerID uuid.UUID) {
	query := `DELETE FROM tournaments WHERE owner_id = $1 AND title LIKE '%Demo%'`
	res, err := pool.Exec(ctx, query, ownerID)
	if err == nil {
		fmt.Printf("🧹 Membersihkan %d turnamen demo lama...\n", res.RowsAffected())
	} else {
		fmt.Printf("⚠️ Gagal membersihkan turnamen demo: %v\n", err)
	}
}

func strPtr(s string) *string {
	return &s
}

// ----------------------------------------------------------------------------
// 1. Single Elimination (Valorant)
// ----------------------------------------------------------------------------
func seedSingleElimination(
	ctx context.Context,
	pool *pgxpool.Pool,
	tourRepo *repository.TournamentRepository,
	stageRepo *repository.StageRepository,
	partRepo *repository.ParticipantRepository,
	matchRepo *repository.MatchRepository,
	progression *bracket.ProgressionEngine,
	ownerID uuid.UUID,
	game domain.Game,
) {
	fmt.Println("\n📦 Seeding 1/5: Single Elimination (Valorant)...")

	tourID := uuid.New()
	stageID := uuid.New()

	tour := &domain.Tournament{
		ID:            tourID,
		OwnerID:       ownerID,
		GameID:        game.ID,
		Title:         "[Demo] Valorant Champions Invitational",
		Description:   "Turnamen eliminasi tunggal 8 tim tier-1 dengan perebutan Juara 3 (Bronze Match).",
		FormatType:    "SINGLE_ELIMINATION",
		Status:        "ONGOING",
		BestOf:        3,
		HasThirdPlace: true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	stage := &domain.Stage{
		ID:            stageID,
		TournamentID:  tourID,
		Name:          "Main Bracket",
		Type:          "SINGLE_ELIMINATION",
		SequenceOrder: 1,
		Status:        "ONGOING",
		CreatedAt:     time.Now().UTC(),
	}

	teams := []string{
		"Fnatic", "Sentinels", "Paper Rex", "DRX",
		"Team Liquid", "LOUD", "EDward Gaming", "Evil Geniuses",
	}

	var participants []domain.Participant
	err := database.WithTx(ctx, pool, func(tx pgx.Tx) error {
		if err := tourRepo.CreateTournamentWithTx(ctx, tx, tour); err != nil {
			return err
		}
		if err := stageRepo.CreateStageWithTx(ctx, tx, stage); err != nil {
			return err
		}
		for i, name := range teams {
			p := domain.Participant{
				ID:           uuid.New(),
				TournamentID: tourID,
				Name:         name,
				SeedNumber:   i + 1,
				IsVerified:   true,
				CreatedAt:    time.Now().UTC(),
			}
			if err := partRepo.CreateParticipantWithTx(ctx, tx, &p); err != nil {
				return err
			}
			participants = append(participants, p)
		}

		_, err := bracket.GenerateSingleElimination(ctx, tx, matchRepo, tour, stage, participants)
		return err
	})
	if err != nil {
		log.Fatalf("Gagal membuat Single Elimination: %v", err)
	}

	// Advance Quarterfinal matches
	matches, _ := matchRepo.GetMatchesByStageID(ctx, stageID)
	round1Matches := make([]*domain.Match, 0)
	for i := range matches {
		if matches[i].RoundNumber == 1 && matches[i].BracketType == "UPPER" {
			round1Matches = append(round1Matches, &matches[i])
		}
	}

	// Update QF results
	scores := [][2]int{{2, 0}, {1, 2}, {2, 1}, {2, 0}}
	for i, m := range round1Matches {
		if i < len(scores) {
			_ = database.WithTx(ctx, pool, func(tx pgx.Tx) error {
				_, err := progression.UpdateMatchScore(ctx, tx, tourID, m.ID, scores[i][0], scores[i][1], "COMPLETED")
				return err
			})
		}
	}
	fmt.Printf("   ✓ Single Elimination siap (8 Tim, QF selesai, Semifinals live)\n")
}

// ----------------------------------------------------------------------------
// 2. Double Elimination (Dota 2)
// ----------------------------------------------------------------------------
func seedDoubleElimination(
	ctx context.Context,
	pool *pgxpool.Pool,
	tourRepo *repository.TournamentRepository,
	stageRepo *repository.StageRepository,
	partRepo *repository.ParticipantRepository,
	matchRepo *repository.MatchRepository,
	progression *bracket.ProgressionEngine,
	ownerID uuid.UUID,
	game domain.Game,
) {
	fmt.Println("\n📦 Seeding 2/5: Double Elimination (Dota 2)...")

	tourID := uuid.New()
	stageID := uuid.New()

	tour := &domain.Tournament{
		ID:          tourID,
		OwnerID:     ownerID,
		GameID:      game.ID,
		Title:       "[Demo] Dota 2 Aegis Masters",
		Description: "Turnamen eliminasi ganda 8 tim dengan Upper Bracket, Lower Bracket, dan Grand Final.",
		FormatType:  "DOUBLE_ELIMINATION",
		Status:      "ONGOING",
		BestOf:      3,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	stage := &domain.Stage{
		ID:            stageID,
		TournamentID:  tourID,
		Name:          "Double Elimination Bracket",
		Type:          "DOUBLE_ELIMINATION",
		SequenceOrder: 1,
		Status:        "ONGOING",
		CreatedAt:     time.Now().UTC(),
	}

	teams := []string{
		"Team Spirit", "Gaimin Gladiators", "Team Liquid", "BetBoom Team",
		"Azure Ray", "LGD Gaming", "Tundra Esports", "OG",
	}

	var participants []domain.Participant
	err := database.WithTx(ctx, pool, func(tx pgx.Tx) error {
		if err := tourRepo.CreateTournamentWithTx(ctx, tx, tour); err != nil {
			return err
		}
		if err := stageRepo.CreateStageWithTx(ctx, tx, stage); err != nil {
			return err
		}
		for i, name := range teams {
			p := domain.Participant{
				ID:           uuid.New(),
				TournamentID: tourID,
				Name:         name,
				SeedNumber:   i + 1,
				IsVerified:   true,
				CreatedAt:    time.Now().UTC(),
			}
			if err := partRepo.CreateParticipantWithTx(ctx, tx, &p); err != nil {
				return err
			}
			participants = append(participants, p)
		}

		_, err := bracket.GenerateDoubleElimination(ctx, tx, matchRepo, tour, stage, participants)
		return err
	})
	if err != nil {
		log.Fatalf("Gagal membuat Double Elimination: %v", err)
	}

	// Advance UB Round 1 matches
	matches, _ := matchRepo.GetMatchesByStageID(ctx, stageID)
	ubR1 := make([]*domain.Match, 0)
	for i := range matches {
		if matches[i].RoundNumber == 1 && matches[i].BracketType == "UPPER" {
			ubR1 = append(ubR1, &matches[i])
		}
	}

	scores := [][2]int{{2, 0}, {2, 1}, {2, 0}, {2, 1}}
	for i, m := range ubR1 {
		if i < len(scores) {
			_ = database.WithTx(ctx, pool, func(tx pgx.Tx) error {
				_, err := progression.UpdateMatchScore(ctx, tx, tourID, m.ID, scores[i][0], scores[i][1], "COMPLETED")
				return err
			})
		}
	}
	fmt.Printf("   ✓ Double Elimination siap (8 Tim, UB R1 selesai, drop-down LB aktif)\n")
}

// ----------------------------------------------------------------------------
// 3. Round Robin (EA FC 24)
// ----------------------------------------------------------------------------
func seedRoundRobin(
	ctx context.Context,
	pool *pgxpool.Pool,
	tourRepo *repository.TournamentRepository,
	stageRepo *repository.StageRepository,
	partRepo *repository.ParticipantRepository,
	matchRepo *repository.MatchRepository,
	progression *bracket.ProgressionEngine,
	ownerID uuid.UUID,
	game domain.Game,
) {
	fmt.Println("\n📦 Seeding 3/5: Round Robin (EA Sports FC 24)...")

	tourID := uuid.New()
	stageID := uuid.New()

	tour := &domain.Tournament{
		ID:            tourID,
		OwnerID:       ownerID,
		GameID:        game.ID,
		Title:         "[Demo] EA FC European Super League",
		Description:   "Format liga round robin penuh sistem Home & Away (2-Legs) dengan klasemen otomatis.",
		FormatType:    "ROUND_ROBIN",
		Status:        "ONGOING",
		HomeAway:      true,
		PointsPerWin:  3,
		PointsPerDraw: 1,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	stage := &domain.Stage{
		ID:            stageID,
		TournamentID:  tourID,
		Name:          "Klasemen Liga",
		Type:          "ROUND_ROBIN",
		SequenceOrder: 1,
		Status:        "ONGOING",
		CreatedAt:     time.Now().UTC(),
	}

	teams := []string{
		"Manchester City", "Real Madrid", "Bayern Munich", "Paris Saint-Germain",
	}

	var participants []domain.Participant
	err := database.WithTx(ctx, pool, func(tx pgx.Tx) error {
		if err := tourRepo.CreateTournamentWithTx(ctx, tx, tour); err != nil {
			return err
		}
		if err := stageRepo.CreateStageWithTx(ctx, tx, stage); err != nil {
			return err
		}
		for i, name := range teams {
			p := domain.Participant{
				ID:           uuid.New(),
				TournamentID: tourID,
				Name:         name,
				SeedNumber:   i + 1,
				IsVerified:   true,
				CreatedAt:    time.Now().UTC(),
			}
			if err := partRepo.CreateParticipantWithTx(ctx, tx, &p); err != nil {
				return err
			}
			participants = append(participants, p)
		}

		_, err := bracket.GenerateRoundRobin(ctx, tx, matchRepo, tour, stage, participants)
		return err
	})
	if err != nil {
		log.Fatalf("Gagal membuat Round Robin: %v", err)
	}

	// Populate match scores for Round 1 & Round 2
	matches, _ := matchRepo.GetMatchesByStageID(ctx, stageID)
	simScores := [][2]int{{3, 1}, {2, 2}, {2, 1}, {1, 0}}
	for i := 0; i < len(simScores) && i < len(matches); i++ {
		_ = database.WithTx(ctx, pool, func(tx pgx.Tx) error {
			_, err := progression.UpdateMatchScore(ctx, tx, tourID, matches[i].ID, simScores[i][0], simScores[i][1], "COMPLETED")
			return err
		})
	}
	fmt.Printf("   ✓ Round Robin siap (4 Tim, 12 Match Home/Away, Klasemen aktif)\n")
}

// ----------------------------------------------------------------------------
// 4. Hybrid UCL (Mobile Legends)
// ----------------------------------------------------------------------------
func seedHybridUCL(
	ctx context.Context,
	pool *pgxpool.Pool,
	tourRepo *repository.TournamentRepository,
	stageRepo *repository.StageRepository,
	partRepo *repository.ParticipantRepository,
	matchRepo *repository.MatchRepository,
	progression *bracket.ProgressionEngine,
	ownerID uuid.UUID,
	game domain.Game,
) {
	fmt.Println("\n📦 Seeding 4/5: Hybrid UCL (Mobile Legends)...")

	tourID := uuid.New()
	groupStageID := uuid.New()

	tour := &domain.Tournament{
		ID:            tourID,
		OwnerID:       ownerID,
		GameID:        game.ID,
		Title:         "[Demo] MLBB World Championship (Hybrid UCL)",
		Description:   "Format hybrid Champions League: Fase Grup (Grup A & B) dilanjutkan ke Babak Knockout Single Elimination.",
		FormatType:    "HYBRID_UCL",
		Status:        "ONGOING",
		BestOf:        3,
		PointsPerWin:  3,
		PointsPerDraw: 1,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	groupStage := &domain.Stage{
		ID:            groupStageID,
		TournamentID:  tourID,
		Name:          "Fase Grup",
		Type:          "ROUND_ROBIN",
		SequenceOrder: 1,
		Status:        "ONGOING",
		CreatedAt:     time.Now().UTC(),
	}

	teamsA := []string{"ONIC Esports", "Blacklist International", "Geek Fam ID", "EVOS Glory"}
	teamsB := []string{"RRQ Hoshi", "ECHO Philippines", "Bigetron Alpha", "AP Bren"}

	var participants []domain.Participant
	err := database.WithTx(ctx, pool, func(tx pgx.Tx) error {
		if err := tourRepo.CreateTournamentWithTx(ctx, tx, tour); err != nil {
			return err
		}
		if err := stageRepo.CreateStageWithTx(ctx, tx, groupStage); err != nil {
			return err
		}

		seedNum := 1
		for _, name := range teamsA {
			p := domain.Participant{
				ID:           uuid.New(),
				TournamentID: tourID,
				Name:         name,
				SeedNumber:   seedNum,
				GroupName:    strPtr("A"),
				IsVerified:   true,
				CreatedAt:    time.Now().UTC(),
			}
			seedNum++
			if err := partRepo.CreateParticipantWithTx(ctx, tx, &p); err != nil {
				return err
			}
			participants = append(participants, p)
		}
		for _, name := range teamsB {
			p := domain.Participant{
				ID:           uuid.New(),
				TournamentID: tourID,
				Name:         name,
				SeedNumber:   seedNum,
				GroupName:    strPtr("B"),
				IsVerified:   true,
				CreatedAt:    time.Now().UTC(),
			}
			seedNum++
			if err := partRepo.CreateParticipantWithTx(ctx, tx, &p); err != nil {
				return err
			}
			participants = append(participants, p)
		}

		_, err := bracket.GenerateRoundRobin(ctx, tx, matchRepo, tour, groupStage, participants)
		return err
	})
	if err != nil {
		log.Fatalf("Gagal membuat Hybrid UCL: %v", err)
	}

	// Simulate completing group matches so top 2 from each group are clear
	groupMatches, _ := matchRepo.GetMatchesByStageID(ctx, groupStageID)
	for i := range groupMatches {
		scoreA := 2
		scoreB := 0
		if i%2 == 1 {
			scoreA = 2
			scoreB = 1
		}
		_ = database.WithTx(ctx, pool, func(tx pgx.Tx) error {
			_, err := progression.UpdateMatchScore(ctx, tx, tourID, groupMatches[i].ID, scoreA, scoreB, "COMPLETED")
			return err
		})
	}

	// Advance qualified teams to Knockout Stage!
	_ = database.WithTx(ctx, pool, func(tx pgx.Tx) error {
		_, _, err := progression.AdvanceGroupToKnockout(ctx, tx, tourID, groupStageID)
		return err
	})

	fmt.Printf("   ✓ Hybrid UCL siap (Fase Grup A & B selesai, Knockout Bracket terbuat)\n")
}

// ----------------------------------------------------------------------------
// 5. Battle Royale (PUBG Mobile)
// ----------------------------------------------------------------------------
func seedBattleRoyale(
	ctx context.Context,
	pool *pgxpool.Pool,
	tourRepo *repository.TournamentRepository,
	stageRepo *repository.StageRepository,
	partRepo *repository.ParticipantRepository,
	matchRepo *repository.MatchRepository,
	brRepo *repository.BattleRoyaleRepository,
	ownerID uuid.UUID,
	game domain.Game,
) {
	fmt.Println("\n📦 Seeding 5/5: Battle Royale (PUBG Mobile)...")

	tourID := uuid.New()
	stageID := uuid.New()

	tour := &domain.Tournament{
		ID:                 tourID,
		OwnerID:            ownerID,
		GameID:             game.ID,
		Title:              "[Demo] PUBG Mobile Pro League Finals",
		Description:        "Turnamen Battle Royale 16 tim 4 ronde pertandingan dengan sistem poin resmi (10-6-5-4-3-2-1).",
		FormatType:         "BATTLE_ROYALE",
		Status:             "ONGOING",
		BRTotalMatches:     4,
		BRMaxTeamsPerMatch: 16,
		BRPointSystem:      json.RawMessage(`{"1": 10, "2": 6, "3": 5, "4": 4, "5": 3, "6": 2, "7": 1, "8": 1}`),
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	stage := &domain.Stage{
		ID:            stageID,
		TournamentID:  tourID,
		Name:          "Grand Finals Lobby",
		Type:          "LEADERBOARD",
		SequenceOrder: 1,
		Status:        "ONGOING",
		CreatedAt:     time.Now().UTC(),
	}

	teams := []string{
		"Bigetron RA", "Alter Ego LIMAX", "Morph GPX", "BOOM Esports",
		"Vampire Esports", "FaZe Clan", "Nova Esports", "Stalwart Esports",
		"D'Xavier", "IHC Esports", "S2G Esports", "Four Angry Men",
		"Reject", "Alpha7 Esports", "Tianba", "Loops Esports",
	}

	var participants []domain.Participant
	err := database.WithTx(ctx, pool, func(tx pgx.Tx) error {
		if err := tourRepo.CreateTournamentWithTx(ctx, tx, tour); err != nil {
			return err
		}
		if err := stageRepo.CreateStageWithTx(ctx, tx, stage); err != nil {
			return err
		}
		for i, name := range teams {
			p := domain.Participant{
				ID:           uuid.New(),
				TournamentID: tourID,
				Name:         name,
				SeedNumber:   i + 1,
				IsVerified:   true,
				CreatedAt:    time.Now().UTC(),
			}
			if err := partRepo.CreateParticipantWithTx(ctx, tx, &p); err != nil {
				return err
			}
			participants = append(participants, p)
		}

		_, err := bracket.GenerateBattleRoyale(ctx, tx, matchRepo, tour, stage, participants)
		return err
	})
	if err != nil {
		log.Fatalf("Gagal membuat Battle Royale: %v", err)
	}

	// Populate match results for Match 1 & Match 2
	matches, _ := matchRepo.GetMatchesByStageID(ctx, stageID)
	placementPts := []int{10, 6, 5, 4, 3, 2, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0}
	killsDistribution := []int{12, 8, 7, 5, 4, 3, 2, 4, 1, 2, 0, 1, 0, 1, 0, 0}

	if len(matches) >= 2 {
		// Match 1
		var resultsM1 []domain.BattleRoyaleMatchResult
		for i, p := range participants {
			place := i + 1
			pts := placementPts[i]
			kills := killsDistribution[i]
			resultsM1 = append(resultsM1, domain.BattleRoyaleMatchResult{
				ID:              uuid.New(),
				MatchID:         matches[0].ID,
				ParticipantID:   p.ID,
				PlacementRank:   place,
				Kills:           kills,
				PlacementPoints: pts,
				KillPoints:      kills,
				TotalPoints:     pts + kills,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			})
		}
		_ = database.WithTx(ctx, pool, func(tx pgx.Tx) error {
			if err := brRepo.SaveMatchResults(ctx, tx, matches[0].ID, resultsM1); err != nil {
				return err
			}
			matches[0].Status = "COMPLETED"
			return matchRepo.UpdateMatch(ctx, tx, &matches[0])
		})

		// Match 2 (Rotated winners)
		var resultsM2 []domain.BattleRoyaleMatchResult
		for i := range participants {
			idx := (i + 4) % len(participants)
			p := participants[idx]
			place := i + 1
			pts := placementPts[i]
			kills := killsDistribution[(i+2)%len(killsDistribution)]
			resultsM2 = append(resultsM2, domain.BattleRoyaleMatchResult{
				ID:              uuid.New(),
				MatchID:         matches[1].ID,
				ParticipantID:   p.ID,
				PlacementRank:   place,
				Kills:           kills,
				PlacementPoints: pts,
				KillPoints:      kills,
				TotalPoints:     pts + kills,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			})
		}
		_ = database.WithTx(ctx, pool, func(tx pgx.Tx) error {
			if err := brRepo.SaveMatchResults(ctx, tx, matches[1].ID, resultsM2); err != nil {
				return err
			}
			matches[1].Status = "COMPLETED"
			return matchRepo.UpdateMatch(ctx, tx, &matches[1])
		})
	}

	fmt.Printf("   ✓ Battle Royale siap (16 Tim, 4 Matches, Match 1 & 2 selesai, Leaderboard aktif)\n")
}
