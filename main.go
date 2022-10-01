package main

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

type gameStatus struct {
	Player    int      `json:"player"`
	Board     [6][]int `json:"board"`
	Unused    [3]int   `json:"unused"`
	Scores    [3]int   `json:"scores"`
	MaxScore  int      `json:"maxScore"`
	ForceMove [2]int   `json:"forceMove"`
}

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

func printScores(scores [3]int, maxScore int) {
	for v := range scores {
		fmt.Printf("%v -> %v/%v\n", v, scores[v], maxScore)
	}
}

func movePoints(board [6][]int, player int) int {
	for i := len(board) - 1; i >= 0; i-- {
		for j := range board[i] {
			if board[i][j] == player {
				return (i + 1)
			}
		}
	}
	return 0
}

func validMoves(board [6][]int, player int, unused [3]int, leftScore int, forceMove [2]int) [][4]int {
	retVal := [][4]int{}
	if forceMove[0] == 9 {
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
	} else {
		// adding this as a flag of sorts to let the bots to stop early in move chains
		retVal = append(retVal, [4]int{forceMove[0], forceMove[1], forceMove[0], forceMove[1]})
	}
	validDirs := [6][2]int{{-1, 0}, {-1, 1}, {0, -1}, {1, 0}, {1, -1}, {0, 1}}
	for i := range board {
		for j := range board[i] {
			if board[i][j] == player {
				for k := range validDirs {
					deltaI := validDirs[k][0]
					deltaJ := validDirs[k][1]
					if forceMove[0] == 9 {
						// we can move to empty cells
						newI := i + deltaI
						newJ := j + deltaJ
						if newI >= 0 && newJ >= 0 && newI < len(board) && newJ < len(board[newI]) && board[newI][newJ] == 9 {
							retVal = append(retVal, [4]int{i, j, newI, newJ})
						}
					}
					// not allowed to jump over pieces downwards
					if deltaI != -1 && (forceMove[0] == 9 || (forceMove[0] == i && forceMove[1] == j)) {
						// we can kill opponents pieces
						hopoverI := i + deltaI
						hopoverJ := j + deltaJ
						targetI := i + (2 * deltaI)
						targetJ := j + (2 * deltaJ)
						if targetI >= 0 && targetJ >= -1 && ((targetI < len(board) && targetJ <= len(board[targetI])) || (targetI == len(board) && hopoverJ == 0)) &&
							(targetI == len(board) || targetJ == len(board[targetI]) || targetJ == -1 || board[targetI][targetJ] == 9) && board[hopoverI][hopoverJ] != 9 {
							retVal = append(retVal, [4]int{i, j, targetI, targetJ})
						}
					}
				}
			}
		}
	}
	return retVal
}

func main() {
	r := gin.Default()

	r.Static("/static", "web/static")
	r.StaticFile("/board", "web/board.html")
	r.POST("/availableMoves", availableMoves)
	r.POST("/botMove", botMove)

	r.Run(":8080")
}

func availableMoves(c *gin.Context) {
	var currStatus gameStatus
	if err := c.BindJSON(&currStatus); err != nil {
		return
	}
	valMoves := validMoves(currStatus.Board, currStatus.Player, currStatus.Unused, currStatus.MaxScore-currStatus.Scores[currStatus.Player], currStatus.ForceMove)
	if len(valMoves) > 0 {
		c.IndentedJSON(http.StatusOK, valMoves)
	} else {
		c.IndentedJSON(http.StatusOK, [][4]int{})
	}
}

func botMove(c *gin.Context) {
	var currStatus gameStatus
	if err := c.BindJSON(&currStatus); err != nil {
		return
	}
	valMoves := validMoves(currStatus.Board, currStatus.Player, currStatus.Unused, currStatus.MaxScore-currStatus.Scores[currStatus.Player], currStatus.ForceMove)
	moveInd := rand.Intn(len(valMoves))
	c.IndentedJSON(http.StatusOK, valMoves[moveInd])
}
