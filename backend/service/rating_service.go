package service

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"

	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/repository"
)

type RatingService struct {
	ratingRepo repository.RatingRepository
}

func NewRatingService(ratingRepo repository.RatingRepository) *RatingService {
	return &RatingService{
		ratingRepo: ratingRepo,
	}
}

// CalculateNewRatings applies the ELO formula. Result: "1-0" white won, "0-1" black won, "0.5-0.5" draw.
func (s *RatingService) CalculateNewRatings(
	whiteRating, blackRating int,
	result string,
	whiteGameCount, blackGameCount int,
) (int, int, error) {
	var whiteScore, blackScore float64

	switch result {
	case "1-0":
		whiteScore = 1.0
		blackScore = 0.0
	case "0-1":
		whiteScore = 0.0
		blackScore = 1.0
	case "0.5-0.5":
		whiteScore = 0.5
		blackScore = 0.5
	default:
		return 0, 0, fmt.Errorf("rating_service.CalculateNewRatings: invalid result: %s", result)
	}

	whiteKFactor := s.getKFactor(whiteGameCount)
	blackKFactor := s.getKFactor(blackGameCount)

	whiteExpected := s.calculateExpected(float64(whiteRating), float64(blackRating))
	blackExpected := s.calculateExpected(float64(blackRating), float64(whiteRating))

	whiteNewRating := int(float64(whiteRating) + float64(whiteKFactor)*(whiteScore-whiteExpected))
	blackNewRating := int(float64(blackRating) + float64(blackKFactor)*(blackScore-blackExpected))

	if whiteNewRating < 100 {
		whiteNewRating = 100
	}
	if blackNewRating < 100 {
		blackNewRating = 100
	}

	return whiteNewRating, blackNewRating, nil
}

// ApplyRatingChange updates both players' ratings in the database.
func (s *RatingService) ApplyRatingChange(
	ctx context.Context,
	whiteID, blackID uuid.UUID,
	newWhiteRating, newBlackRating int,
) (int, int, error) {
	whiteOld, err := s.ratingRepo.GetUserRating(ctx, whiteID)
	if err != nil {
		return 0, 0, fmt.Errorf("rating_service.ApplyRatingChange: get white rating: %w", err)
	}

	blackOld, err := s.ratingRepo.GetUserRating(ctx, blackID)
	if err != nil {
		return 0, 0, fmt.Errorf("rating_service.ApplyRatingChange: get black rating: %w", err)
	}

	whiteDelta := newWhiteRating - whiteOld
	blackDelta := newBlackRating - blackOld

	if err := s.ratingRepo.ApplyChange(ctx, &models.RatingChange{
		UserID:    whiteID,
		OldRating: whiteOld,
		NewRating: newWhiteRating,
		Delta:     whiteDelta,
	}); err != nil {
		return 0, 0, fmt.Errorf("rating_service.ApplyRatingChange: apply white change: %w", err)
	}

	if err := s.ratingRepo.ApplyChange(ctx, &models.RatingChange{
		UserID:    blackID,
		OldRating: blackOld,
		NewRating: newBlackRating,
		Delta:     blackDelta,
	}); err != nil {
		return 0, 0, fmt.Errorf("rating_service.ApplyRatingChange: apply black change: %w", err)
	}

	return whiteDelta, blackDelta, nil
}

func (s *RatingService) getKFactor(gameCount int) int {
	if gameCount < 30 {
		return 32
	}
	return 16
}

func (s *RatingService) calculateExpected(playerRating, opponentRating float64) float64 {
	return 1.0 / (1.0 + math.Pow(10.0, (opponentRating-playerRating)/400.0))
}
