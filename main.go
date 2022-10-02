package main

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const TARGET_SCORE = 60

type Game struct {
	GameId string    `json:"gameId"`
	State  gameState `json:"gameState"`
}

type gameState struct {
	Player      int      `json:"player"`
	Board       [6][]int `json:"board"`
	Unused      [3]int   `json:"unused"`
	Scores      [3]int   `json:"scores"`
	ForceMove   [2]int   `json:"forceMove"`
	AfterTurnNo int      `json:"afterTurnNo"`
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

func printScores(scores [3]int) {
	for v := range scores {
		fmt.Printf("%v -> %v/%v\n", v, scores[v], TARGET_SCORE)
	}
}

func movePoints(board [6][]int, player int) int {
	for i := len(board) - 1; i >= 0; i-- {
		for j := range board[i] {
			if board[i][j] == player {
				return i + 1
			}
		}
	}
	return 0
}

func validMoves(state GameState) [][]int {
	var retVal [][]int
	if state.ForceMove[0] == 9 {
		// we can add new pieces
		if state.Unused[state.Player] > 0 {
			for i := range state.Board[0] {
				// can only move to an empty cell
				if state.Board[0][i] == 9 {
					retVal = append(retVal, []int{9, 9, 0, i, 18, 19, 20, 21})
				}
			}
		}

		if state.Unused[state.Player] < 3 {
			// we have at least one piece on the board - can we just take points?
			if movePoints(state.Board, state.Player) <= TARGET_SCORE-state.Scores[state.Player] {
				retVal = append(retVal, []int{9, 9, 9, 9, 18, 19, 20, 21})
			}
		}
	} else {
		// adding this as a flag of sorts to let the bots stop early in move chains
		retVal = append(retVal, []int{state.ForceMove[0], state.ForceMove[1], state.ForceMove[0], state.ForceMove[1], 18, 19, 20, 21})
	}

	validDirections := [6][2]int{{-1, 0}, {-1, 1}, {0, -1}, {1, 0}, {1, -1}, {0, 1}}
	for i := range state.Board {
		for j := range state.Board[i] {
			if state.Board[i][j] == state.Player {
				for k := range validDirections {
					deltaI := validDirections[k][0]
					deltaJ := validDirections[k][1]
					if state.ForceMove[0] == 9 {
						// we can move to empty cells
						newI := i + deltaI
						newJ := j + deltaJ
						if newI >= 0 && newJ >= 0 && newI < len(state.Board) && newJ < len(state.Board[newI]) && state.Board[newI][newJ] == 9 {
							retVal = append(retVal, []int{i, j, newI, newJ, 18, 19, 20, 21})
						}
					}

					// not allowed to jump over pieces downwards
					if deltaI != -1 && (state.ForceMove[0] == 9 || (state.ForceMove[0] == i && state.ForceMove[1] == j)) {
						// we can eliminate opponent pawns
						hopoverI := i + deltaI
						hopoverJ := j + deltaJ
						targetI := i + (2 * deltaI)
						targetJ := j + (2 * deltaJ)
						if targetI >= 0 && targetJ >= -1 && ((targetI < len(state.Board) && targetJ <= len(state.Board[targetI])) || (targetI == len(state.Board) && hopoverJ == 0)) &&
							(targetI == len(state.Board) || targetJ == len(state.Board[targetI]) || targetJ == -1 || state.Board[targetI][targetJ] == 9) && state.Board[hopoverI][hopoverJ] != 9 {
							retVal = append(retVal, []int{i, j, targetI, targetJ, 18, 19, 20, 21})
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
	r.POST("/suggestBotMove", suggestBotMove)
	r.GET("/newGame", initializeGame)

	//goland:noinspection GoUnhandledErrorResult
	r.Run(":8080")
}

func initializeGame(c *gin.Context) {
	gameId := uuid.New()
	//fmt.Println("New game: " + gameId.String())
	initialState := GameState{
		Player:      0,
		Board:       [6][]int{{9, 9, 9, 9, 9, 9}, {9, 9, 9, 9, 9}, {9, 9, 9, 9}, {9, 9, 9}, {9, 9}, {9}},
		Unused:      [3]int{3, 3, 3},
		Scores:      [3]int{0, 0, 0},
		ForceMove:   [2]int{9, 9},
		AfterTurnNo: 0,
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
		c. /*Indented*/ JSON(http.StatusOK, valMoves)
	} else {
		c. /*Indented*/ JSON(http.StatusOK, [][4]int{})
	}
}

func suggestBotMove(c *gin.Context) {
	var currStatus GameState
	if err := c.BindJSON(&currStatus); err != nil {
		return
	}
	valMoves := validMoves(currStatus)
	moveInd := rand.Intn(len(valMoves))
	c. /*Indented*/ JSON(http.StatusOK, valMoves[moveInd])
}
