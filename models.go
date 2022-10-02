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
