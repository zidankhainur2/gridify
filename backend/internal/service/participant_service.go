package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/gridify/backend/internal/domain"
	"github.com/gridify/backend/internal/repository"
)

type ParticipantService struct {
	participantRepo *repository.ParticipantRepository
	tournamentRepo  *repository.TournamentRepository
}

func NewParticipantService(
	participantRepo *repository.ParticipantRepository,
	tournamentRepo *repository.TournamentRepository,
) *ParticipantService {
	return &ParticipantService{
		participantRepo: participantRepo,
		tournamentRepo:  tournamentRepo,
	}
}

type AddParticipantInput struct {
	TournamentID uuid.UUID
	UserID       uuid.UUID
	Name         string
	ContactInfo  *string
	SeedNumber   int
}

func (s *ParticipantService) AddParticipant(ctx context.Context, in AddParticipantInput) (*domain.Participant, error) {
	// 1. Verify Ownership
	if err := s.tournamentRepo.VerifyTournamentOwnership(ctx, in.TournamentID, in.UserID); err != nil {
		return nil, err
	}

	// 2. Check Tournament Status (Must be DRAFT to add)
	tourney, err := s.tournamentRepo.GetTournamentByID(ctx, in.TournamentID)
	if err != nil {
		return nil, err
	}
	if tourney.Status != "DRAFT" {
		return nil, domain.ErrTournamentLocked
	}

	// 3. Validate Name
	cleanName := strings.TrimSpace(in.Name)
	if len(cleanName) < 1 {
		return nil, fmt.Errorf("%w: nama peserta tidak boleh kosong", domain.ErrBadRequest)
	}

	// 4. Assign Seed
	seed := in.SeedNumber
	if seed <= 0 {
		nextSeed, err := s.participantRepo.GetNextSeedNumber(ctx, in.TournamentID)
		if err != nil {
			return nil, err
		}
		seed = nextSeed
	}

	p := &domain.Participant{
		ID:           uuid.New(),
		TournamentID: in.TournamentID,
		Name:         cleanName,
		ContactInfo:  in.ContactInfo,
		SeedNumber:   seed,
		IsVerified:   true,
	}

	if err := s.participantRepo.CreateParticipant(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

type UpdateParticipantInput struct {
	TournamentID  uuid.UUID
	ParticipantID uuid.UUID
	UserID        uuid.UUID
	Name          *string
	ContactInfo   *string
	SeedNumber    *int
	GroupName     *string
	IsVerified    *bool
}

func (s *ParticipantService) UpdateParticipant(ctx context.Context, in UpdateParticipantInput) (*domain.Participant, error) {
	// 1. Verify Ownership
	if err := s.tournamentRepo.VerifyTournamentOwnership(ctx, in.TournamentID, in.UserID); err != nil {
		return nil, err
	}

	// 2. Fetch existing participant
	p, err := s.participantRepo.GetParticipantByID(ctx, in.ParticipantID)
	if err != nil {
		return nil, err
	}
	if p.TournamentID != in.TournamentID {
		return nil, domain.ErrNotFound
	}

	// 3. Apply updates
	if in.Name != nil {
		clean := strings.TrimSpace(*in.Name)
		if len(clean) < 1 {
			return nil, fmt.Errorf("%w: nama peserta tidak boleh kosong", domain.ErrBadRequest)
		}
		p.Name = clean
	}
	if in.ContactInfo != nil {
		p.ContactInfo = in.ContactInfo
	}
	if in.SeedNumber != nil && *in.SeedNumber >= 1 {
		p.SeedNumber = *in.SeedNumber
	}
	if in.GroupName != nil {
		p.GroupName = in.GroupName
	}
	if in.IsVerified != nil {
		p.IsVerified = *in.IsVerified
	}

	if err := s.participantRepo.UpdateParticipant(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ParticipantService) DeleteParticipant(ctx context.Context, tournamentID, participantID, userID uuid.UUID) error {
	// 1. Verify Ownership
	if err := s.tournamentRepo.VerifyTournamentOwnership(ctx, tournamentID, userID); err != nil {
		return err
	}

	// 2. Check Tournament Status (Must be DRAFT to delete)
	tourney, err := s.tournamentRepo.GetTournamentByID(ctx, tournamentID)
	if err != nil {
		return err
	}
	if tourney.Status != "DRAFT" {
		return domain.ErrTournamentLocked
	}

	return s.participantRepo.DeleteParticipant(ctx, participantID, tournamentID)
}

func (s *ParticipantService) ListParticipants(ctx context.Context, tournamentID uuid.UUID) ([]domain.Participant, error) {
	return s.participantRepo.ListParticipantsByTournamentID(ctx, tournamentID)
}
