package main

import "fmt"

const TargetScore = 60

const BoardHeight = 6

// Arrays cannot be constants in Go, so the following function is pro forma:

func EmptyStrategyBoard() [BoardHeight][]int {
	return [BoardHeight][]int{{9, 9, 9, 9, 9, 9}, {9, 9, 9, 9, 9}, {9, 9, 9, 9}, {9, 9, 9}, {9, 9}, {9}}
}

type Board int

const (
	STRATEGY Board = iota
	SCORE
)

func (b Board) toString() string {
	switch b {
	case STRATEGY:
		return "Strategy board"
	case SCORE:
		return "Progress board"
	default:
		return fmt.Sprintf("Unknown(%d)", b)
	}
}

type Player int

const (
	RED Player = iota
	GREEN
	BLUE
)

func (p Player) toString() string {
	switch p {
	case RED:
		return "Red"
	case GREEN:
		return "Green"
	case BLUE:
		return "Blue"
	default:
		return fmt.Sprintf("Unknown(%d)", p)
	}
}

func (p Player) toInt() int {
	switch p {
	case RED:
		return 0
	case GREEN:
		return 1
	case BLUE:
		return 2
	default:
		panic("No such player!")
	}
}

type Game struct {
	GameId string    `json:"gameId"`
	State  GameState `json:"GameState"`
}

type GameState struct {
	Player        Player   `json:"player"`
	StrategyBoard [6][]int `json:"board"`
	ProgressBoard [3]int   `json:"scores"`
	UnusedPawns   [3]int   `json:"unusedPawns"`
	ForceMovePawn [2]int   `json:"forceMovePawn"`
	AfterTurnNo   int      `json:"afterTurnNo"`
}

type Move struct {
	Player Player
	Board  Board
	Path   [][2]int // Any number of coordinates for the strategy board, but
	// if StrategyBoard is SCORE, there will only be two Path ints: the FROM and the TO
}
