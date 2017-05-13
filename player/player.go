// +build !windows

package player

import (
	"io"
	"os/exec"
)

func Play(sound io.ReadCloser) error {
	cmd := exec.Command("play", "-t", "mp3", "-")
	wr, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		logger.Println(err)
	}
	go func() {
		io.Copy(wr, sound)
		wr.Close()
	}()
	cmd.Wait()
	return nil
}
