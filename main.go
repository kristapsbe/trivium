package main

import (
	"fmt"
	"math"
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
	for i := range board {
		for j := range board[i] {
			if board[i][j] == player {
				// we can move to empty cells
				// down
				if i > 0 && board[i-1][j] == 9 {
					retVal = append(retVal, [4]int{i, j, i - 1, j})
				}
				// down-left
				if i > 0 && j < len(board[i-1])-1 && board[i-1][j+1] == 9 {
					retVal = append(retVal, [4]int{i, j, i - 1, j + 1})
				}
				// right
				if j > 0 && board[i][j-1] == 9 {
					retVal = append(retVal, [4]int{i, j, i, j - 1})
				}
				// up
				if i < len(board)-1 && j < len(board[i+1])-1 && board[i+1][j] == 9 {
					retVal = append(retVal, [4]int{i, j, i + 1, j})
				}
				// up-right
				if i < len(board)-1 && j > 0 && board[i+1][j-1] == 9 {
					retVal = append(retVal, [4]int{i, j, i + 1, j - 1})
				}
				// left
				if j < len(board[i])-1 && board[i][j+1] == 9 {
					retVal = append(retVal, [4]int{i, j, i, j + 1})
				}
				// TODO: we can kill
				// down
				if i > 1 && board[i-2][j] == 9 && board[i-1][j] != player && board[i-1][j] != 9 {
					retVal = append(retVal, [4]int{i, j, i - 2, j})
				}
				// down-left
				if i > 1 && j < len(board[i-2])-2 && board[i-2][j+2] == 9 && board[i-1][j+1] != player && board[i-1][j+1] != 9 {
					retVal = append(retVal, [4]int{i, j, i - 2, j + 2})
				}
				// right
				if j > 0 && (j == 1 || board[i][j-2] == 9) && board[i][j-1] != player && board[i][j-1] != 9 {
					retVal = append(retVal, [4]int{i, j, i, j - 2})
				}
				// up
				// up-right
				// left
				if j < len(board[i])-1 && (j == len(board[i])-2 || board[i][j+2] == 9) && board[i][j+1] != player && board[i][j+1] != 9 {
					retVal = append(retVal, [4]int{i, j, i, j + 2})
				}
				// TODO: we can double kill
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
		currI := valMoves[moveInd][0]
		currJ := valMoves[moveInd][1]
		nextI := valMoves[moveInd][2]
		nextJ := valMoves[moveInd][3]
		if currI == 9 {
			if nextI == 9 {
				scores[currPlayer] += movePoints(board, currPlayer)
			} else {
				unused[currPlayer] -= 1
				board[nextI][nextJ] = currPlayer
			}
		} else {
			board[valMoves[moveInd][0]][valMoves[moveInd][1]] = 9
			if nextI < 0 || nextI >= len(board) || nextJ < 0 || nextJ >= len(board[nextI]) {
				unused[currPlayer] += 1
			} else {
				board[nextI][nextJ] = currPlayer
			}
			fmt.Println(math.Abs(float64(currI) - float64(nextI)))
			fmt.Println(math.Abs(float64(currJ) - float64(nextJ)))
			if math.Abs(float64(currI)-float64(nextI)) > 1 || math.Abs(float64(currJ)-float64(nextJ)) > 1 {
				fmt.Println("delete stuff")
				if currI > nextI {
					// TODO: this is wrong
					for tempI := currI + 1; tempI > nextI; tempI-- {
						if currJ > nextJ {
							for tempJ := currJ + 1; tempJ > nextJ; tempJ-- {
								fmt.Printf("%v, %v = %v\n", tempI, tempJ, board[tempI][tempJ])
								if board[tempI][tempJ] != 9 {
									unused[board[tempI][tempJ]] += 1
									board[tempI][tempJ] = 9
								}
							}
						} else {
							for tempJ := currJ + 1; tempJ < int(math.Min(float64(len(board[tempI])), float64(nextJ))); tempJ++ {
								fmt.Printf("%v, %v = %v\n", tempI, tempJ, board[tempI][tempJ])
								if board[tempI][tempJ] != 9 {
									unused[board[tempI][tempJ]] += 1
									board[tempI][tempJ] = 9
								}
							}
						}
					}
				} else {
					for tempI := currI + 1; tempI < nextI; tempI++ {
						if currJ > nextJ {
							for tempJ := currJ + 1; tempJ > nextJ; tempJ-- {
								fmt.Printf("%v, %v = %v\n", tempI, tempJ, board[tempI][tempJ])
								if board[tempI][tempJ] != 9 {
									unused[board[tempI][tempJ]] += 1
									board[tempI][tempJ] = 9
								}
							}
						} else {
							for tempJ := currJ + 1; tempJ < int(math.Min(float64(len(board[tempI])), float64(nextJ))); tempJ++ {
								fmt.Printf("%v, %v = %v\n", tempI, tempJ, board[tempI][tempJ])
								if board[tempI][tempJ] != 9 {
									unused[board[tempI][tempJ]] += 1
									board[tempI][tempJ] = 9
								}
							}
						}
					}
				}
			}
		}
		printBoard(board)
		printScores(scores, maxScore)

		currPlayer = (currPlayer + 1) % 3
		//time.Sleep(1 * time.Second)
	}
}
