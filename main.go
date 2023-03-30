package main

// #cgo pkg-config: python3
// #include <Python.h>
import "C"

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"unsafe"

	// "github.com/DataDog/go-python3"
	"github.com/bwmarrin/discordgo"
	"github.com/nfnt/resize"
)

const TWO_HUNDRED_FIFTY_SIX_KB_IN_BYTES = 262144

var (
	Token     string
	BotPrefix string

	config *configStruct
)

type configStruct struct {
	Token     string `json:"Token"`
	BotPrefix string `json:"BotPrefix"`
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
	if len(m.Attachments) == 0 {
		return errors.New("Image attachment must be provided.")
	}

	attachment := m.Attachments[0]
	imgUrl := attachment.URL
	if imgUrl == "" {
		return errors.New("The attached image must have an image URL.")
	}

	img, err := DownloadImage("emote", imgUrl)
	// Verify it's okay to defer Close here.
	defer img.Close()
	if err != nil {
		return err
	}

	newImg, err := ResizeImage(img)
	defer newImg.Close()
	if err != nil {
		return err
	}

	fileInfo, err := newImg.Stat()
	if err != nil {
		return err
	}
	if fileInfo.Size() > TWO_HUNDRED_FIFTY_SIX_KB_IN_BYTES {
		// Resize to smaller size
		fmt.Println("PANIC")
	}
	newImg.Seek(0, 0)

	// Remove background
	rembgImg, err := Py_RemoveBackground(newImg)
	defer rembgImg.Close()
	if err != nil {
		fmt.Println("Error removing background:", err)
		return err
	}

	data, err := ioutil.ReadAll(rembgImg)
	if err != nil {
		fmt.Println("Error reading file into []byte:", err)
		return err
	}
	// Encode the bytes to base64
	encoded := base64.StdEncoding.EncodeToString(data)
	emojiParams := discordgo.EmojiParams{
		Name:  "GoBot",
		Image: "data:image/png;base64," + encoded, // valid URI
	}
	emoji, err := s.GuildEmojiCreate(m.Message.GuildID, &emojiParams)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<:%s>", emoji.APIName()))

	return nil
}

// NEED to set PYTHONPATH, PKG_CONFIG_PATH and LD_LIBRARY_PATH for python3 to work
func Py_RemoveBackground(img *os.File) (*os.File, error) {
	// defer python3.Py_Finalize()
	// python3.Py_Initialize()

	// dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	// //...
	// _ = python3.PyRun_SimpleString("import sys\nsys.path.append(\"" + dir + "\")")

	// Issue: rembg reqs numpy 1.23.5 which doesn't support python 3.7
	// Currently, Go thinks I'm using python 3.7
	// python3.PyRun_SimpleString("from rembg import remove; print('hello world')")

	pycodeGo := "from rembg import remove; print('hello world')"

	defer C.Py_Finalize()
	C.Py_Initialize()
	pycodeC := C.CString(pycodeGo)
	defer C.free(unsafe.Pointer(pycodeC))
	C.PyRun_SimpleString(pycodeC)

	// sys_path := python3.PySys_GetObject("path")
	// Check what sys_path is here to see if maybe it's
	// some other dir that doesn't contain rembg

	// rembg := python3.PyImport_ImportModule("rembg")
	// if rembg == nil {
	// 	fmt.Println("Failed to import rembg module")
	// 	return nil, nil
	// }
	// defer rembg.DecRef()
	//  _ = python3.PyImport_AddModule("rembg")

	// _ = python3.PyModule_GetDict(rembg)

	// removeFn := python3.PyDict_GetItemString(oDict, "remove")

	// // removeFn := rembg.GetAttrString("remove")
	// defer removeFn.DecRef()

	// _, err := python3.PyRun_AnyFile("pyrembg/remove_background.py")
	// if err != nil {
	// 	return nil, fmt.Errorf("Error removing background execution: %v", err)
	// }
	// _ = python3.PyTuple_New(1)
	// imgPath := python3.PyUnicode_FromString("resized-emote")
	// python3.PyTuple_SetItem(args, 0, imgPath)
	// _ = removeFn.Call(args, python3.Py_None)

	fmt.Println("CALL SUCCESSFUL!")

	// Convert the `result` PyObject to a []uint8 slice
	// pyBytes := result.Bytes()
	// cBytes := python3.PyBytes_AsString(pyBytes)
	// imgBytes := make([]byte, python3.PyBytes_Size(pyBytes))
	// copy(imgBytes, cBytes)

	// // Create a new bytes.Buffer and write the image bytes to it
	// buffer := bytes.NewBuffer(imgBytes)

	// // Use the image.Decode function to decode the image from the buffer
	// rembgImg, _, err := image.Decode(buffer)
	// if err != nil {
	// 	return nil, err
	// }

	// out, err := os.Create("rembg-emote")
	// if err != nil {
	// 	return nil, err
	// }
	// // defer out.Close()
	// out.Seek(0, 0)

	// // Encode the resized image as PNG
	// err = png.Encode(out, rembgImg)
	// if err != nil {
	// 	fmt.Println("Error encoding!")
	// 	return nil, err
	// }

	// rembgImg, err := os.Open("rembg-emote")
	// if err != nil {
	// 	fmt.Println("Error opening file:", err)
	// 	return nil, fmt.Errorf("Error opening file:", err)
	// }
	// // defer rembgImg.Close()

	return nil, nil
}

// Does not defer closing the output image file
func ResizeImage(img *os.File) (*os.File, error) {
	decodedImg, _, err := image.Decode(img)
	if err != nil {
		return nil, fmt.Errorf("Error decoding image: %v", err)
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

func GetImageFormat(img *os.File) string {
	// Read the first 8 bytes of the file
	var header [8]byte
	_, err := io.ReadFull(img, header[:])
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Do this before or after it's called
	// Shouldn't be this method's responsibility
	// img.Seek(0, 0)

	// Check the format signature
	if bytes.Equal(header[:], []byte("\x89PNG\r\n\x1A\n")) {
		return "PNG"
	} else if bytes.Equal(header[:], []byte("\xff\xd8\xff")) {
		return "JPEG"
	} else {
		return ""
	}
}

// Does not defer closing file
func DownloadImage(fileName, imageUrl string) (*os.File, error) {
	// Get the response bytes from the url
	response, err := http.Get(imageUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New("Received non 200 response code")
	}
	// Create an empty file
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	// Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return nil, err
	}

	// Go back to beginning of copied file
	file.Seek(0, 0)

	return file, nil
}
