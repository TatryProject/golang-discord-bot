package main

import (
	"encoding/json"
	"fmt"
	"golang-discord-bot/client"
	handler "golang-discord-bot/handler"
	"os"
)

func main() {
	config, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	discordClient, err := client.NewDiscordClient(config.Token, config.BotPrefix)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	discordHandler := handler.NewDiscordHandler(discordClient)
	discordClient.Session.AddHandler(discordHandler.Handle)

	if err = discordClient.Session.Open(); err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bot is running!")
	<-make(chan struct{})
	defer discordClient.Session.Close()
}

func readConfig() (*client.DiscordClientConfig, error) {
	file, err := os.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	var config *client.DiscordClientConfig
	if err = json.Unmarshal(file, &config); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return config, nil
}
