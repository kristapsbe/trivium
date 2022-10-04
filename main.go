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
	if state.ForceMovePawn[0] == 9 {
		// we can add new pieces
		if state.Unused[state.Player] > 0 {
			for i := range state.StrategyBoard[0] {
				// can only move to an empty cell
				if state.StrategyBoard[0][i] == 9 {
					path := []int{9, 9, 0, i}
					moves = append(moves, Move{state.Player, STRATEGY, path})
				}
			}
		}

		if state.Unused[state.Player] < 3 {
			// we have at least one piece on the board - can we just take points?
			if movePoints(state.StrategyBoard, state.Player) <= TargetScore-state.ProgressBoard[state.Player] {
				path := []int{9, 9, 9, 9}
				moves = append(moves, Move{state.Player, PROGRESS, path})
			}
		}
	} else {
		// adding this as a flag of sorts to let the bots stop early in move chains
		path := []int{state.ForceMovePawn[0], state.ForceMovePawn[1], state.ForceMovePawn[0], state.ForceMovePawn[1]}
		moves = append(moves, Move{state.Player, STRATEGY, path})
	}

	validDirections := [6][2]int{{-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, 0}, {1, -1}}
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
							path := []int{y, x, newY, newX}
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
						newY := y + (2 * deltaY)
						newX := x + (2 * deltaX)

						// Only add this as a possibility if ...
						// 1) new position is above bottom of the bord and not two cells off horizontally
						// 2) new position is not above the top cell or out-of-board on the right-hand side OR
						// 3) ... new position is above the top of the board and the eliminated pawn is at position {y,0}
						// 4) new position is above the top of the board OR
						// 5) ... new position is out-of-board on the right-hand side OR
						// 6) ... new position is out-of-board on the left-hand side or new position isn't occupied
						// 7) the cell we jump over is occupied (by anyone, even ourselves)
						if newY >= 0 && newX >= -1 &&
							((newY < BoardHeight && newX <= len(state.StrategyBoard[newY])) ||
								(newY == BoardHeight && gonerX == 0)) &&
							(newY == BoardHeight ||
								newX == len(state.StrategyBoard[newY]) ||
								newX == -1 || state.StrategyBoard[newY][newX] == 9) &&
							state.StrategyBoard[gonerY][gonerX] != 9 {

							// Now get rid of that pesky pawn
							path := []int{y, x, newY, newX}
							moves = append(moves, Move{state.Player, STRATEGY, path})

							// Create a new game state for this jump:
							followingGameState := state
							// Remove the pawn we just jumped over. Remember to put it back among the unused pawns:
							followingGameState.Unused[followingGameState.StrategyBoard[gonerY][gonerX]]++
							followingGameState.StrategyBoard[gonerY][gonerX] = 9
							// remove ourselves:
							followingGameState.StrategyBoard[y][x] = 9
							if newX == -1 || newX == len(state.StrategyBoard[newY]) || newY == BoardHeight {
								// We jumped off of the board!, get us into the Unused:
								followingGameState.Unused[state.Player.toInt()]++
							} else {
								// Insert us at new position:
								followingGameState.StrategyBoard[newY][newX] = state.Player.toInt()
							}
							// Finally, let's make it clear that we want follow-ups from a jump:
							followingGameState.ForceMovePawn = [2]int{newY, newX}

							// Now that we have imagined what the board would look like with this move, let's get its
							// followup alternatives. All of those moves depart from the wrong state, so we append
							// them to the current:
							followingMoves := validMoves(followingGameState)
							for i := 0; i < len(followingMoves); i++ {
								p := followingMoves[i].Path
								if len(p) > 3 && p[0] == p[2] && p[1] == p[3] {
									moves = append(moves, Move{state.Player, STRATEGY, []int{p[0], p[1]}})
									fmt.Printf("Adding: %s\n", []int{p[0], p[1]})
								} else {
									mPath := append(path, p...)
									moves = append(moves, Move{state.Player, STRATEGY, mPath})
									fmt.Printf("Adding: %s\n", mPath)
								}
								fmt.Printf("So we now have: %s\n", moves)
								fmt.Printf("---\n")
							}

							// This is where we should recalculate the board state and do a recursive call,
							// so we get a full jumping path and not just the first move ...
						}
					}
				}
			}
		}
	}
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
		Unused:        [3]int{3, 3, 3},
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
