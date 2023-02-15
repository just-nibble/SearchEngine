package pdf

import (
	"log"
	"os"
	"path/filepath"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

type Entry struct {
	ID    string
	Title string
	Body  string
	Meta  string
}

func NewEntry(title string, body string) *Entry {
	entry := &Entry{
		Title: title,
		Body:  body,
	}
	return entry
}

func ExtractText(dir string) ([]*Entry, error) {
	output := []*Entry{}
	// Reading all files of directory
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	// Iterating through all the files
	for _, f := range files {
		// checking for file extensions
		ext := filepath.Ext(f.Name())
		// if extension relevant
		if ext == ".pdf" {
			// then open file
			f, err := os.Open(f.Name())
			if err != nil {
				panic(err)
			}
			defer f.Close()

			reader, err := model.NewPdfReaderLazy(f)
			if err != nil {
				panic(err)
			}
			// Read page
			p, err := reader.GetPage(1)
			if err != nil {
				panic(err)
			}
			// create extractor for page
			ex, err := extractor.New(p)
			if err != nil {
				panic(err)
			}
			// read text using extractor
			text, err := ex.ExtractText()
			output = append(
				output, NewEntry(f.Name(), text),
			)
			if err != nil {
				return nil, err
			}
		}
	}
	return output, nil
}
