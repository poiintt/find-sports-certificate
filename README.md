# find-sports-certificate

Download PDFs from the `URL` specified in the `.env` by increasing the query param `id` number.

Downloaded PDFs are saved in `./pdfs/`

The text `content` and the `name` from PdfDocumentInformation.Subject are saved in `sqlite.db`.

Already downloaded PDFs will be skipped.
Filsize and download time are logged.

This was created for one specific website, which I won't mention.

```sh
cp dist.env .env
# add url to .env

go run main.go
```