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
	fmt.Println("Starting Peril server...")

	const connectionString string = "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal("connection couldnt be dialed:", err)
	}
	defer connection.Close()
	fmt.Println("Connection was successful")

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal("Channel couldnt be created:", err)
	}

	_, _, err = pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.Durable)
	if err != nil {
		log.Fatal("connection couldnt be bind:", err)
	}

	gamelogic.PrintServerHelp()

	for {
		inputs := gamelogic.GetInput()

		if inputs == nil {
			continue
		} else if inputs[0] == "pause" {

			fmt.Println("sending the pause message")

			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				log.Fatal("couldnt publish json:", err)
			}
			fmt.Println("Pause message sent!")

		} else if inputs[0] == "resume" {

			fmt.Println("sending the resume message")

			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Fatal("couldnt publish json:", err)
			}
			fmt.Println("resume message sent!")

		} else if inputs[0] == "quit" {

			fmt.Println("exiting")
			return

		} else {

			fmt.Println("couldnt understand the command")
			continue
		}

	}

}
