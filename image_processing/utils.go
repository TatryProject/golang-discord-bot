package image_processing

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	"github.com/nfnt/resize"
)

type ImageFormat string

const (
	PNG ImageFormat = "PNG"
	JPG ImageFormat = "JPEG"
)

func ResizeImageForDiscord(path string, width uint, height uint) (*os.File, error) {
	var err error

	imgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	imgFormat, err := GetImageFormat(imgFile)
	if err != nil {
		return nil, err
	}

	var img image.Image
	if imgFormat == PNG {
		img, err = png.Decode(imgFile)
	} else if imgFormat == JPG {
		img, err = jpeg.Decode(imgFile)
	}
	if err != nil {
		return nil, err
	}

	// Add logic to determine if width or height is the min dimension
	// Set the min dimension's arg to 0 in the call to Resize
	resizedImg := resize.Resize(width, height, img, resize.Lanczos3)

	// Create a new file to store the resized image
	out, err := os.Create(fmt.Sprintf("resized-%s", imgFile.Name()))
	if err != nil {
		return nil, err
	}
	out.Seek(0, 0)

	// Encode the resized image as PNG
	err = png.Encode(out, resizedImg)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func GetImageFormat(img *os.File) (ImageFormat, error) {
	// Read the first 8 bytes of the file
	var header [8]byte
	_, err := io.ReadFull(img, header[:])
	if err != nil {
		return "", err
	}

	img.Seek(0, 0)

	// Check the format signature
	if bytes.Equal(header[:], []byte("\x89PNG\r\n\x1A\n")) {
		return PNG, nil
	} else if bytes.Equal(header[:], []byte("\xff\xd8\xff")) {
		return JPG, nil
	} else {
		return "", errors.New("could not determine image format of file")
	}
}
