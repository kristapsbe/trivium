package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var gameState GameState

func TestMain(m *testing.M) {
	fmt.Print("OK, then, let's do those tests!\n")

	gameBoard := EmptyStrategyBoard()
	// Populate our board so the tests have something to work on:
	gameBoard[0][1] = 0 // red is at the bottom row
	gameBoard[0][2] = 0 // red is at the bottom row
	gameBoard[0][5] = 0 // red is at the bottom row

	gameBoard[4][0] = 1 // green is at the next-to-top row
	gameBoard[2][0] = 1 // green is at the next-to-top row
	gameBoard[1][1] = 1 // ... and at the next-to-bottom row

	gameBoard[5][0] = 2 // blue is lucky ... or actually ...

	gameState = GameState{
		Player:        RED,
		StrategyBoard: gameBoard,
		ProgressBoard: [3]int{59, 55, 56},
		UnusedPawns:   [3]int{0, 0, 2}, // since we just placed three pawns on the board
		ForceMovePawn: [2]int{9, 9},
		AfterTurnNo:   0,
	}

	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

func TestAbs(t *testing.T) {
	got := abs(-1)
	if got != 1 {
		t.Errorf("Abs(-1) = %d; want 1", got)
	}
}

func TestPointsForTopPawn(t *testing.T) {
	// Red can win:
	redPoints := pointsForTopPawn(gameState.StrategyBoard, 0)
	greenPoints := pointsForTopPawn(gameState.StrategyBoard, 1)
	bluePoints := pointsForTopPawn(gameState.StrategyBoard, 2)

	fmt.Printf("Testing points for top pawn. (Red: %d, Green: %d, Blue: %d)\n",
		redPoints, greenPoints, bluePoints)

	if redPoints != 1 {
		t.Errorf("Red really should get one point.")
	}
	if greenPoints != 5 {
		t.Errorf("Green really should get five points.")
	}
	if bluePoints != 6 {
		t.Errorf("Blue really should get six points.")
	}
}

func TestIsOnBoard(t *testing.T) {
	fmt.Println("Testing the isOnBoard function")

	// These should represent all valid positions:
	validPositions := [21][2]int{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5},
		{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4},
		{2, 0}, {2, 1}, {2, 2}, {2, 3},
		{3, 0}, {3, 1}, {3, 2},
		{4, 0}, {4, 1},
		{5, 0}}

	for i := range validPositions {
		assert.True(t, isOnBoard(validPositions[i]),
			"We believe %v should represent a position on the strategy board!", validPositions[i])
	}

	// Now let's examine som invalid ones:
	if isOnBoard([2]int{4, 8}) {
		t.Errorf("{4, 8} is too far out on the X axis")
	}
	if isOnBoard([2]int{6, 0}) {
		t.Errorf("{6, 0} is too high up on the Y axis")
	}

	if isOnBoard([2]int{-1, 0}) {
		t.Errorf("{-1, 0} is under the board")
	}

	if isOnBoard([2]int{2, -1}) {
		t.Errorf("{2, -1} is to the left of the board")
	}

}

func TestIsInLimbo(t *testing.T) {
	fmt.Println("Testing the isInLimbo function")

	// These should represent all valid positions:
	validPositions := [14][2]int{
		{6, -1}, {6, 0},
		{5, -1}, {5, 1},
		{4, -1}, {4, 2},
		{3, -1}, {3, 3},
		{2, -1}, {2, 4},
		{1, -1}, {1, 5},
		{0, -1}, {0, 6}}

	for i := range validPositions {
		assert.True(t, isInLimbo(validPositions[i]),
			"We believe %v should represent a position on the limbo line!", validPositions[i])
	}

	// Now let's examine som invalid ones:
	if isInLimbo([2]int{0, 0}) {
		t.Errorf("{0, 0} is on the board!")
	}

	if isInLimbo([2]int{5, 0}) {
		t.Errorf("{5, 0} is on the board!")
	}

	if isInLimbo([2]int{-1, 0}) {
		t.Errorf("{-1, 0} is below the board")
	}

	if isInLimbo([2]int{-2, 0}) {
		t.Errorf("{-2, 0} is far below the board")
	}

	if isInLimbo([2]int{3, -2}) {
		t.Errorf("{3, -2} is too far to the left")
	}

	if isInLimbo([2]int{2, 5}) {
		t.Errorf("{2, 5} is too far to the right")
	}

}

func TestValidMoves(t *testing.T) {
	fmt.Println("Testing the validMoves function")

	// As we enter this method, red is in turn.
	var moves = validMoves(gameState)

	// red should be able to grab a point:
	assert.Contains(t, moves, Move{0, SCORE, [][2]int{{9, 9}, {9, 9}}},
		"Player Red should be able to take a point on the score board")
	assert.Equal(t, 11, len(moves), "There should be 9 options for player Red")

	assert.Contains(t, moves, Move{0, STRATEGY, [][2]int{{0, 1}, {2, 1}, {2, -1}}},
		"Player Red should be able to do a double jump.")

	// Now let's turn to blue
	gameState.Player = BLUE
	// Shouldn't be able to take points:
	moves = validMoves(gameState)
	assert.NotContains(t, moves, Move{0, SCORE, [][2]int{{9, 9}, {9, 9}}},
		"Player Blue shouldn't be able to take a point on the score board")

}
