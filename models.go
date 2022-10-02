package main

const TARGET_SCORE = 60

type Game struct {
	GameId string    `json:"gameId"`
	State  GameState `json:"GameState"`
}

type GameState struct {
	Player      int      `json:"player"`
	Board       [6][]int `json:"board"`
	Unused      [3]int   `json:"unused"`
	Scores      [3]int   `json:"scores"`
	ForceMove   [2]int   `json:"forceMove"`
	AfterTurnNo int      `json:"afterTurnNo"`
}

type Board int

const (
	STRATEGY Board = 0
	PROGRESS Board = 1
)

type Move struct {
	Player int
	Board  Board
	Path   []int // Any number of coordinates for the strategy board, but
	// if Board is PROGRESS, there will only be two Path ints: the FROM and the TO
}
