package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/nfnt/resize"
	openai "github.com/sashabaranov/go-openai"
)

const TWO_HUNDRED_FIFTY_SIX_KB_IN_BYTES = 262144

var (
	Token        string
	BotPrefix    string
	OpenAiApiKey string

	config *configStruct
)

type configStruct struct {
	Token        string `json:"Token"`
	BotPrefix    string `json:"BotPrefix"`
	OpenAiApiKey string `json:"OpenAiApiKey"`
}

func main() {
	err := ReadConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	Start()

	return
}

func ReadConfig() error {
	fmt.Println("Reading config file...")
	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	err = json.Unmarshal(file, &config)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	Token = config.Token
	BotPrefix = config.BotPrefix
	OpenAiApiKey = config.OpenAiApiKey

	return nil
}

var BotId string
var goBot *discordgo.Session

func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	BotId = u.ID
	goBot.AddHandler(messageHandler)
	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	/* Constructing a channel of type struct{} (generic? any object/struct can be sent?)
	 * make allocs and inits new channel.
	 * <- is used to perform send/receive operations.
	 * Here, nothing is receiving the value from the channel.
	 * NOTE: <- is bidirectional: can send and receive vals from same channel.
	 */
	fmt.Println("Bot is running !")
	<-make(chan struct{})
	defer goBot.Close()
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotId {
		return
	}

	if m.Content == fmt.Sprintf("%semote", BotPrefix) {
		handleEmojiAdd(s, m)
	}

	if m.Content == BotPrefix+"ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "pong")
	}
}

func handleEmojiAdd(s *discordgo.Session, m *discordgo.MessageCreate) error {
	fmt.Println("Handling emoji add")
	ctx := context.Background()

	if len(m.Attachments) == 0 {
		return errors.New("Image attachment must be provided.")
	}

	attachment := m.Attachments[0]
	imgUrl := attachment.URL
	if imgUrl == "" {
		return errors.New("The attached image must have an image URL.")
	}

	img, err := downloadImage("emote", imgUrl)
	// Verify it's okay to defer Close here.
	defer img.Close()
	if err != nil {
		return err
	}

	// resize not working correctly yet
	newImg, err := resizeImage(img)
	defer newImg.Close()
	if err != nil {
		return err
	}

	// Remove Background
	_, err = dallE2RemoveBackground(ctx, newImg)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// fileInfo, err := newImg.Stat()
	// if err != nil {
	// 	return err
	// }

	// if fileInfo.Size() > TWO_HUNDRED_FIFTY_SIX_KB_IN_BYTES {
	// 	// Resize to smaller size
	// 	fmt.Println("PANIC")
	// }

	data, err := ioutil.ReadAll(newImg)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}

	// Encode the bytes to base64
	encoded := base64.StdEncoding.EncodeToString(data)

	err = sendBase64ImageToDiscordChannel(encoded, s, m)
	if err != nil {
		return err
	}

	newImg.Seek(0, 0)

	return nil
}

// Does not defer closing file
func downloadImage(fileName, imageUrl string) (*os.File, error) {
	//Get the response bytes from the url
	response, err := http.Get(imageUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	//Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return nil, err
	}

	// Go back to beginning of copied file
	file.Seek(0, 0)

	return file, nil
}

// Does not defer closing the output image file
func resizeImage(img *os.File) (*os.File, error) {
	imgFormat := getImageFormat(img)
	if imgFormat == "" {
		return nil, fmt.Errorf("Could not determine image format of file %s.", img.Name())
	}

	var decodedImg image.Image
	var err error
	if imgFormat == "PNG" {
		decodedImg, err = png.Decode(img)
		if err != nil {
			fmt.Println("We have errored in decoding the png.")
			fmt.Println(err.Error())
			return nil, err
		}
	} else if imgFormat == "JPEG" {
		decodedImg, err = jpeg.Decode(img)
		if err != nil {
			return nil, err
		}
	}

	resizedImg := resize.Resize(128, 128, decodedImg, resize.Lanczos3)

	// Create a new file to store the resized image
	out, err := os.Create(fmt.Sprintf("resized-%s", img.Name()))
	if err != nil {
		return nil, err
	}
	// defer out.Close()
	out.Seek(0, 0)

	// Encode the resized image as PNG
	err = png.Encode(out, resizedImg)
	if err != nil {
		fmt.Println("Error encoding!")
		return nil, err
	}

	return out, nil
}

func getImageFormat(img *os.File) string {
	// Read the first 8 bytes of the file
	var header [8]byte
	_, err := io.ReadFull(img, header[:])
	if err != nil {
		fmt.Println(err)
		return ""
	}

	img.Seek(0, 0)

	// Check the format signature
	if bytes.Equal(header[:], []byte("\x89PNG\r\n\x1A\n")) {
		return "PNG"
	} else if bytes.Equal(header[:], []byte("\xff\xd8\xff")) {
		return "JPEG"
	} else {
		return ""
	}
}

func dallE2RemoveBackground(ctx context.Context, img *os.File) ([]byte, error) {
	client := openai.NewClient(OpenAiApiKey)

	request := openai.ImageEditRequest{
		Image:  img,
		N:      1,
		Prompt: "Remove the image's background such that it is transparent.",
		Size:   "256x256",
	}

	response, err := client.CreateEditImage(ctx, request)
	if err != nil {
		return []byte{}, err
	}

	// fmt.Println(len(response.Data))
	// url := response.Data[0].URL
	// httpClient := &http.Client{
	// 	Timeout: time.Second * 10,
	// }

	// // Make a request using the client
	// req, _ := http.NewRequest("GET", url, nil)
	// resp, _ := httpClient.Do(req)

	return []byte{}, nil
}

func sendBase64ImageToDiscordChannel(
	encodedImg string,
	s *discordgo.Session,
	m *discordgo.MessageCreate,
) error {
	emojiParams := discordgo.EmojiParams{
		Name:  "GoBot",
		Image: "data:image/png;base64," + encodedImg, // valid URI
	}
	emoji, err := s.GuildEmojiCreate(m.Message.GuildID, &emojiParams)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<:%s>", emoji.APIName()))

	return nil
}
