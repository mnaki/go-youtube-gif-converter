package main

import (
	"fmt"
	"io"
	"os"

	youtube "github.com/kkdai/youtube/v2"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
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

func ConvertStream(r *io.ReadCloser, _ int64, w *os.File) error {

	fluentffmpeg.
		NewCommand("").
		PipeInput(*r).
		OutputFormat("gif").
		PipeOutput(w).
		Run()

	return nil
}

func main() {
	url := os.Args[len(os.Args)-1]

	fmt.Println(url)

	youtubeReader, size, err := GetYoutubeStreamFromURL(url)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(os.Args[len(os.Args)-2])
	if err != nil {
		panic(err)
	}
	ConvertStream(youtubeReader, size, file)
	if err != nil {
		panic(err)
	}
}
