package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	ESCustom "github.com/just-nibble/SearchEngine/pkg/elasticsearch"
	"github.com/just-nibble/SearchEngine/pkg/pdf"
)

var (
	listenAddr  string
	addr        string
	esAddresses string
)

func main() {
	flag.StringVar(&listenAddr, "listen-addr", ":5000", "server listen address")
	flag.StringVar(&esAddresses, "es-addresses", "http://es01:9200", "elasticsearch addresses")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	es := newEsClient(logger, strings.Split(esAddresses, ","))
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)

	server := newWebserver(logger, es)
	go gracefullShutdown(server, logger, quit, done)

	go indexDirectory("./bin/static", es)

	logger.Println("Server is ready to handle requests at", listenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")
}

func gracefullShutdown(server *http.Server, logger *log.Logger, quit <-chan os.Signal, done chan<- bool) {
	<-quit
	logger.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
	close(done)
}

func newEsClient(logger *log.Logger, addresses []string) *elasticsearch.Client {
	cfg := elasticsearch.Config{Addresses: addresses}
	client, err := elasticsearch.NewClient(cfg)

	if err != nil {
		logger.Println(err)
		panic(err)
	}

	return client
}

const searchMatch = `
	"query": {
		"multi_match": {
      		"query": %q,
      		"fields": ["title^100", "body^10",],
      		"operator": "and"
    	}
  	},
  	"highlight": {
    	"fields": {
      		"body": { "number_of_fragments": 0 },
      		"title": { "number_of_fragments": 0 }
    	}
  	},
  	"size": 25,
  	"sort": [{ "_score": "desc" }, { "_doc": "asc" }]`

func buildQuery(query string) io.Reader {
	var b strings.Builder

	b.WriteString("{\n")
	b.WriteString(fmt.Sprintf(searchMatch, query))
	b.WriteString("\n}")

	return strings.NewReader(b.String())
}

func newWebserver(logger *log.Logger, es *elasticsearch.Client) *http.Server {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

		read, write := io.Pipe()

		go func() {
			defer write.Close()
			esInfo, err := es.Info()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				defer esInfo.Body.Close()
				io.Copy(write, esInfo.Body)
			}
		}()
		io.Copy(w, read)
		w.WriteHeader(http.StatusOK)
	})

	router.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		logger.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
		fmt.Println("query: ............................")
		q := r.URL.Query().Get("q")
		fmt.Println("query: +++++++++++++++++++++++++++++")
		fmt.Println("query: ", q)
		read, write := io.Pipe()

		go func() {
			defer write.Close()
			res, err := es.Search(
				es.Search.WithContext(r.Context()),
				es.Search.WithIndex("books"),
				es.Search.WithBody(buildQuery(q)),
				es.Search.WithTrackTotalHits(true),
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				defer res.Body.Close()
				logger.Panicln("Got here...")
				io.Copy(write, res.Body)
			}
		}()
		io.Copy(w, read)
	})

	return &http.Server{
		Addr:         listenAddr,
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
}

func indexDirectory(dir string, es *elasticsearch.Client) {
	// Read PDF
	pdf_texts, err := pdf.ExtractText(dir)

	if err != nil {
		panic(err)
	}
	// Send content to elasticsearch for indexing
	err = ESCustom.Bootstrap(es, pdf_texts)
	if err != nil {
		panic(err)
	}
}
