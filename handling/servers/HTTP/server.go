package HTTP

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	process "pcap_receiver/handling/process"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

var dirName string
var DB process.Saver

func HandleHTTP(addr, dirN string, DBase process.Saver) {
	DB = DBase
	dirName = dirN
	err := os.MkdirAll(dirName, os.ModePerm)

	if err != nil {
		log.Fatalf("couldn't create path, %v", err)
	}
	fmt.Println("HTTP Server has started")
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", uploadFile)
	http.ListenAndServe(addr, nil)

}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong request method"))
		return
	}

	if err := r.ParseMultipartForm(100 * 1024 * 1024); err != nil {
		fmt.Printf("could not parse multipart form: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("CANT_PARSE_FORM"))
		return
	}

	file, fileHeader, err := r.FormFile("uploadFile")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("INVALID_FILE"))
		return
	}
	defer file.Close()

	fileSize := fileHeader.Size

	fileContent, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("INVALID_FILE"))
		return
	}

	fileType := http.DetectContentType(fileContent)
	if fileType != "application/octet-stream" {
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte("Wrong file type!"))
		if err != nil {
			log.Fatal(err)
			return

		}
		return
	}

	fileName := time.Now().Format("01-02-2022-454545.059")

	newPath := filepath.Join(dirName, fileName)
	fmt.Printf("FileType: %s, File: %s\n", fileType, newPath)
	fmt.Printf("File size (bytes): %v\n", fileSize)

	err = saveFile(fileContent, newPath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("CANT_SAVE_FILE"))
		log.Fatal(err)
		return
	} else {

		results, _ := process.Handle{}.Process(newPath)

		res, _ := json.Marshal(&results)

		w.Write(res)

		DB.SaveToDB(results, newPath)
		return
	}
}

func saveFile(content []byte, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return errors.New("couldn't create file")
	}

	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return errors.New("counldn't write to file")
	}
	return nil
}
