package image_processing

import (
	"errors"
	"io"
	"net/http"
	"os"
)

func WriteImageToFile(fileName, imageUrl string) (*os.File, error) {
	// Get the response bytes from the url
	response, err := http.Get(imageUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, errors.New(response.Body.Close().Error())
	}

	// Create an empty file
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	// Write the bytes to the file
	if _, err := io.Copy(file, response.Body); err != nil {
		return nil, err
	}

	// Go back to beginning of copied file
	file.Seek(0, 0)
	return file, nil
}
