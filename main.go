package main

import (
	"fmt"
	"math/rand"
	"time"
)

func printBoard(board [][]int) {
	for i := range board {
		fmt.Printf("%v -> ", i)
		for j := range board[i] {
			fmt.Printf("%v ", board[i][j])
		}
		fmt.Printf("\n")
	}
}

func printScores(scores [3]int, maxScore int) {
	for v := range scores {
		fmt.Printf("%v -> %v/%v\n", v, scores[v], maxScore)
	}
}

func movePoints(board [][]int, player int) int {
	for i := len(board) - 1; i >= 0; i-- {
		for j := range board[i] {
			if board[i][j] == player {
				return (i + 1)
			}
		}
	}
	return 0
}

func validMoves(board [][]int, player int, unused [3]int, leftScore int) [][4]int {
	retVal := [][4]int{}
	// we can add new pieces
	if unused[player] > 0 {
		for i := range board[0] {
			// can only move to an empty cell
			if board[0][i] == 9 {
				retVal = append(retVal, [4]int{9, 9, 0, i})
			}
		}
	}
	smallEnough := movePoints(board, player) <= leftScore
	// we have at least one piece on the board - can just take points
	if unused[player] < 3 && smallEnough {
		retVal = append(retVal, [4]int{9, 9, 9, 9})
	}
	// we can move to empty cells
	for i := range board {
		for j := range board[i] {
			if board[i][j] == player {
				if i > 0 && board[i-1][j] == 9 {
					retVal = append(retVal, [4]int{i, j, i - 1, j})
				}
				if i > 0 && j < len(board[i-1])-1 && board[i-1][j+1] == 9 {
					retVal = append(retVal, [4]int{i, j, i - 1, j + 1})
				}
				if j > 0 && board[i][j-1] == 9 {
					retVal = append(retVal, [4]int{i, j, i, j - 1})
				}
				if smallEnough && i < len(board)-1 && j < len(board[i+1])-1 && board[i+1][j] == 9 {
					retVal = append(retVal, [4]int{i, j, i + 1, j})
				}
				if smallEnough && i < len(board)-1 && j > 0 && board[i+1][j-1] == 9 {
					retVal = append(retVal, [4]int{i, j, i + 1, j - 1})
				}
				if j < len(board[i])-1 && board[i][j+1] == 9 {
					retVal = append(retVal, [4]int{i, j, i, j + 1})
				}
			}
		}
	}
	return retVal
}

func main() {
	rand.Seed(time.Now().UnixNano())

	maxScore := 60
	scores := [3]int{0, 0, 0}
	unused := [3]int{3, 3, 3}
	board := [][]int{
		{9, 9, 9, 9, 9, 9},
		{9, 9, 9, 9, 9},
		{9, 9, 9, 9},
		{9, 9, 9},
		{9, 9},
		{9},
	}

	currPlayer := 0
	for scores[0] < maxScore && scores[1] < maxScore && scores[2] < maxScore {
		valMoves := validMoves(board, currPlayer, unused, maxScore-scores[currPlayer])
		moveInd := rand.Intn(len(valMoves))
		// do move - either move piece or increase score
		fmt.Printf("\n%v: [%v, %v] -> [%v, %v]\n", currPlayer, valMoves[moveInd][0], valMoves[moveInd][1], valMoves[moveInd][2], valMoves[moveInd][3])
		if valMoves[moveInd][0] == 9 {
			if valMoves[moveInd][2] == 9 {
				scores[currPlayer] += movePoints(board, currPlayer)
			} else {
				unused[currPlayer] -= 1
				board[valMoves[moveInd][2]][valMoves[moveInd][3]] = currPlayer
			}
		} else {
			board[valMoves[moveInd][0]][valMoves[moveInd][1]] = 9
			board[valMoves[moveInd][2]][valMoves[moveInd][3]] = currPlayer
		}
		printBoard(board)
		printScores(scores, maxScore)

		currPlayer = (currPlayer + 1) % 3
		//time.Sleep(1 * time.Second)
	}
}
