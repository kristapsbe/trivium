package main

import (
	"fmt"
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
		fmt.Printf("%v -> %v/%v \n", v, scores[v], maxScore)
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

func validMoves(board [][]int, player int, unused [3]int) [][4]int {
	retVal := [][4]int{
		{-1, -1, -1, -1},
	}
	if unused[player] > 0 {
		retVal = append(retVal, [4]int{-1, -1, -1, -1})
	}
	return retVal
}

func main() {
	maxScore := 60
	scores := [3]int{0, 15, 17}
	unused := [3]int{3, 3, 3}
	board := [][]int{
		{-1, -1, -1, -1, -1, -1},
		{-1, -1, -1, -1, -1},
		{-1, -1, -1, -1},
		{-1, -1, -1},
		{-1, -1},
		{-1},
	}

	printScores(scores, maxScore)
	printBoard(board)
	fmt.Printf("%v", movePoints(board, -1))

	for scores[0] < maxScore && scores[1] < maxScore && scores[2] < maxScore {
		validMoves(board, 0, unused)
		// do move - either move piece or increase score
		scores[0] += 1
		scores[1] += 1
		scores[2] += 1
		printScores(scores, maxScore)
	}
}
