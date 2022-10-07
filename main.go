package main

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func availableScorePoints(strategyBoard [6][]int, scoreBoard [3]int, player Player) int {
	// The final move has to end _exactly_ at TargetScore. (Can't move 6 to go from 57 to 60.)
	for y := len(strategyBoard) - 1; y >= 0; y-- {
		if y <= TargetScore-scoreBoard[player] {
			for x := range strategyBoard[y] {
				if Player(strategyBoard[y][x]) == player {
					return y + 1
				}
			}
		}
	}
	return 0
}

// Determines if a given cell coordinate is in the limbo between the outer score board
// and the inner strategy board. This will be the case if the X value is -1 or equal to
// the length of the array representing the given row, or if the Y value is equal to the
// length of the strategy board itself.
func isInLimbo(coordinate [2]int) bool {
	// If Y is within the board, check X:
	if coordinate[0] >= 0 && coordinate[0] < BoardHeight {
		return coordinate[1] == -1 || coordinate[1] == BoardHeight-coordinate[0]
	}

	// So Y is not within board. To be on the limbo now, the cell
	// must have Y equal to BoardHeight and X within the board.
	return coordinate[0] == BoardHeight && (coordinate[1] == -1 || coordinate[1] == -0)
}

func isOnBoard(coordinate [2]int) bool {
	return coordinate[0] >= 0 && coordinate[0] < BoardHeight &&
		coordinate[1] >= 0 && coordinate[1] < BoardHeight-coordinate[0]
}

func isEmpty(coordinate [2]int, state GameState) bool {
	// Instead of handling errors in here as errors,
	// we simply return *false* if the coordinates are off of the strategy board
	return isOnBoard(coordinate) && state.StrategyBoard[coordinate[0]][coordinate[1]] == 9
}

