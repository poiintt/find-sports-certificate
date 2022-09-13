package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/joho/godotenv"
	"github.com/ledongthuc/pdf"
	"github.com/unidoc/unipdf/v3/model"
	"golang.org/x/net/html/charset"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	downloadPdfs(1, 10)
}

func downloadPdfs(start int, end int) {
	pdfUrl := os.Getenv("URL")
	fileURL, err := url.Parse(pdfUrl)
	if err != nil {
		log.Fatal(err)
	}
	urlValues := fileURL.Query()

	pdfFolder := "./pdfs/"

	for i := start; i <= end; i++ {
		start := time.Now()

		fileName := strconv.Itoa(i) + ".pdf"
		file, _ := os.Open(pdfFolder + fileName)

		fmt.Printf("%d: ", i)

		// skip downloaded files
		if file != nil {
			fmt.Printf("Skipping")
		} else {
			urlValues.Set("ID", strconv.Itoa(i))
			fileURL.RawQuery = urlValues.Encode()

			client := http.Client{
				CheckRedirect: func(r *http.Request, via []*http.Request) error {
					r.URL.Opaque = r.URL.Path
					return nil
				},
			}

			resp, err := client.Get(fileURL.String())
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			// create empty file
			file, err := os.Create(pdfFolder + fileName)
			if err != nil {
				log.Fatal(err)
			}

			size, err := io.Copy(file, resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			fileSize := units.HumanSize(float64(size))
			fmt.Printf("Downloaded (%s)", fileSize)
		}

		content, err := getPdfContent(pdfFolder + fileName)
		if err != nil {
			panic(err)
		}

		subject, err := getPdfSubject(file)
		if err != nil {
			panic(err)
		}

		fmt.Println()
		fmt.Println("subject: ", subject)
		fmt.Println("content: ", content)

		duration := time.Since(start)
		fmt.Printf(" in %s", duration)
		fmt.Println()
	}
}

func getPdfContent(path string) (string, error) {
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

func getPdfSubject(file *os.File) (string, error) {
	pdfReader, _ := model.NewPdfReader(file)

	pdfInfo, _ := pdfReader.GetPdfInfo()
	subject := pdfInfo.Subject.String()
	subject = strings.Replace(subject, "Urkunde ", "", 1)
	subject = convertToUTF8(subject, "ascii")

	return subject, nil
}

func convertToUTF8(str string, origEncoding string) string {
	strBytes := []byte(str)
	byteReader := bytes.NewReader(strBytes)
	reader, _ := charset.NewReaderLabel(origEncoding, byteReader)
	strBytes, _ = ioutil.ReadAll(reader)
	return string(strBytes)
}
