package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/chess-nfac/backend/chess"
	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/repository"
	"github.com/chess-nfac/backend/utils"
)

const startingFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

type GameService struct {
	gameRepo      repository.GameRepository
	moveRepo      repository.MoveRepository
	ratingService *RatingService
	engine        *chess.Engine
}

func NewGameService(
	gameRepo repository.GameRepository,
	moveRepo repository.MoveRepository,
	ratingService *RatingService,
	engine *chess.Engine,
) *GameService {
	return &GameService{
		gameRepo:      gameRepo,
		moveRepo:      moveRepo,
		ratingService: ratingService,
		engine:        engine,
	}
}

func (s *GameService) CreateGame(ctx context.Context, whiteID, blackID uuid.UUID, whiteRating, blackRating int) (*models.Game, error) {
	game := &models.Game{
		WhiteID:           whiteID,
		BlackID:           blackID,
		Status:            models.GameStatusActive,
		FEN:               startingFEN,
		WhiteRatingBefore: &whiteRating,
		BlackRatingBefore: &blackRating,
	}

	if err := s.gameRepo.Create(ctx, game); err != nil {
		return nil, fmt.Errorf("game_service.CreateGame: %w", err)
	}

	return game, nil
}

func (s *GameService) GetGame(ctx context.Context, gameID uuid.UUID) (*models.Game, error) {
	game, err := s.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("game_service.GetGame: %w", err)
	}
	if game == nil {
		return nil, utils.NewAppError("game_not_found", "Game not found", 404)
	}
	return game, nil
}

func (s *GameService) GetMoves(ctx context.Context, gameID uuid.UUID) ([]models.Move, error) {
	moves, err := s.moveRepo.FindByGameID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("game_service.GetMoves: %w", err)
	}
	return moves, nil
}

func (s *GameService) GetUserGames(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Game, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	games, total, err := s.gameRepo.FindByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("game_service.GetUserGames: %w", err)
	}
	return games, total, nil
}

func (s *GameService) ProcessMove(ctx context.Context, gameID, playerID uuid.UUID, moveNotation string) (*models.Game, *models.Move, bool, error) {
	game, err := s.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return nil, nil, false, fmt.Errorf("game_service.ProcessMove: %w", err)
	}
	if game == nil {
		return nil, nil, false, utils.NewAppError("game_not_found", "Game not found", 404)
	}
	if game.Status != models.GameStatusActive {
		return nil, nil, false, utils.NewAppError("game_not_active", "Game is not active", 400)
	}

	// Determine whose turn it is from FEN (field[1]: "w" or "b")
	fenParts := strings.Fields(game.FEN)
	if len(fenParts) < 2 {
		return nil, nil, false, fmt.Errorf("game_service.ProcessMove: invalid FEN")
	}
	sideToMove := fenParts[1]

	if sideToMove == "w" && playerID != game.WhiteID {
		return nil, nil, false, utils.NewAppError("not_your_turn", "It is not your turn", 400)
	}
	if sideToMove == "b" && playerID != game.BlackID {
		return nil, nil, false, utils.NewAppError("not_your_turn", "It is not your turn", 400)
	}

	newFEN, err := s.engine.ValidateMove(game.FEN, moveNotation)
	if err != nil {
		return nil, nil, false, utils.NewAppError("invalid_move", "That move is not legal", 400)
	}

	// Count existing moves for move number
	existingMoves, err := s.moveRepo.FindByGameID(ctx, gameID)
	if err != nil {
		return nil, nil, false, fmt.Errorf("game_service.ProcessMove: get moves: %w", err)
	}

	move := &models.Move{
		GameID:     gameID,
		PlayerID:   playerID,
		MoveNumber: len(existingMoves) + 1,
		Notation:   moveNotation,
		FENAfter:   newFEN,
	}

	if err := s.moveRepo.Create(ctx, move); err != nil {
		return nil, nil, false, fmt.Errorf("game_service.ProcessMove: save move: %w", err)
	}

	// Check if game is over
	over, reason, err := s.engine.IsGameOver(newFEN)
	if err != nil {
		return nil, nil, false, fmt.Errorf("game_service.ProcessMove: check game over: %w", err)
	}

	if over {
		result := s.determineResult(sideToMove, reason)
		if err := s.finishGame(ctx, game, newFEN, result); err != nil {
			return nil, nil, false, fmt.Errorf("game_service.ProcessMove: finish game: %w", err)
		}
		game.Status = models.GameStatusCompleted
		game.Result = &result
	} else {
		if err := s.gameRepo.UpdateFEN(ctx, gameID, newFEN); err != nil {
			return nil, nil, false, fmt.Errorf("game_service.ProcessMove: update FEN: %w", err)
		}
		game.FEN = newFEN
	}

	return game, move, over, nil
}

