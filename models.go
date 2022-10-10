package main

import (
	"fmt"
	"github.com/google/uuid"
)

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

func (b Board) String() string {
	switch b {
	case STRATEGY:
		return "Strategy board"
	case SCORE:
		return "Scoreboard"
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

func (p Player) String() string {
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
	GameId       uuid.UUID        `json:"gameId"`
	State        GameState        `json:"gameState"`
	Participants map[Player]AiBot `json:"participants"`
}

type GameState struct {
	Player        Player   `json:"player"`
	StrategyBoard [6][]int `json:"board"`
	ScoreBoard    [3]int   `json:"scores"`
	UnusedPawns   [3]int   `json:"unusedPawns"`
	ForceMovePawn [2]int   `json:"forceMovePawn"`
	AfterTurnNo   int      `json:"afterTurnNo"`
}

func (state GameState) String() string {
	return fmt.Sprintf("State { Turn no #%d / Player: %s / ForceMove: %v / Unused: %v / Score: %v / Board: %v }",
		state.AfterTurnNo, state.Player, state.ForceMovePawn, state.UnusedPawns, state.ScoreBoard, state.StrategyBoard)
}

func (state GameState) Copy() GameState {
	return GameState{
		Player: state.Player,
		StrategyBoard: [BoardHeight][]int{
			{state.StrategyBoard[0][0], state.StrategyBoard[0][1], state.StrategyBoard[0][2], state.StrategyBoard[0][3], state.StrategyBoard[0][4], state.StrategyBoard[0][5]},
			{state.StrategyBoard[1][0], state.StrategyBoard[1][1], state.StrategyBoard[1][2], state.StrategyBoard[1][3], state.StrategyBoard[1][4]},
			{state.StrategyBoard[2][0], state.StrategyBoard[2][1], state.StrategyBoard[2][2], state.StrategyBoard[2][3]},
			{state.StrategyBoard[3][0], state.StrategyBoard[3][1], state.StrategyBoard[3][2]},
			{state.StrategyBoard[4][0], state.StrategyBoard[4][1]},
			{state.StrategyBoard[5][0]},
		},
		ScoreBoard:    state.ScoreBoard,
		UnusedPawns:   state.UnusedPawns,
		ForceMovePawn: state.ForceMovePawn,
		AfterTurnNo:   state.AfterTurnNo, // not yet implemented
	}
}

type Move struct {
	Player Player
	Board  Board
	Path   [][2]int // Any number of coordinates for the strategy board, but
	// if StrategyBoard is SCORE, there will only be two Path ints: the FROM and the TO
}

func (move Move) String() string {
	if move.Board == SCORE {
		return fmt.Sprintf("Move { Player: %s / Take points }", move.Player)
	}
	return fmt.Sprintf("Move { Player: %s / Path: %v}", move.Player, move.Path)
}

type AiBot struct {
	Name   string `json:"botName"`
	Slogan string `json:"botSlogan"`
}
