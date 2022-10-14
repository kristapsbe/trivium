package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
)

func main() {
	r := gin.Default()

	connectToDb()

	r.Static("/static", "web/static")
	r.StaticFile("/board", "web/board.html")
	r.POST("/availableMoves", availableMoves)
	r.POST("/suggestBotMove", suggestBotMove)
	r.POST("/joinGame", joinGame)

	// Before we start the server, let's have a game ready for someone:
	InitializeNewGame()

	//goland:noinspection GoUnhandledErrorResult
	r.Run(":8080")
}

// We should have the same message returned from here no matter how many participants.
// Then keep a websocket and alert all three with the waitingGame once three players are in.
func joinGame(c *gin.Context) {
	var incomingBot AiBot
	if err := c.BindJSON(&incomingBot); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "{\"Error\":\"You need to supply some information, sweetheart\"}")
		return
	}
	if incomingBot.Name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, "{\"Error\":\"You need a bot name, sweetheart\"}")
		return
	}

	fmt.Printf("Someone wants to play: %v\n", incomingBot)
	fmt.Printf("We already had: %v", waitingGame)

	if len(waitingGame.Participants) > 0 && waitingGame.Participants[0].Name == incomingBot.Name {
		c.AbortWithStatusJSON(http.StatusPreconditionFailed, "{\"Error\":\"Bot name already taken\"}")
		return
	}
	if len(waitingGame.Participants) > 1 && waitingGame.Participants[1].Name == incomingBot.Name {
		c.AbortWithStatusJSON(http.StatusPreconditionFailed, "{\"Error\":\"Bot name already taken\"}")
		return
	}

	if len(waitingGame.Participants) == 2 &&
		waitingGame.Participants[RED].Name != incomingBot.Name &&
		waitingGame.Participants[GREEN].Name != incomingBot.Name {
		fmt.Println("Hey,we already had two other players waiting. Let's play!")
		waitingGame.Participants[BLUE] = incomingBot
		ongoingGames = append(ongoingGames, waitingGame)
		fmt.Println("NOW PUSH A MESSAGE TO RED IN ORDER TO START THE GAME")
		c.JSON(http.StatusOK, waitingGame)

		// Now that the message is returned, let's create a new game here:
		InitializeNewGame()
		return
	} else if len(waitingGame.Participants) == 1 &&
		waitingGame.Participants[0].Name != incomingBot.Name {
		waitingGame.Participants[GREEN] = incomingBot
		fmt.Println("We now have two participants waiting in this game")
		c.JSON(http.StatusOK, waitingGame)
		return
	}

	// else:
	c.JSON(http.StatusTeapot, "I actually am not a teapot. It's just an April's fool.")
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

var waitingGame = Game{}
var ongoingGames []Game
