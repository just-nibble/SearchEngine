package pdf

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
			var full_name string = dir + "/" + f.Name()
			outp := "bin/static/store/" + f.Name() + ".txt"
			_, err := exec.Command("pdftotext", full_name, outp).Output()
			if err != nil {
				panic(err)
			}

			out, err := os.ReadFile(outp)
			if err != nil {
				panic(err)
			}
			output = append(output, NewEntry(f.Name(), string(out)))
		}
	}
	return output, nil
}
