package main

import (
	"go-youtube-gif-converter/service"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	db, err := service.InitializeDatabase()
	if err == nil {
		log.Println(err)
	}
	if db == nil {
		log.Println("DB is still nil!")
	}
	r := mux.NewRouter()
	r.HandleFunc("/gif", service.PostGIFHandler).Methods("POST")
	r.HandleFunc("/gif/{videoID}.gif", service.GetGIFHandler).Name("videoID").Methods("GET")
	r.HandleFunc("/gif/status/{videoID}", service.GetGIFStatusHandler).Name("videoID").Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

// func _legacy_main() {
// 	url := os.Args[len(os.Args)-1]

// 	fmt.Println(url)

// 	video, err := service.GetYoutubeStream(url)
// 	if err != nil {
// 		panic(err)
// 	}
// 	file, err := os.Create(os.Args[len(os.Args)-2])
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()

// 	// Create new Writer from file
// 	writer := io.Writer(file)

// 	service.ConvertStream(video.Stream, video.Size, &writer, os.Args[len(os.Args)-3])
// 	if err != nil {
// 		panic(err)
// 	}
// }
