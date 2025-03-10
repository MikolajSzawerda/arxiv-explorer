package arxiv

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"arxiv_explorer/pkg/db"
	"arxiv_explorer/pkg/gpt"
	"arxiv_explorer/pkg/markdown"
	"arxiv_explorer/pkg/models"
)

// ConstructArxivAPIURL builds the API URL using the query string.
func ConstructArxivAPIURL(query string) string {
	baseURL := "http://export.arxiv.org/api/query"
	queryParam := "?search_query=" + query
	return baseURL + queryParam
}

// CallArxivAPI calls the ArXiv API and returns the response body.
func CallArxivAPI(queryURL string) ([]byte, error) {
	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ParseArxivXMLResponse parses the XML response from ArXiv.
func ParseArxivXMLResponse(data []byte) (*models.Feed, error) {
	var feed models.Feed
	err := xml.Unmarshal(data, &feed)
	if err != nil {
		return nil, err
	}
	return &feed, nil
}

// ReadQueriesFromJSON reads the queries file and unmarshals it.
func ReadQueriesFromJSON(filename string) ([]models.Query, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var queries models.Queries
	err = json.Unmarshal(bytes, &queries)
	if err != nil {
		return nil, err
	}
	return queries.Queries, nil
}

// ProcessQuery processes a single query by calling ArXiv, enriching with GPT data, and storing results.
func ProcessQuery(dbConn *sql.DB, query models.Query, openAIKey string) {
	fmt.Printf("Processing query ID: %s with query: %s\n", query.ID, query.Query)
	url := ConstructArxivAPIURL(query.Query)
	data, err := CallArxivAPI(url)
	if err != nil {
		log.Fatalf("Failed to call ArXiv API: %v", err)
	}
	feed, err := ParseArxivXMLResponse(data)
	if err != nil {
		log.Fatalf("Failed to parse XML response: %v", err)
	}

	var paperIDs []string
	for _, entry := range feed.Entries {
		paperIDs = append(paperIDs, entry.ID)
	}
	if len(paperIDs) == 0 {
		log.Printf("No entries for query: %s", query.ID)
		return
	}

	existingPapers, err := db.GetExistingPaperIDs(dbConn, paperIDs)
	if err != nil {
		log.Fatalf("Failed to check existing papers: %v", err)
	}

	var newEntries []models.Entry
	enrichedEntries := make(chan models.Entry)
	var wg sync.WaitGroup

	// Process each paper concurrently.
	for _, entry := range feed.Entries {
		wg.Add(1)
		go func(entry models.Entry) {
			defer wg.Done()
			if _, exists := existingPapers[entry.ID]; exists {
				return
			}
			log.Printf("Calling GPT for %s", entry.ID)
			gptResponse, err := gpt.GetGPTInfoForPaper(openAIKey, entry.Summary)
			if err != nil {
				log.Printf("Failed GPT call for paper %s: %v", entry.ID, err)
				return
			}
			entry.GPTSummary = gptResponse.Summary
			entry.GPTNewContributions = gptResponse.NewContributions
			entry.GPTTags = gptResponse.Tags
			enrichedEntries <- entry
		}(entry)
	}
	go func() {
		wg.Wait()
		close(enrichedEntries)
	}()
	for entry := range enrichedEntries {
		newEntries = append(newEntries, entry)
	}

	err = db.SaveToDatabaseBatch(dbConn, newEntries, query.ID)
	if err != nil {
		log.Printf("Failed to save entries to DB for query ID %s: %v", query.ID, err)
	} else {
		fmt.Printf("Saved %d new papers for query ID %s to the database.\n", len(newEntries), query.ID)
	}
	if len(newEntries) == 0 {
		return
	}
	fileName := fmt.Sprintf("%3d_%s.md", time.Now().Nanosecond()/1e6, query.ID)
	err = markdown.SaveEntriesToMarkdown(newEntries, fmt.Sprintf("summary/%s/%s", time.Now().Format("02-01"), fileName))
	if err != nil {
		log.Print("Failed to save entries to md file")
	}
}
