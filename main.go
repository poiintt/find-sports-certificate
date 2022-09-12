package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/docker/go-units"
	"github.com/joho/godotenv"
	"github.com/ledongthuc/pdf"
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

	for i := 1; i <= 10; i++ {
		start := time.Now()

		values.Set("ID", strconv.Itoa(i))
		fileURL.RawQuery = values.Encode()
		fmt.Printf("%s", fileURL.String())
		fmt.Println()

		fileName = strconv.Itoa(i) + ".pdf"
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

		// Create empty file
		file, err := os.Create("./pdfs/" + fileName)
		if err != nil {
			log.Fatal(err)
		}
		size, err := io.Copy(file, resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		content, err := readPdf("./pdfs/" + fileName)
		if err != nil {
			panic(err)
		}
		fmt.Println(content)

		humanSize := units.HumanSize(float64(size))

		duration := time.Since(start)
		fmt.Printf("Downloaded '%s' with size %s in %s", fileName, humanSize, duration)
		fmt.Println()
	}
}

func readPdf(path string) (string, error) {
	_, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	totalPage := r.NumPage()

	var textBuilder bytes.Buffer
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		s, _ := p.GetPlainText(nil)
		textBuilder.WriteString(s)
	}
	return textBuilder.String(), nil
}
