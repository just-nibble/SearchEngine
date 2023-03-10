package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"

	"os/exec"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/just-nibble/SearchEngine/pkg/pdf"
)

func Bootstrap(es *elasticsearch.Client, entries []*pdf.Entry) error {
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

		newUUID, err := exec.Command("uuidgen").Output()
		if err != nil {
			return err
		}
		_, err3 := esapi.CreateRequest{
			Index:      idx,
			DocumentID: string(newUUID),
			Body:       bytes.NewReader(payload),
		}.Do(ctx, es)
		if err != nil {
			return err3
		}
	}

	return nil
}
