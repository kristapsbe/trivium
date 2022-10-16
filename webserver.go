package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"math/rand"
	"net/http"
)

func InitializeGameWaitingForPlayers() {
	gameId := uuid.New()
	initialState := GameState{
		ColourInTurn:  RED,
		StrategyBoard: EmptyStrategyBoard(),
		ScoreBoard:    [3]int{0, 0, 0},
		UnusedPawns:   [3]int{3, 3, 3},
		ForceMovePawn: [2]int{9, 9},
		AfterTurnNo:   0,
	}
	participants := map[Colour]Player{}
	GameWaitingForPlayers = Game{gameId, participants, initialState, nil}
	GameWaitingForPlayers.Participants[RED] = invinciBot
	fmt.Printf("Created a new game: %v\n", gameId)
}

func jsonError(err string) map[string]interface{} {
	return map[string]interface{}{"Error": err}
}

func main() {
	r := gin.Default()

	connectToDb()

	r.Static("/static", "web/static")
	r.StaticFile("/board", "web/board.html")
	r.POST("/availableMoves", availableMoves)
	r.POST("/suggestBotMove", suggestBotMove)
	r.POST("/joinGame", joinGame)
	r.POST("/makeMove", makeMove)

	// Before we start the server, let's have a game ready for someone:
	InitializeGameWaitingForPlayers()

	fmt.Println("Starting webserver ...")

	//goland:noinspection GoUnhandledErrorResult
	r.Run(":8080")
	fmt.Println("There!")
}

func makeMove(c *gin.Context) {
	var payload MakeMovePayload
	if err := c.BindJSON(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsonError("We need the game ID and a move to process"))
		return
	}

	gameId, err := uuid.Parse(payload.GameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsonError(err.Error()))
		return
	}

	game, exists := OngoingGames[gameId]
	if !exists {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsonError("The supplied UUID refers to a game that does not or no longer exist"))
		return
	}

	newState, err := move(game.State, payload.Move)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, jsonError(err.Error()))
		return
	}
	game.State = newState

	c.JSON(http.StatusOK, newState)
}

// We should have the same message returned from here no matter how many participants.
// Then keep a websocket and alert all three with the GameWaitingForPlayers once three players are in.
func joinGame(c *gin.Context) {
	var incomingPlayer Player
	if err := c.BindJSON(&incomingPlayer); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "{\"Error\":\"You need to supply some information, sweetheart\"}")
		return
	}
	if incomingPlayer.Name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, "{\"Error\":\"You need to specify your name, sweetheart\"}")
		return
	}

	fmt.Printf("Someone wants to play: %v\n", incomingPlayer)
	fmt.Printf("We already had: %v", GameWaitingForPlayers)

	if len(GameWaitingForPlayers.Participants) > 0 && GameWaitingForPlayers.Participants[0].Name == incomingPlayer.Name {
		c.AbortWithStatusJSON(http.StatusPreconditionFailed, "{\"Error\":\"Player name already taken\"}")
		return
	}
	if len(GameWaitingForPlayers.Participants) > 1 && GameWaitingForPlayers.Participants[1].Name == incomingPlayer.Name {
		c.AbortWithStatusJSON(http.StatusPreconditionFailed, "{\"Error\":\"Player name already taken\"}")
		return
	}

	if len(GameWaitingForPlayers.Participants) == 2 &&
		GameWaitingForPlayers.Participants[RED].Name != incomingPlayer.Name &&
		GameWaitingForPlayers.Participants[GREEN].Name != incomingPlayer.Name {
		fmt.Println("Hey,we already had two other players waiting. Let's play!")
		GameWaitingForPlayers.Participants[BLUE] = incomingPlayer
		fmt.Println("NOW PUSH A MESSAGE TO RED IN ORDER TO START THE GAME")
		c.JSON(http.StatusOK, GameWaitingForPlayers)

		// Now that the message is returned, let's move the waiting-for-players-game into Ongoing:
		gameForMap := GameWaitingForPlayers
		OngoingGames[GameWaitingForPlayers.GameId] = &gameForMap
		// ... and create a new game for coming players:
		InitializeGameWaitingForPlayers()
		return
	} else if len(GameWaitingForPlayers.Participants) == 1 &&
		GameWaitingForPlayers.Participants[0].Name != incomingPlayer.Name {
		GameWaitingForPlayers.Participants[GREEN] = incomingPlayer
		fmt.Println("We now have two participants waiting in this game")
		c.JSON(http.StatusOK, GameWaitingForPlayers)
		return
	}

	// else:
	c.JSON(http.StatusTeapot, "I actually am not a teapot. It's just an April's fool joke.")
}

func availableMoves(c *gin.Context) {
	var reportedState GameState
	if err := c.BindJSON(&reportedState); err != nil {
		return
	}
	valMoves := validMoves(reportedState)
	if len(valMoves) > 0 {
		c.JSON(http.StatusOK, valMoves)
	} else {
		c.JSON(http.StatusOK, [][4]int{})
	}
}

func suggestBotMove(c *gin.Context) {
	var reportedState GameState
	if err := c.BindJSON(&reportedState); err != nil {
		return
	}
	valMoves := validMoves(reportedState)
	fmt.Printf("reportedState: %s\n", reportedState)
	checkPawnCounts(reportedState)
	moveInd := rand.Intn(len(valMoves))
	fmt.Printf("suggested move: %s\n", valMoves[moveInd])
	c.JSON(http.StatusOK, valMoves[moveInd])
}
