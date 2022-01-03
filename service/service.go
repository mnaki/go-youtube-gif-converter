package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"net/http"

	"github.com/google/uuid"

	youtube "github.com/kkdai/youtube/v2"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var gDatabase *gorm.DB = nil

func InitializeDatabase() (*gorm.DB, error) {
	var err error
	_db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Printf("Error opening database: %s\n", err)
		return nil, err
	}
	err = _db.AutoMigrate(&GIFMetaData{})
	if err != nil {
		log.Printf("Error auto-migrating database: %s\n", err)
		return nil, err
	}
	gDatabase = _db
	return _db, nil
}

// Contains metadata about a GIF's video conversion status and metadata
type GIFMetaData struct {
	gorm.Model
	ID               uint   `gorm:"primary_key"`
	VideoID          string `json:"video_id"`
	GIFFileName      string `json:"gif_file_name"`
	VideoFileName    string `json:"video_file_name"`
	ConversionUUID   string `json:"conversion_uuid"`
	ConversionStatus string `json:"conversation_status"`
}

func getYoutubeStream(youtubeVideoID string) (*Video, error) {
	youtubeClient := youtube.Client{}

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
	Stream *io.ReadCloser
	Size   int64
	Format *youtube.Format
	Client *youtube.Client
}

func convertStream(gifFile *io.Writer, videoFile *io.ReadCloser) error {

	cmd := fluentffmpeg.NewCommand("")
	cmd.PipeInput(*videoFile)
	cmd.OutputFormat("gif")
	cmd.PipeOutput(*gifFile)
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

// Convert the video to a GIF using FFMPEG and updates gifMetaData status
func convertVideoToGIF(db *gorm.DB, gifMetaData *GIFMetaData) error {
	var err error = nil

	log.Printf("UUID=%s | Starting conversion\n", gifMetaData.ConversionUUID)
	gifMetaData.ConversionStatus = "converting"
	db.Save(gifMetaData)
	// Get stream from videoID
	log.Printf("UUID=%s | getting video stream\n", gifMetaData.ConversionUUID)
	video, err := getYoutubeStream(gifMetaData.VideoID)
	if err != nil {
		log.Printf("UUID=%s | getting video stream failed: %s\n", gifMetaData.ConversionUUID, err)
		return err
	}
	// Create file to write GIF to
	log.Printf("UUID=%s | creating gif file", gifMetaData.ConversionUUID)
	gifFile, err := os.Create(gifMetaData.GIFFileName)
	if err != nil {
		log.Printf("UUID=%s | creating gif file failed: %s\n", gifMetaData.ConversionUUID, err)
		return err
	}
	defer gifFile.Close()
	gifFileReader := io.Writer(gifFile)
	// Convert video stream to GIF
	log.Printf("UUID=%s | converting video stream to gif\n", gifMetaData.ConversionUUID)
	err = convertStream(&gifFileReader, video.Stream)
	if err != nil {
		log.Printf("UUID=%s | converting video stream to gif failed: %s\n", gifMetaData.ConversionUUID, err)
		return err
	}
	// Update gifMetaData status
	gifMetaData.ConversionStatus = "done"
	log.Printf("UUID=%s | Updating DB entry", gifMetaData.ConversionUUID)
	db.Save(gifMetaData)

	return nil
}

func PostGIFHandler(response http.ResponseWriter, request *http.Request) {
	conversionUUID := uuid.New().String()

	// Unmarshall request body
	var body map[string]interface{}
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		log.Printf("UUID=%s | Error decoding request body: %s\n", conversionUUID, err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}

	// Take videoID from request body
	videoID := body["videoID"].(string)

	// Create GIFMetaData and populate it
	var gifMetaData = GIFMetaData{
		ConversionUUID:   conversionUUID,
		ConversionStatus: "pending",
		VideoID:          videoID,
		GIFFileName:      fmt.Sprintf("%s_%s.gif", conversionUUID, videoID),
		VideoFileName:    fmt.Sprintf("%s_%s.mp4", conversionUUID, videoID),
	}
	err = gDatabase.Create(&gifMetaData).Error
	if err != nil {
		fmt.Printf("UUID=%s | Error creating gifMetaData: %s\n", conversionUUID, err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	go convertVideoToGIF(gDatabase, &gifMetaData)
	// Return conversionUUID in a JSON response
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]string{
		"conversionUUID": conversionUUID,
	})
}

// Get GIF from the database, using the videoID
func GetGIFHandler(response http.ResponseWriter, request *http.Request) {
	// Get videoID from request URL
	videoID := request.URL.Query().Get("videoID")
	// Get first GIFMetaData with videoID
	var gifMetaData GIFMetaData
	gDatabase.Where(GIFMetaData{VideoID: videoID, ConversionStatus: "done"}).First(&gifMetaData)
	// Open GIF file
	gifFile, err := os.Open(gifMetaData.GIFFileName)
	// If GIFMetaData found, return GIF
	if err != nil {
		log.Printf("UUID=%s | Error opening GIF file: %s\n", gifMetaData.ConversionUUID, err)
		response.WriteHeader(http.StatusInternalServerError)
	}
	defer gifFile.Close()
	// Write GIF to response
	_, err = io.Copy(response, gifFile)
	if err != nil {
		log.Printf("UUID=%s | Error writing GIF (%v) to response: %s\n", gifMetaData.ConversionUUID, gifFile, err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Fetch status from GIFMetaData and return JSON response with status and videoID of GIF
func GetGIFStatusHandler(response http.ResponseWriter, request *http.Request) {
	// Get videoID from request URL
	videoID := request.
		URL.
		Query().
		Get("videoID")
	// Get all GIFMetaData with videoID

	var gifMetaDatas []*GIFMetaData
	gDatabase.
		Where(GIFMetaData{VideoID: videoID}).
		Find(&gifMetaDatas)

	var dataMap []map[string]string
	for _, data := range gifMetaDatas {
		// Print debug
		log.Printf("%+v\n", data)
		dataMap = append(dataMap, map[string]string{
			"videoID":        data.VideoID,
			"conversionUUID": data.ConversionUUID,
			"status":         data.ConversionStatus,
		})
	}

	// Return JSON response with status and videoID
	response.
		Header().
		Set("Content-Type", "application/json")
	response.
		WriteHeader(http.StatusOK)
	// Write JSON response to response
	json.NewEncoder(response).Encode(dataMap)
}
