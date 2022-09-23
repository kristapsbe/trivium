package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type gameStatus struct {
	Player   int      `json:"player"`
	Board    [6][]int `json:"board"`
	Unused   [3]int   `json:"unused"`
	Scores   [3]int   `json:"scores"`
	MaxScore int      `json:"maxScore"`
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

func validMoves(board [6][]int, player int, unused [3]int, leftScore int) [][4]int {
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
	validDirs := [6][2]int{{-1, 0}, {-1, 1}, {0, -1}, {1, 0}, {1, -1}, {0, 1}}
	for i := range board {
		for j := range board[i] {
			if board[i][j] == player {
				for k := range validDirs {
					deltaI := validDirs[k][0]
					deltaJ := validDirs[k][1]
					// we can move to empty cells
					newI := i + deltaI
					newJ := j + deltaJ
					if newI >= 0 && newJ >= 0 && newI < len(board) && newJ < len(board[newI]) && board[newI][newJ] == 9 {
						retVal = append(retVal, [4]int{i, j, newI, newJ})
					}
					// we can kill opponents pieces
					hopoverI := i + deltaI
					hopoverJ := j + deltaJ
					targetI := i + (2 * deltaI)
					targetJ := j + (2 * deltaJ)
					for targetI >= 0 && targetJ >= -1 && ((targetI < len(board) && targetJ <= len(board[targetI])) || (targetI == len(board) && hopoverJ == 0)) &&
						(targetI == len(board) || targetJ == len(board[targetI]) || targetJ == -1 || board[targetI][targetJ] == 9) &&
						board[hopoverI][hopoverJ] != 9 && board[hopoverI][hopoverJ] != player {
						retVal = append(retVal, [4]int{i, j, targetI, targetJ})
						hopoverI = targetI + deltaI
						hopoverJ = targetJ + deltaJ
						targetI += (2 * deltaI)
						targetJ += (2 * deltaJ)
					}
				}
			}
		}
	}
	return retVal
}

func doMove(currStatus gameStatus) gameStatus {
	valMoves := validMoves(currStatus.Board, currStatus.Player, currStatus.Unused, currStatus.MaxScore-currStatus.Scores[currStatus.Player])
	moveInd := rand.Intn(len(valMoves))
	currI := valMoves[moveInd][0]
	currJ := valMoves[moveInd][1]
	nextI := valMoves[moveInd][2]
	nextJ := valMoves[moveInd][3]
	if currI == 9 {
		if nextI == 9 {
			currStatus.Scores[currStatus.Player] += movePoints(currStatus.Board, currStatus.Player)
		} else {
			currStatus.Unused[currStatus.Player]--
			currStatus.Board[nextI][nextJ] = currStatus.Player
		}
	} else {
		currStatus.Board[currI][currJ] = 9
		if nextI < 0 || nextI >= len(currStatus.Board) || nextJ < 0 || nextJ >= len(currStatus.Board[nextI]) {
			currStatus.Unused[currStatus.Player]++
		} else {
			currStatus.Board[nextI][nextJ] = currStatus.Player
		}
		if abs(currI-nextI) > 1 || abs(currJ-nextJ) > 1 {
			deltaI := getDelta(currI, nextI)
			deltaJ := getDelta(currJ, nextJ)
			tempI := currI + deltaI
			tempJ := currJ + deltaJ
			for nextI != tempI || nextJ != tempJ {
				if currStatus.Board[tempI][tempJ] != 9 {
					currStatus.Unused[currStatus.Board[tempI][tempJ]]++
				}
				currStatus.Board[tempI][tempJ] = 9
				tempI += deltaI
				tempJ += deltaJ
			}
		}
	}
	currStatus.Player = (currStatus.Player + 1) % 3

	return currStatus
}

func fullGame() {
	rand.Seed(time.Now().UnixNano())

	currStatus := gameStatus{
		Player: 0,
		Board: [6][]int{
			{9, 9, 9, 9, 9, 9},
			{9, 9, 9, 9, 9},
			{9, 9, 9, 9},
			{9, 9, 9},
			{9, 9},
			{9},
		},
		Unused:   [3]int{3, 3, 3},
		Scores:   [3]int{0, 0, 0},
		MaxScore: 60,
	}
	for currStatus.Scores[0] < currStatus.MaxScore && currStatus.Scores[1] < currStatus.MaxScore && currStatus.Scores[2] < currStatus.MaxScore {
		currStatus = doMove(currStatus)
		printBoard(currStatus.Board)
		printScores(currStatus.Scores, currStatus.MaxScore)
	}
}

func main() {
	router := gin.Default()

	//fullGame()

	router.GET("/board", renderBoard)
	router.POST("/availableMoves", availableMoves)
	router.POST("/botMove", botMove)

	router.Run("localhost:8080")
}

func renderBoard(c *gin.Context) {
	fileContent, err := ioutil.ReadFile("board.html")
	if err != nil {
		c.AbortWithStatus(404)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", fileContent)
}

func availableMoves(c *gin.Context) {
	var currStatus gameStatus
	if err := c.BindJSON(&currStatus); err != nil {
		return
	}
	c.IndentedJSON(http.StatusOK, validMoves(currStatus.Board, currStatus.Player, currStatus.Unused, currStatus.MaxScore-currStatus.Scores[currStatus.Player]))
}

func botMove(c *gin.Context) {
	var currStatus gameStatus
	if err := c.BindJSON(&currStatus); err != nil {
		return
	}
	c.IndentedJSON(http.StatusOK, doMove(currStatus))
}
