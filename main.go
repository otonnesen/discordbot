package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type config struct {
	Token  string `json:"token"`
	Prefix string `json:"prefix"`
}

func main() {
	lock := make(chan int)

	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(file)
	config := config{}
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", config.Token))

	if err != nil {
		panic(err)
	}

	_, err = discord.Guild("133767800090001408")
	if err != nil {
		panic(err)
	}

	discord.AddHandler(func(session *discordgo.Session, message *discordgo.MessageCreate) {
		if me, _ := session.User("@me"); me.ID == message.Author.ID {
			return
		}

		if strings.Index(message.Content, config.Prefix) != 0 {
			return
		}
		args := strings.Fields(message.Content)
		command, args := args[0][1:], args[1:]
		fmt.Printf("%s\n%v\n", command, args)

		switch command {
		case "say":
			del := " "
			if args[0] == "-d" && len(args) >= 2 {
				del, args = args[1], args[2:]
			}
			session.ChannelMessageSend(message.ChannelID, strings.Join(args, del))
		case "ping":
			session.ChannelMessageSend(message.ChannelID, "Pong!")
		case "shutdown":
			session.ChannelMessageSend(message.ChannelID, "Shutting down...")
			discord.Close()
			os.Exit(0)
		}
	})

	err = discord.Open()

	if err != nil {
		panic(err)
	}

	<-lock
}
