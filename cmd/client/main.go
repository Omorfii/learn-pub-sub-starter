package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	const connectionString string = "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal("connection couldnt be dialed:", err)
	}
	defer connection.Close()
	fmt.Println("Connection was successful")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal("username couldnt be retrieve", err)
	}

	gamestate := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, "pause."+username, routing.PauseKey, pubsub.Transient, handlerPause(gamestate))
	if err != nil {
		log.Fatal("couldnt subscribe to json:", err)
	}
	err = pubsub.SubscribeJSON(connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(gamestate))
	if err != nil {
		log.Printf("couldnt subscribe to move unit, err: %v", err)
		return
	}
	publishCh, err := connection.Channel()
	if err != nil {
		log.Fatal("couldnt t create channel:", err)
	}

	for {
		inputs := gamelogic.GetInput()
		if len(inputs) == 0 {
			continue
		}
		switch inputs[0] {
		case "spawn":
			err = gamestate.CommandSpawn(inputs)
			if err != nil {
				log.Printf("couldnt spawn unit, err: %v", err)

			}
		case "move":
			move, err := gamestate.CommandMove(inputs)
			if err != nil {
				log.Printf("couldnt move unit: %v", err)

			}
			log.Printf("move was successfull: %v", move)

			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, move)
			if err != nil {
				log.Printf("couldnt publish move: %v", err)
				return
			}
			log.Printf("move was published successfully")

		case "status":
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			log.Println("Spamming is not allowed yet")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
