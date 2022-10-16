package main

import (
	"fmt"
	"github.com/google/uuid"
)

const TargetScore = 60

const BoardHeight = 6

// EmptyStrategyBoard Arrays cannot be constants in Go, so the following function is pro forma:
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

type Colour int

const (
	RED Colour = iota
	GREEN
	BLUE
)

func (p Colour) String() string {
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

func (p Colour) toInt() int {
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
	GameId       uuid.UUID         `json:"gameId"`
	Participants map[Colour]Player `json:"participants"`
	State        GameState         `json:"gameState"`
	Moves        []Move            `json:"moves"`
}

type GameState struct {
	ColourInTurn  Colour   `json:"player"`
	StrategyBoard [6][]int `json:"board"`
	ScoreBoard    [3]int   `json:"scores"`
	UnusedPawns   [3]int   `json:"unusedPawns"`
	ForceMovePawn [2]int   `json:"forceMovePawn"`
	AfterTurnNo   int      `json:"afterTurnNo"`
}

func (state GameState) String() string {
	return fmt.Sprintf("State { Turn no #%d / Colour: %s / ForceMove: %v / Unused: %v / Score: %v / Board: %v }",
		state.AfterTurnNo, state.ColourInTurn, state.ForceMovePawn, state.UnusedPawns, state.ScoreBoard, state.StrategyBoard)
}

func (state GameState) Copy() GameState {
	return GameState{
		ColourInTurn: state.ColourInTurn,
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
	Colour Colour   `json:"colour"`
	Board  Board    `json:"board"`
	Path   [][2]int `json:"path"` // Any number of coordinates for the strategy board, but
	// if the Board is SCORE, there will only be one Path [2]int, namely { FROM, TO } on the scoreboard
}

func (m Move) String() string {
	if m.Board == SCORE {
		n := m.Path[0][1] - m.Path[0][0]
		return fmt.Sprintf("Move { Colour: %s / Take %d point(s) }", m.Colour, n)
	}
	return fmt.Sprintf("Move {  Colour: %s / Path: %v}", m.Colour, m.Path)
}

type Player struct {
	Name   string `json:"name"`
	Slogan string `json:"slogan"`
}

func (p Player) String() string {
	return fmt.Sprintf("Player { Id: %s / Slogan: %s }", p.Name, p.Slogan)
}
