package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gridify/backend/internal/config"
	"github.com/gridify/backend/internal/realtime"
	"github.com/gridify/backend/internal/repository"
	"github.com/gridify/backend/internal/service"
	"github.com/gridify/backend/internal/service/bracket"
	"github.com/gridify/backend/internal/transport/http/handler"
	"github.com/gridify/backend/internal/transport/http/httputil"
	"github.com/gridify/backend/internal/transport/http/middleware"
)

type RouterDeps struct {
	Config          *config.Config
	Pool            *pgxpool.Pool
	Logger          *slog.Logger
	AuthSvc         *service.AuthService
	TournamentSvc   *service.TournamentService
	ParticipantSvc  *service.ParticipantService
	StandingsSvc    *service.StandingsService
	Progression     *bracket.ProgressionEngine
	Hub             *realtime.Hub
	UserRepo        *repository.UserRepository
	GameRepo        *repository.GameRepository
	TournamentRepo  *repository.TournamentRepository
	StageRepo       *repository.StageRepository
	ParticipantRepo *repository.ParticipantRepository
	MatchRepo       *repository.MatchRepository
	BRRepo          *repository.BattleRoyaleRepository
}

func NewRouter(cfg *config.Config, pool *pgxpool.Pool, logger *slog.Logger) *chi.Mux {
	userRepo := repository.NewUserRepository(pool)
	gameRepo := repository.NewGameRepository(pool)
	tournamentRepo := repository.NewTournamentRepository(pool)
	stageRepo := repository.NewStageRepository(pool)
	participantRepo := repository.NewParticipantRepository(pool)
	matchRepo := repository.NewMatchRepository(pool)
	brRepo := repository.NewBattleRoyaleRepository(pool)

	hub := realtime.NewHub()
	authSvc := service.NewAuthService(userRepo)
	tournamentSvc := service.NewTournamentService(pool, tournamentRepo, stageRepo, gameRepo)
	participantSvc := service.NewParticipantService(participantRepo, tournamentRepo)
	standingsSvc := service.NewStandingsService(matchRepo, brRepo, participantRepo, stageRepo, tournamentRepo)
	progression := bracket.NewProgressionEngine(matchRepo, stageRepo, tournamentRepo, participantRepo)

	return NewRouterWithDeps(RouterDeps{
		Config:          cfg,
		Pool:            pool,
		Logger:          logger,
		AuthSvc:         authSvc,
		TournamentSvc:   tournamentSvc,
		ParticipantSvc:  participantSvc,
		StandingsSvc:    standingsSvc,
		Progression:     progression,
		Hub:             hub,
		UserRepo:        userRepo,
		GameRepo:        gameRepo,
		TournamentRepo:  tournamentRepo,
		StageRepo:       stageRepo,
		ParticipantRepo: participantRepo,
		MatchRepo:       matchRepo,
		BRRepo:          brRepo,
	})
}

func NewRouterWithDeps(deps RouterDeps) *chi.Mux {
	if deps.Hub == nil {
		deps.Hub = realtime.NewHub()
	}
	r := chi.NewRouter()

	// Global middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(deps.Logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS(deps.Config.CORSAllowedOrigins))

	// Health check endpoints
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httputil.JSON(w, http.StatusOK, map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := deps.Pool.Ping(ctx); err != nil {
			deps.Logger.Error("Readiness probe database ping failed", slog.String("error", err.Error()))
			httputil.RespondError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database connection is not ready", nil, middleware.GetRequestID(r.Context()))
			return
		}

		httputil.JSON(w, http.StatusOK, map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"database":  "connected",
		})
	})

	// Handlers
	authHandler := handler.NewAuthHandler(deps.AuthSvc, deps.Config)
	userHandler := handler.NewUserHandler(deps.AuthSvc)
	gameHandler := handler.NewGameHandler(deps.GameRepo)
	dashboardHandler := handler.NewDashboardHandler(deps.TournamentRepo)
	tournamentHandler := handler.NewTournamentHandler(deps.TournamentSvc, deps.Hub)
	participantHandler := handler.NewParticipantHandler(deps.ParticipantSvc, deps.Hub)
	bracketHandler := handler.NewBracketHandler(
		deps.Pool,
		deps.TournamentRepo,
		deps.StageRepo,
		deps.ParticipantRepo,
		deps.MatchRepo,
		deps.StandingsSvc,
		deps.Progression,
		deps.Hub,
	)
	matchHandler := handler.NewMatchHandler(
		deps.Pool,
		deps.TournamentRepo,
		deps.MatchRepo,
		deps.BRRepo,
		deps.Progression,
		deps.Hub,
	)
	publicHandler := handler.NewPublicHandler(
		deps.TournamentRepo,
		deps.StageRepo,
		deps.StandingsSvc,
		deps.Hub,
		deps.Logger,
	)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
		})

		// Public Games route
		r.Get("/games", gameHandler.ListGames)

		// Protected User routes
		r.Route("/users", func(r chi.Router) {
			r.Use(middleware.RequireAuth(deps.AuthSvc))
			r.Get("/me", userHandler.GetMe)
			r.Patch("/me", userHandler.UpdateMe)
			r.Post("/me/password", userHandler.ChangePassword)
		})

		// Protected Dashboard routes
		r.Route("/dashboard", func(r chi.Router) {
			r.Use(middleware.RequireAuth(deps.AuthSvc))
			r.Get("/summary", dashboardHandler.GetSummary)
		})

		// Tournament routes
		r.Route("/tournaments", func(r chi.Router) {
			// Require auth for organizer mutations
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth(deps.AuthSvc))
				r.Get("/", tournamentHandler.List)
				r.Post("/", tournamentHandler.Create)
				r.Get("/{id}", tournamentHandler.Get)
				r.Patch("/{id}", tournamentHandler.Update)
				r.Delete("/{id}", tournamentHandler.Delete)

				// Participant mutations (require auth)
				r.Post("/{id}/participants", participantHandler.Add)
				r.Patch("/{id}/participants/{pid}", participantHandler.Update)
				r.Delete("/{id}/participants/{pid}", participantHandler.Delete)

				// Stage bracket mutations (require auth)
				r.Post("/{id}/stages/{stageId}/generate", bracketHandler.Generate)
				r.Post("/{id}/stages/{stageId}/advance", bracketHandler.AdvanceKnockout)

				// Match mutations (require auth)
				r.Patch("/{id}/matches/{matchId}/score", matchHandler.UpdateScore)
				r.Put("/{id}/matches/{matchId}/br-results", matchHandler.UpdateBRResults)
				r.Patch("/{id}/matches/{matchId}/slot", matchHandler.UpdateSlot)
			})

			// Public / Spectator access
			r.Get("/{id}/participants", participantHandler.List)
			r.Get("/{id}/stages/{stageId}/view", bracketHandler.GetView)
		})

		// Public spectator & SSE stream routes
		r.Route("/public", func(r chi.Router) {
			r.Get("/tournaments/{id}", publicHandler.GetTournament)
			r.Get("/tournaments/{id}/stream", publicHandler.StreamTournament)
		})
	})

	return r
}
