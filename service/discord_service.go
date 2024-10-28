package discord_service

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/bwmarrin/discordgo"
)

func DeleteEmojiByName(name string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	emojis, err := s.GuildEmojis(m.Message.GuildID)
	if err != nil {
		fmt.Println("Error getting emojis:", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error getting emojis: %s", err.Error()))
		return err
	}
	for _, e := range emojis {
		if e.Name == name {
			err := s.GuildEmojiDelete(m.Message.GuildID, e.ID)
			if err != nil {
				fmt.Println("Error deleting emoji:", err)
				s.ChannelMessageSend(
					m.ChannelID,
					fmt.Sprintf("Error deleting emoji %s: %s", name, err.Error()),
				)
				return err
			}
			break
		}
	}

	return nil
}

func CreateEmoji(
	imgFile *os.File,
	emoteName string,
	s *discordgo.Session,
	m *discordgo.MessageCreate,
) error {
	data, err := io.ReadAll(imgFile)
	if err != nil {
		s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("Error reading image: %s", err.Error()),
		)
		return err
	}

	name := "GoBot"
	if emoteName != "" {
		name = emoteName
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	emojiParams := discordgo.EmojiParams{
		Name:  name,
		Image: "data:image/png;base64," + encoded,
	}
	emoji, err := s.GuildEmojiCreate(m.Message.GuildID, &emojiParams)
	if err != nil {
		s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("Error creating emoji: %s", err.Error()),
		)
		return err
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<:%s>", emoji.APIName()))

	return nil
}
