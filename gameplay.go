package main

import (
	"fmt"
	"os"
)

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func availableScorePoints(strategyBoard [6][]int, scoreBoard [3]int, colour Colour) int {
	// The final move has to end _exactly_ at TargetScore. (Can't move 6 to go from 57 to 60.)
	for y := len(strategyBoard) - 1; y >= 0; y-- {
		if y <= TargetScore-scoreBoard[colour] {
			for x := range strategyBoard[y] {
				if Colour(strategyBoard[y][x]) == colour {
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

// Increments the AfterTurnNo variable and sets the PlayerInTurn variable accordingly
func moveTurn(state *GameState) {
	state.AfterTurnNo++
	state.ColourInTurn = Colour(state.AfterTurnNo % 3)
}

// Given a game state and a move, calculates and returns the following game state
func move(state GameState, move Move) (GameState, error) {

	if state.ColourInTurn != move.Colour {
		return state, fmt.Errorf("%v player is in turn (not %v)", state.ColourInTurn, move.Colour)
	}

	if !move.isIn(validMoves(state)) {
		return state, fmt.Errorf("move %v is invalid in the current state", move.Path)
	}

	// Create the game state representing the state after this jump:
	followingGameState := state.Copy()

	if move.Board == SCORE {
		// Grab the points:
		followingGameState.ScoreBoard[move.Colour] = move.Path[0][1]
		moveTurn(&followingGameState)
		return followingGameState, nil
	}

	// If we get here, we need to move.

	if move.Path[0][0] == 9 && move.Path[0][1] == 9 {
		// We're entering the board
		followingGameState.StrategyBoard[move.Path[1][0]][move.Path[1][1]] = move.Colour.toInt()
		followingGameState.UnusedPawns[move.Colour]--
		moveTurn(&followingGameState)
		return followingGameState, nil
	}

	originY := move.Path[0][0]
	originX := move.Path[0][1]

	for i := 1; i < len(move.Path); i++ {
		destinationY := move.Path[i][0]
		destinationX := move.Path[i][1]
		deltaX := destinationX - originX
		deltaY := destinationY - originY

		if deltaY < 2 && abs(deltaX) < 2 {
			// we are not jumping over a pawn. In this case, we're not either allowed to leave
			// the board, so we can now just move the pawn and go on to the next path element.
			followingGameState.StrategyBoard[originY][originX] = 9
			followingGameState.StrategyBoard[destinationY][destinationX] = move.Colour.toInt()

			// no jumping took place, so the pawn cannot move further.
			// we can simply return from this method:
			moveTurn(&followingGameState)
			return followingGameState, nil
		}

		// We seem to have jumped over a pawn. Let's find it:
		var gonerY, gonerX int
		if deltaY == 2 {
			// Jumping upwards.
			gonerY = originY + 1

			// destinationX will be same as now or 2 less.
			if deltaX == 0 {
				// If jumping towards the right, the gonerX will be equal to originX
				gonerX = originX
			} else {
				// If jumping towards the left, the gonerX will be 1 less than originX.
				gonerX = originX - 1
			}
		} else {
			// Since one cannot jump downwards, deltaY must now be 0.
			// deltaX can then be 2 or -2, and gonerX will be half of that
			gonerX = deltaX / 2
		}

		// Now first augment the number of unused pawns for the concerned player:
		followingGameState.UnusedPawns[followingGameState.StrategyBoard[gonerY][gonerX]]++
		// ... and then eliminate the pawn we jumped over:
		followingGameState.StrategyBoard[gonerY][gonerX] = 9

		// Check if we jumped off of the board:
		if isInLimbo([2]int{destinationY, destinationX}) {
			followingGameState.UnusedPawns[move.Colour.toInt()]++
		} else {
			followingGameState.StrategyBoard[destinationY][destinationX] = move.Colour.toInt()
		}

		// free up the cell we just left:
		followingGameState.StrategyBoard[originY][originX] = 9

		// and finally (before the next pass in the loop),
		// set the new origin for the next jump (if any)
		originY = destinationY
		originX = destinationX

	} // loop end

	moveTurn(&followingGameState)
	return followingGameState, nil
}

func (m Move) isIn(validMoves []Move) bool {
outerLoop:
	for _, validMove := range validMoves {
		if len(validMove.Path) != len(m.Path) ||
			validMove.Colour != m.Colour ||
			validMove.Board != m.Board {
			continue
		}

		for i, p := range validMove.Path {
			if p != m.Path[i] {
				continue outerLoop
			}
		}
		// last statement in outerLoop. If we've made it through to here,
		// this validMove item is equal to m, so:
		return true
	}
	// got through the whole slice without finding m
	return false
}

func validMoves(state GameState) []Move {
	var moves []Move

	// If ForceMovePawn[0] is 9, we're not in the middle of a jump path
	if state.ForceMovePawn[0] == 9 {

		if state.UnusedPawns[state.ColourInTurn] > 0 {
			// Ah, we can add new pieces to the strategy board
			for i := range state.StrategyBoard[0] {
				// can only move to an empty cell on the bottom level
				if state.StrategyBoard[0][i] == 9 {
					moves = append(moves, Move{state.ColourInTurn, STRATEGY, [][2]int{{9, 9}, {0, i}}})
				}
			}
		}

		if state.UnusedPawns[state.ColourInTurn] < 3 {
			// we have at least one piece on the board already - can we just take points?
			availableScorePoints := availableScorePoints(state.StrategyBoard, state.ScoreBoard, state.ColourInTurn)
			if availableScorePoints > 0 {
				// The final move has to end _exactly_ at TargetScore. (Can't move 6 to go from 57 to 60.)
				currentScore := state.ScoreBoard[state.ColourInTurn]
				moves = append(moves, Move{state.ColourInTurn, SCORE, [][2]int{{currentScore, currentScore + availableScorePoints}}})
				// path coordinates 9,9,9,9 is shorthand for "take points on the score board"
			}
		}
	}

	if state.UnusedPawns[state.ColourInTurn] == 3 {
		// All pawns are off of the board. No need to look for further moves.
		return moves
	}

	validDirections := [6][2]int{{-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, 0}, {1, -1}}
	// Now iterate over the cells on the strategy board and look for our pawns:
	for y := range state.StrategyBoard {
		for x := range state.StrategyBoard[y] {
			if Colour(state.StrategyBoard[y][x]) == state.ColourInTurn {
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
							moves = append(moves, Move{state.ColourInTurn, STRATEGY, path})
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
							moves = append(moves, Move{state.ColourInTurn, STRATEGY, jumpPath})

							// Create the game state representing the state after this jump:
							followingGameState := state.Copy()

							// Remove the pawn we just jumped over. Remember to put it back among the unused pawns.
							// We put it back on the limbo line first, before we empty the board cell, while we
							// still have the correct player value in that cell:
							followingGameState.UnusedPawns[followingGameState.StrategyBoard[gonerY][gonerX]]++
							followingGameState.StrategyBoard[gonerY][gonerX] = 9
							// remove ourselves from our previous position:
							followingGameState.StrategyBoard[y][x] = 9
							if jumpToX == -1 || jumpToY == BoardHeight || jumpToX == len(state.StrategyBoard[jumpToY]) {
								// We jumped off of the board!, get us into the set of UnusedPawns pawns:
								followingGameState.UnusedPawns[state.ColourInTurn.toInt()]++
							} else {
								// Insert us at new position:
								followingGameState.StrategyBoard[jumpToY][jumpToX] = state.ColourInTurn.toInt()
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
								moves = append(moves, Move{state.ColourInTurn, STRATEGY, nextJumpPath})
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

// This was added to track down a bug. It might no longer be needed
func checkPawnCounts(state GameState) {
	var y, x, p int
	for p = 0; p < 3; p++ {
		var pawnCount = state.UnusedPawns[p]
		for y = 0; y < BoardHeight; y++ {
			for x = 0; x < len(state.StrategyBoard[y]); x++ {
				if p == state.StrategyBoard[y][x] {
					pawnCount++
				}
			}
		}
		if pawnCount != 3 {
			fmt.Printf("Well, well, well, %s seems to have %d pawns, all of a sudden!\n", Colour(p), pawnCount)
			os.Exit(10)
		}
	}
}
