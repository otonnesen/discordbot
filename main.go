package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/otonnesen/discordbot/commands"
)

type config struct {
	Token  string `json:"token"`
	Prefix string `json:"prefix"` // Couldn't figure out how to use :|
}

func main() {
	lock := make(chan int)

	file, err := os.Open("./config.json")
	if err != nil {
		panic(err)
	}
	config := config{}
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		panic(err)
	}
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", config.Token))

	if err != nil {
		panic(err)
	}

	discord.AddHandler(commands.Parse)

	err = discord.Open()

	if err != nil {
		panic(err)
	}

	<-lock
}
