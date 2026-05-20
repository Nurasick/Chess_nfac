package chess

import (
	"fmt"

	chesslib "github.com/notnil/chess"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) ValidateMove(fen string, move string) (newFEN string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("chess.Engine.ValidateMove: panic: %v", r)
		}
	}()

	fenOpt, fenErr := chesslib.FEN(fen)
	if fenErr != nil {
		return "", fmt.Errorf("chess.Engine.ValidateMove: invalid FEN: %w", fenErr)
	}

	game := chesslib.NewGame(chesslib.UseNotation(chesslib.UCINotation{}), fenOpt)
	if err := game.MoveStr(move); err != nil {
		return "", fmt.Errorf("chess.Engine.ValidateMove: illegal move: %w", err)
	}

	return game.FEN(), nil
}

func (e *Engine) IsGameOver(fen string) (over bool, reason string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("chess.Engine.IsGameOver: panic: %v", r)
		}
	}()

	fenOpt, fenErr := chesslib.FEN(fen)
	if fenErr != nil {
		return false, "", fmt.Errorf("chess.Engine.IsGameOver: invalid FEN: %w", fenErr)
	}

	game := chesslib.NewGame(fenOpt)
	outcome := game.Outcome()
	if outcome == chesslib.NoOutcome {
		return false, "", nil
	}

	switch game.Method() {
	case chesslib.Checkmate:
		reason = "checkmate"
	case chesslib.Stalemate:
		reason = "stalemate"
	case chesslib.ThreefoldRepetition:
		reason = "threefold_repetition"
	case chesslib.FiftyMoveRule:
		reason = "fifty_move_rule"
	case chesslib.InsufficientMaterial:
		reason = "insufficient_material"
	default:
		reason = "draw_agreement"
	}

	return true, reason, nil
}