func (s *GameService) ResignGame(ctx context.Context, gameID, playerID uuid.UUID) (*models.Game, error) {
	game, err := s.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("game_service.ResignGame: %w", err)
	}
	if game == nil {
		return nil, utils.NewAppError("game_not_found", "Game not found", 404)
	}
	if game.Status != models.GameStatusActive {
		return nil, utils.NewAppError("game_not_active", "Game is not active", 400)
	}
	if playerID != game.WhiteID && playerID != game.BlackID {
		return nil, utils.NewAppError("forbidden", "You are not a player in this game", 403)
	}

	var result models.GameResult
	if playerID == game.WhiteID {
		result = models.GameResultBlackWins
	} else {
		result = models.GameResultWhiteWins
	}

	if err := s.finishGame(ctx, game, game.FEN, result); err != nil {
		return nil, fmt.Errorf("game_service.ResignGame: %w", err)
	}

	game.Status = models.GameStatusCompleted
	game.Result = &result
	return game, nil
}

func (s *GameService) finishGame(ctx context.Context, game *models.Game, finalFEN string, result models.GameResult) error {
	var eloResult string
	switch result {
	case models.GameResultWhiteWins:
		eloResult = "1-0"
	case models.GameResultBlackWins:
		eloResult = "0-1"
	default:
		eloResult = "0.5-0.5"
	}

	whiteGameCount, err := s.gameRepo.GetUserGameCount(ctx, game.WhiteID)
	if err != nil {
		return fmt.Errorf("game_service.finishGame: white game count: %w", err)
	}
	blackGameCount, err := s.gameRepo.GetUserGameCount(ctx, game.BlackID)
	if err != nil {
		return fmt.Errorf("game_service.finishGame: black game count: %w", err)
	}

	whiteRatingBefore := 1200
	if game.WhiteRatingBefore != nil {
		whiteRatingBefore = *game.WhiteRatingBefore
	}
	blackRatingBefore := 1200
	if game.BlackRatingBefore != nil {
		blackRatingBefore = *game.BlackRatingBefore
	}

	newWhiteRating, newBlackRating, err := s.ratingService.CalculateNewRatings(
		whiteRatingBefore, blackRatingBefore,
		eloResult,
		whiteGameCount, blackGameCount,
	)
	if err != nil {
		return fmt.Errorf("game_service.finishGame: calculate ratings: %w", err)
	}

	if _, _, err := s.ratingService.ApplyRatingChange(ctx, game.WhiteID, game.BlackID, newWhiteRating, newBlackRating); err != nil {
		return fmt.Errorf("game_service.finishGame: apply rating change: %w", err)
	}

	if err := s.gameRepo.UpdateFEN(ctx, game.ID, finalFEN); err != nil {
		return fmt.Errorf("game_service.finishGame: update FEN: %w", err)
	}

	if err := s.gameRepo.Finish(ctx, game.ID, result, newWhiteRating, newBlackRating); err != nil {
		return fmt.Errorf("game_service.finishGame: save game: %w", err)
	}

	return nil
}

// determineResult: sideToMove is who just moved ("w" or "b")
func (s *GameService) determineResult(sideToMove, reason string) models.GameResult {
	switch reason {
	case "checkmate":
		if sideToMove == "w" {
			return models.GameResultWhiteWins
		}
		return models.GameResultBlackWins
	case "stalemate", "threefold_repetition", "fifty_move_rule", "insufficient_material", "draw_agreement":
		return models.GameResultDraw
	}
	return models.GameResultDraw
}
