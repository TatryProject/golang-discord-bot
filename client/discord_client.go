package client

import (
	"github.com/bwmarrin/discordgo"
)

type DiscordClientConfig struct {
	BotPrefix string `json:"BotPrefix"`
	Token     string `json:"Token"`
}

type DiscordClient struct {
	BotPrefix string
	User      *discordgo.User
	Session   *discordgo.Session
	token     string
}

func NewDiscordClient(token string, botPrefix string) (*DiscordClient, error) {
	botSession, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	user, err := botSession.User("@me")
	if err != nil {
		return nil, err
	}

	client := &DiscordClient{
		BotPrefix: botPrefix,
		Session:   botSession,
		User:      user,
		token:     token,
	}
	return client, nil
}
