package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	fileName string
	pdfUrl   string
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	pdfUrl = os.Getenv("URL")
	fileURL, err := url.Parse(pdfUrl)
	if err != nil {
		log.Fatal(err)
	}
	values := fileURL.Query()

	for i := 0; i <= 10; i++ {
		values.Set("ID", strconv.Itoa(i))
		fileURL.RawQuery = values.Encode()
		fmt.Printf("%s", fileURL.String())
		fmt.Println()

		fileName = strconv.Itoa(i) + ".pdf"

		// Create blank file
		file, err := os.Create("./pdfs/" + fileName)
		if err != nil {
			log.Fatal(err)
		}
		client := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}
		// Put content on file
		resp, err := client.Get(fileURL.String())
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		size, err := io.Copy(file, resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		fmt.Printf("Downloaded a file %s with size %d", fileName, size)
		fmt.Println()
	}
}
