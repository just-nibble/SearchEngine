package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type Entry struct {
	ID    string
	Title string
	Body  string
	Meta  string
}

func NewEntry(id string, title string, body string, meta string) *Entry {
	entry := &Entry{
		ID:    id,
		Title: title,
		Body:  body,
		Meta:  meta,
	}
	return entry
}

func Bootstrap(es *elasticsearch.Client, entries []*Entry) error {
	idx := "books"
	ctx := context.Background()
	_, err := esapi.IndicesDeleteRequest{Index: []string{idx}}.Do(ctx, es)
	if err != nil {
		return err
	}
	_, err2 := esapi.IndicesCreateRequest{Index: idx}.Do(ctx, es)
	if err2 != nil {
		return err2
	}

	for _, e := range entries {
		payload, err := json.Marshal(e)
		if err != nil {
			return err
		}

		_, err3 := esapi.CreateRequest{
			Index:      idx,
			DocumentID: e.ID,
			Body:       bytes.NewReader(payload),
		}.Do(ctx, es)
		if err != nil {
			return err3
		}
	}

	return nil
}
