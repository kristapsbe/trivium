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

func getDelta(currX int, nextX int) int {
	if currX < nextX {
		return 1
	} else if currX > nextX {
		return -1
	}
	return 0
}

func printBoard(board [6][]int) {
	for i := range board {
		fmt.Printf("%v -> ", i)
		for j := range board[i] {
			fmt.Printf("%v ", board[i][j])
		}
		fmt.Printf("\n")
	}
}

func printScores(scores [3]int) {
	for v := range scores {
		fmt.Printf("%v -> %v/%v\n", v, scores[v], TargetScore)
	}
}

func movePoints(board [6][]int, player Player) int {
	for i := len(board) - 1; i >= 0; i-- {
		for j := range board[i] {
			if Player(board[i][j]) == player {
				return i + 1
			}
		}
	}
	return 0
}

func validMoves(state GameState) []Move {
	var moves []Move

	// If ForceMovePawn[0] is 9, it means this is a request for all possible moves for all three pawns
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
			if movePoints(state.StrategyBoard, state.Player) <= TargetScore-state.ProgressBoard[state.Player] {
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

						// Valid move conditions (same order as in the underneath if clause):
						// 1) X and Y coordinates cannot be negative
						// 2) The new Y cannot be above the height of the board
						// 3) The new X cannot be a value larger than the concerned row length, and
						// 4) The target board cell must not already be occupied by another pawn
						if newY >= 0 && newX >= 0 &&
							newY < BoardHeight &&
							newX < len(state.StrategyBoard[newY]) &&
							state.StrategyBoard[newY][newX] == 9 {
							path := [][2]int{{y, x}, {newY, newX}}
							moves = append(moves, Move{state.Player, STRATEGY, path})
						}
					}

					// Now check if we can jump over some adjacent pawn
					// (conditions in the same order as the lines in the underneath if clause):
					// 1) Jump direction is not downwards
					// 2) We've not already made a jump, OR
					// 3) The current pawn (given by current {y,x}) is the one in the middle of a jump series
					if deltaY != -1 &&
						(state.ForceMovePawn[0] == 9 ||
							(state.ForceMovePawn[0] == y && state.ForceMovePawn[1] == x)) {
						gonerY := y + deltaY
						gonerX := x + deltaX
						jumpToY := y + (2 * deltaY)
						jumpToX := x + (2 * deltaX)

						// Only add this as a possibility if ...
						// 1) new position is not below bottom of the bord and not two cells off horizontally
						// 2) new position is not above the top cell or out-of-board on the right-hand side OR
						// 3) ... new position is above the top of the board and the eliminated pawn is at position {y,0}
						// 4) new position is above the top of the board OR
						// 5) ... new position is out-of-board on the right-hand side OR
						// 6) ... new position is out-of-board on the left-hand side or new position isn't occupied
						// 7) the cell we jump over is occupied (by anyone, even ourselves)
						if jumpToY >= 0 && jumpToX >= -1 &&
							((jumpToY < BoardHeight && jumpToX <= len(state.StrategyBoard[jumpToY])) ||
								(jumpToY == BoardHeight && gonerX == 0)) &&
							(jumpToY == BoardHeight ||
								jumpToX == len(state.StrategyBoard[jumpToY]) ||
								jumpToX == -1 || state.StrategyBoard[jumpToY][jumpToX] == 9) &&
							state.StrategyBoard[gonerY][gonerX] != 9 {

							// All conditions OK. Here's tha path that we will use (more than once):
							jumpPath := [][2]int{{y, x}, {jumpToY, jumpToX}}

							// Now get rid of that pesky pawn
							moves = append(moves, Move{state.Player, STRATEGY, jumpPath})

							// Create a new game state for this jump:
							followingGameState := state
							// Remove the pawn we just jumped over. Remember to put it back among the unused pawns:
							followingGameState.UnusedPawns[followingGameState.StrategyBoard[gonerY][gonerX]]++
							followingGameState.StrategyBoard[gonerY][gonerX] = 9
							// remove ourselves:
							followingGameState.StrategyBoard[y][x] = 9
							if jumpToX == -1 || jumpToX == len(state.StrategyBoard[jumpToY]) || jumpToY == BoardHeight {
								// We jumped off of the board!, get us into the set of UnusedPawns pawns:
								followingGameState.UnusedPawns[state.Player.toInt()]++
							} else {
								// Insert us at new position:
								followingGameState.StrategyBoard[jumpToY][jumpToX] = state.Player.toInt()
							}
							// Finally, let's make it clear that our next request concerns follow-ups from a jump
							// (using the ForceMovePawn setting):
							followingGameState.ForceMovePawn = [2]int{jumpToY, jumpToX}

							// Now that we have imagined what the board state would be with this move, let's get its
							// followup alternatives:
							followingMoves := validMoves(followingGameState)
							// But all of those moves depart from that other board state, so we need to append them to
							// the current:
							for i := 0; i < len(followingMoves); i++ {
								// The first element in the incoming path will now be identical to the last element
								// in jumpPath, so let's omit it and append the rest:
								nextJumpPath := append(jumpPath, followingMoves[i].Path[1:]...)
								moves = append(moves, Move{state.Player, STRATEGY, nextJumpPath})
								fmt.Printf("Adding: %s\n", nextJumpPath)
								fmt.Printf("So we now have: %s\n", moves)
								fmt.Printf("---\n")
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
	//fmt.Println("New game: " + gameId.String())
	initialState := GameState{
		Player:        RED,
		StrategyBoard: EmptyStrategyBoard(),
		ProgressBoard: [3]int{0, 0, 0},
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
	moveInd := rand.Intn(len(valMoves))
	c.JSON(http.StatusOK, valMoves[moveInd])
}
