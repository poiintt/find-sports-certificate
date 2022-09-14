package main

import (
	"bytes"
	"database/sql"
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
	_ "github.com/mattn/go-sqlite3"
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

	db, _ := sql.Open("sqlite3", "./sqlite.db")
	defer db.Close()

	for i := start; i <= end; i++ {
		start := time.Now()

		fileName := strconv.Itoa(i) + ".pdf"
		file, _ := os.Open(pdfFolder + fileName)

		fmt.Printf("%d: ", i)

		// skip downloaded files
		if file != nil {
			fmt.Println("Found PDF")
		} else {

			cert := getCertificate(db, i)
			if cert != 0 {
				fmt.Println("Found db entry")
			} else {
				urlValues.Set("ID", strconv.Itoa(i))
				fileURL.RawQuery = urlValues.Encode()

				fmt.Println("Request")
				resp, err := http.Get(fileURL.String())

				if err != nil {
					log.Fatal(err)
				}
				defer resp.Body.Close()

				duration := time.Since(start)
				fmt.Printf("Downloaded in %s", duration)
				fmt.Println()
				contentType := resp.Header["Content-Type"][0]

				sql := `INSERT INTO certificate(id, type) VALUES(?, ?)
  ON CONFLICT(id) DO UPDATE SET type=excluded.type;`
				statement, _ := db.Prepare(sql)
				statement.Exec(i, contentType)

				if contentType == "application/pdf" {
					// create empty file
					file, err = os.Create(pdfFolder + fileName)
					if err != nil {
						log.Fatal(err)
					}

					size, err := io.Copy(file, resp.Body)
					if err != nil {
						log.Fatal(err)
					}
					defer file.Close()

					fileSize := units.HumanSize(float64(size))
					fmt.Printf("(%s)", fileSize)
					fmt.Println()
				} else {
					fmt.Println(contentType)
				}
			}
		}

		content, err := getPdfContent(pdfFolder + fileName)
		if err == nil {

			subject, err := getPdfSubject(file)
			if err != nil {
				panic(err)
			}
			name := strings.Replace(subject, "Urkunde ", "", 1)
			content = strings.Replace(content, name, "", 1)

			insertCertificate(db, i, name, content)
			getCertificate(db, i)
		}
	}
}

func getPdfContent(path string) (string, error) {
	_, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}

	var textBuilder bytes.Buffer
	p := r.Page(1)

	s, _ := p.GetPlainText(nil)
	textBuilder.WriteString(s)
	return textBuilder.String(), nil
}

func getPdfSubject(file *os.File) (string, error) {
	pdfReader, _ := model.NewPdfReader(file)

	pdfInfo, _ := pdfReader.GetPdfInfo()
	subject := pdfInfo.Subject.String()
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

func initDb() {
	fmt.Println("Creating sqlite.db...")
	file, err := os.Create("sqlite.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	fmt.Println("sqlite.db created")

	db, _ := sql.Open("sqlite3", "./sqlite.db")
	defer db.Close()
	createTable(db)
}

func createTable(db *sql.DB) {
	sql := `CREATE TABLE certificate (
		"id" integer NOT NULL PRIMARY KEY,		
		"name" TEXT,
		"content" TEXT
		"type" TEXT
	  );`

	fmt.Println("Create certificate table...")
	statement, err := db.Prepare(sql)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
	fmt.Println("certificate table created")
}

func insertCertificate(db *sql.DB, id int, name string, content string) {
	sql := `INSERT INTO certificate(id, name, content) VALUES(?, ?, ?)
  ON CONFLICT(id) DO UPDATE SET name=excluded.name, content=excluded.content;`

	statement, err := db.Prepare(sql)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(id, name, content)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func getCertificate(db *sql.DB, id int) int {
	sql := "SELECT id FROM certificate WHERE id = (?)"

	var id2 int

	err := db.QueryRow(sql, id).Scan(&id2)
	if err != nil {
		return 0
	}
	return 1
}
