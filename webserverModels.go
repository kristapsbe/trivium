package main

import "github.com/google/uuid"

// Some models that will be in use by the webserver,
// where we often will need to receive som client identification
// alongside the «ordinary» game models

type MakeMovePayload struct {
	GameId string `json:"gameId"`
	Move   Move   `json:"move"`
}

var GameWaitingForPlayers = Game{}

var OngoingGames = make(map[uuid.UUID]*Game)
