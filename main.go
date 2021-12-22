package main

import (
	"fmt"
	"io"
	"os"

	youtube "github.com/kkdai/youtube/v2"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
)

func GetYoutubeStreamFromURL(url string) (*Video, error) {
	youtubeClient := youtube.Client{}

	youtubeVideoID, err := youtube.ExtractVideoID(url)
	if err != nil {
		return nil, err
	}
	youtubeVideo, err := youtubeClient.GetVideo(youtubeVideoID)
	if err != nil {
		return nil, err
	}
	videoFormat := youtubeVideo.Formats.FindByItag(133) // 133 = mp4 240p
	youtubeReader, formatSize, err := youtubeClient.GetStream(youtubeVideo, videoFormat)
	if err != nil {
		return nil, err
	}

	return &Video{
		&youtubeReader,
		formatSize,
		videoFormat,
		&youtubeClient,
	}, nil
}

type Video struct {
	stream *io.ReadCloser
	size   int64
	format *youtube.Format
	client *youtube.Client
}

func ConvertStream(r *io.ReadCloser, _ int64, w *io.Writer) error {

	cmd := fluentffmpeg.NewCommand("")
	cmd.PipeInput(*r)
	cmd.OutputFormat("gif")
	cmd.PipeOutput(*w)
	cmd.FrameRate(12)
	cmd.InputOptions(
		"-t", "5",
		"-ss", "30",
	)
	cmd.OutputOptions(
		"-vf", "fps=10,scale=320:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse",
		"-loop", "0",
	)
	cmd.Run()

	return nil
}

func main() {
	url := os.Args[len(os.Args)-1]

	fmt.Println(url)

	video, err := GetYoutubeStreamFromURL(url)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(os.Args[len(os.Args)-2])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create new Writer from file
	writer := io.Writer(file)

	ConvertStream(video.stream, video.size, &writer)
	if err != nil {
		panic(err)
	}
}