func validMoves(state GameState) []Move {
	var moves []Move

	// If ForceMovePawn[0] is 9, it means this is a request for moving a pawn from limbo onto the board
	if state.ForceMovePawn[0] == 9 {

		if state.UnusedPawns[state.Player] > 0 {
			// Ah, we can add new pieces to the strategy board
			for i := range state.StrategyBoard[0] {
				// can only move to an empty cell on the bottom level
				if state.StrategyBoard[0][i] == 9 {
					moves = append(moves, Move{state.Player, STRATEGY, [][2]int{{9, 9}, {0, i}}})
				}
			}
		}

		if state.UnusedPawns[state.Player] < 3 {
			// we have at least one piece on the board already - can we just take points?
			if availableScorePoints(state.StrategyBoard, state.ScoreBoard, state.Player) > 0 {
				// The final move has to end _exactly_ at TargetScore. (Can't move 6 to go from 57 to 60.)
				moves = append(moves, Move{state.Player, SCORE, [][2]int{{9, 9}, {9, 9}}})
				// path coordinates 9,9,9,9 is shorthand for "take points on the score board"
			}
		}
	}

	if state.UnusedPawns[state.Player] == 3 {
		// All pawns are off of the board. No need to look for further moves.
		return moves
	}

	validDirections := [6][2]int{{-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, 0}, {1, -1}}
	// Now iterate over the cells on the strategy board and look for our pawns:
	for y := range state.StrategyBoard {
		for x := range state.StrategyBoard[y] {
			if Player(state.StrategyBoard[y][x]) == state.Player {
				// Found our player in this cell

				for k := range validDirections {
					deltaY := validDirections[k][0]
					deltaX := validDirections[k][1]
					if state.ForceMovePawn[0] == 9 {
						// we can move to empty cells
						newY := y + deltaY
						newX := x + deltaX
						newCell := [2]int{newY, newX}

						if isEmpty(newCell, state) {
							path := [][2]int{{y, x}, newCell}
							moves = append(moves, Move{state.Player, STRATEGY, path})
						}
					}

					// Now check if we can jump over some adjacent pawn
					// (conditions in the same order as the lines in the underneath if clause):
					// 1) Jump direction is not downwards
					// 2) We're not in the middle of a jump series (so the ForceMove coordinates will be 9,9), OR
					// 3) The current pawn is the one in the middle of a jump series
					// The condition 2/3 is needed in recursive calls to make sure we only consider the "correct" pawn
					if deltaY != -1 &&
						(state.ForceMovePawn[0] == 9 ||
							(state.ForceMovePawn[0] == y && state.ForceMovePawn[1] == x)) {
						gonerY := y + deltaY
						gonerX := x + deltaX
						gonerCell := [2]int{gonerY, gonerX}
						jumpToY := y + (2 * deltaY)
						jumpToX := x + (2 * deltaX)
						jumpToCell := [2]int{jumpToY, jumpToX}

						if !isEmpty(gonerCell, state) && (isEmpty(jumpToCell, state) || isInLimbo(jumpToCell)) {

							// All conditions OK. Here's the path that we will use (more than once):
							jumpPath := [][2]int{{y, x}, jumpToCell}

							// Now get rid of that pesky pawn
							moves = append(moves, Move{state.Player, STRATEGY, jumpPath})

							// Create the game state representing the state after this jump:
							followingGameState := state.Copy()

							// Remove the pawn we just jumped over. Remember to put it back among the unused pawns.
							// We put it back on the limbo line first, before we empty the board cell, while we
							// still have the correct player value in that cell:
							followingGameState.UnusedPawns[followingGameState.StrategyBoard[gonerY][gonerX]]++
							followingGameState.StrategyBoard[gonerY][gonerX] = 9
							// remove ourselves:
							followingGameState.StrategyBoard[y][x] = 9
							if jumpToX == -1 || jumpToY == BoardHeight || jumpToX == len(state.StrategyBoard[jumpToY]) {
								// We jumped off of the board!, get us into the set of UnusedPawns pawns:
								followingGameState.UnusedPawns[state.Player.toInt()]++
							} else {
								// Insert us at new position:
								followingGameState.StrategyBoard[jumpToY][jumpToX] = state.Player.toInt()
							}
							// Finally, let's make it clear that our next request concerns follow-ups from a jump
							// using the ForceMovePawn setting. This setting will only be used in recursive calls:
							followingGameState.ForceMovePawn = [2]int{jumpToY, jumpToX}

							// Now that we have imagined what the board state would be with this move, let's get its
							// followup alternatives:
							// fmt.Printf("we *are* in             this state: %v\n", state)
							// fmt.Printf("we *ask* for moves from this state: %v\n", followingGameState)
							followingMoves := validMoves(followingGameState)
							// fmt.Printf("... and the (recursively retrieved) valid moves seem to be: %v\n", followingMoves)
							// But all of those moves depart from that other board state, so we need to append them to
							// the current:
							for i := 0; i < len(followingMoves); i++ {
								// The first element in the incoming path will now be identical to the last element
								// in jumpPath, so let's omit it and append the rest:
								nextJumpPath := append(jumpPath, followingMoves[i].Path[1:]...)
								moves = append(moves, Move{state.Player, STRATEGY, nextJumpPath})
							}

							// This is where we should recalculate the board state and do a recursive call,
							// so we get a full jumping path and not just the first move ...
						}
					}
				}
			}
		} // end x axis
	} // end y axis
	return moves
}

func main() {
	r := gin.Default()

	r.Static("/static", "web/static")
	r.StaticFile("/board", "web/board.html")
	r.POST("/availableMoves", availableMoves)
	r.POST("/suggestBotMove", suggestBotMove)
	r.GET("/newGame", initializeGame)

	//goland:noinspection GoUnhandledErrorResult
	r.Run(":8080")
}

func initializeGame(c *gin.Context) {
	gameId := uuid.New()
	// fmt.Println("New game: " + gameId.String())
	initialState := GameState{
		Player:        RED,
		StrategyBoard: EmptyStrategyBoard(),
		ScoreBoard:    [3]int{0, 0, 0},
		UnusedPawns:   [3]int{3, 3, 3},
		ForceMovePawn: [2]int{9, 9},
		AfterTurnNo:   0,
	}
	c.JSON(http.StatusOK, Game{gameId.String(), initialState})
}

func availableMoves(c *gin.Context) {
	var currStatus GameState
	if err := c.BindJSON(&currStatus); err != nil {
		return
	}
	valMoves := validMoves(currStatus)
	if len(valMoves) > 0 {
		c.JSON(http.StatusOK, valMoves)
	} else {
		c.JSON(http.StatusOK, [][4]int{})
	}
}

func suggestBotMove(c *gin.Context) {
	var currStatus GameState
	if err := c.BindJSON(&currStatus); err != nil {
		return
	}
	valMoves := validMoves(currStatus)
	fmt.Printf("currStatus: %s\n", currStatus)
	moveInd := rand.Intn(len(valMoves))
	fmt.Printf("suggested move: %s\n", valMoves[moveInd])
	c.JSON(http.StatusOK, valMoves[moveInd])
}
