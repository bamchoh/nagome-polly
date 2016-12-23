package player

import (
	"io"
	"os/exec"
	"log"

	"github.com/aws/aws-sdk-go/service/polly"
)

func Play(resp *polly.SynthesizeSpeechOutput, logger *log.Logger) error {
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
		io.Copy(wr, resp.AudioStream)
		wr.Close()
	}()
	cmd.Wait()
	return nil
}
