package image_processing

import "os/exec"

func RemoveBackground() (string, error) {
	cmd := exec.Command("python", "./remove_background.py")

	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Return path to file
	return "output.png", nil
}
