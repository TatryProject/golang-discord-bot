package handler

import (
	"errors"
	"fmt"
	"golang-discord-bot/client"
	"golang-discord-bot/image_processing"
	discord_service "golang-discord-bot/service"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type DiscordHandler struct {
	discordClient *client.DiscordClient
}

const TWO_HUNDRED_FIFTY_SIX_KB_IN_BYTES = 262144

func NewDiscordHandler(dc *client.DiscordClient) *DiscordHandler {
	return &DiscordHandler{
		discordClient: dc,
	}
}

func (h *DiscordHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == h.discordClient.User.ID {
		return
	}

	if m.Content == h.discordClient.BotPrefix+"ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "pong")
	}

	if strings.HasPrefix(m.Content, fmt.Sprintf("%semote", h.discordClient.BotPrefix)) {
		handleEmojiAdd(s, m)
	}
}

func handleEmojiAdd(s *discordgo.Session, m *discordgo.MessageCreate) error {
	if len(m.Attachments) == 0 {
		return errors.New("image attachment must be provided")
	}

	_, emoteName, replaceEmoteName := getEmojiAddArguments(strings.Fields(m.Content))
	attachment := m.Attachments[0]
	imgUrl := attachment.URL
	if imgUrl == "" {
		return errors.New("the attached image must have an image URL")
	}

	file, err := image_processing.WriteImageToFile("input.png", imgUrl)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error downloading image: %s", err.Error()))
		return err
	}
	defer file.Close()

	outputPath, err := image_processing.RemoveBackground()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error removing background: %s", err.Error()))
		return err
	}

	// Could cause errors later if width is the min dimension
	newImg, err := image_processing.ResizeImageForDiscord(outputPath, 128, 0)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error resizing image: %s", err.Error()))
		return err
	}
	defer newImg.Close()

	fileInfo, err := newImg.Stat()
	if err != nil {
		fmt.Println(err)
		return err
	}
	if fileInfo.Size() > TWO_HUNDRED_FIFTY_SIX_KB_IN_BYTES {
		errMsg := "provided image is greater than 256KB after resize; please try a smaller image"
		// Resize to smaller size?
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %s", errMsg))
		return errors.New(errMsg)
	}

	if replaceEmoteName != "" {
		err := discord_service.DeleteEmojiByName(replaceEmoteName, s, m)
		if err != nil {
			return err
		}
	}

	newImg.Seek(0, 0)
	discord_service.CreateEmoji(newImg, emoteName, s, m)

	return nil
}

func getEmojiAddArguments(args []string) (string, string, string) {
	switch len(args) {
	case 2:
		return args[0], args[1], ""
	case 3:
		return args[0], args[1], args[2]
	}
	return args[0], "", ""
}
