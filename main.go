package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	youtube "github.com/kkdai/youtube/v2"
)

func GetYoutubeStreamFromURL(url string) (*io.ReadCloser, int64, error) {
	youtubeClient := youtube.Client{}

	youtubeVideoID, err := youtube.ExtractVideoID(url)
	if err != nil {
		return nil, 0, err
	}
	youtubeVideo, err := youtubeClient.GetVideo(youtubeVideoID)
	if err != nil {
		return nil, 0, err
	}
	videoFormats := youtubeVideo.Formats.FindByItag(133) // 133 = mp4 240p
	youtubeReader, formatSize, err := youtubeClient.GetStream(youtubeVideo, videoFormats)
	if err != nil {
		return nil, 0, err
	}

	return &youtubeReader, formatSize, nil
}

func ConvertStream(youtubeReader *io.ReadCloser, formatSize int64, w *os.File) error {

	ffmpegCmd := exec.Command(
		"ffmpeg",
		"-ss",
		"30",
		"-t",
		"3",
		"-i",
		"pipe:0",
		"-loglevel",
		"0",
		"-vf",
		"fps=10,scale=320:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse",
		"-loop",
		"0",
		"-f",
		"gif",
		"-",
	)

	ffmpegCmd.Stderr = os.Stderr
	ffmpegCmd.Stdout = w
	pipe, err := ffmpegCmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.CopyN(pipe, *youtubeReader, formatSize)
		if err != nil {
			fmt.Println(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := ffmpegCmd.Run()
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()
	return nil
}

func main() {
	url := os.Args[1]

	fmt.Println(url)

	youtubeReader, size, err := GetYoutubeStreamFromURL(url)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	ConvertStream(youtubeReader, size, file)
	if err != nil {
		panic(err)
	}
}
