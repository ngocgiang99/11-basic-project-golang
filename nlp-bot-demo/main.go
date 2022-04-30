package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/krognol/go-wolfram"
	"github.com/shomali11/slacker"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
)

var wolframClient *wolfram.Client

func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		log.Println("Command Events")
		log.Println(event.Timestamp)
		log.Println(event.Command)
		log.Println(event.Parameters)
		log.Println(event.Event)
		log.Println()
	}
}

func main() {
	err := godotenv.Load("secret.env")
	if err != nil {
		log.Fatal("Error when loading godotenv: ")
	}
	log.Println(os.Getenv("SLACK_BOT_TOKEN"))
	log.Println(os.Getenv("SLACK_APP_TOKEN"))
	log.Println(os.Getenv("WIT_AI_TOKEN"))
	log.Println(os.Getenv("WOLFRAM_APP_ID"))

	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"))
	client := witai.NewClient(os.Getenv("WIT_AI_TOKEN"))
	wolframClient := &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	go printCommandEvents(bot.CommandEvents())
	bot.Command("query for bot - <message>", &slacker.CommandDefinition{
		Description: "Send question to wolfram",
		Example: "Answer of 1+1",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			query := request.Param("message")
			log.Println("Param query", query)
			msg, _ := client.Parse(&witai.MessageRequest{
				Query: query,
			})
			data, _ := json.MarshalIndent(msg, "", "    ")
			rough := string(data[:])
			value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
			answer := value.String()
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)
			if err != nil {
				log.Println("there is an error", err)
			}

			log.Println("Param query", value)


			response.Reply(res)

		},
	})


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}

}