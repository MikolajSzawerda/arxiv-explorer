package main

import (
	"flag"
	"fmt"
	"log"

	"arxiv_explorer/pkg/arxiv"
	"arxiv_explorer/pkg/config"
	"arxiv_explorer/pkg/db"
	"arxiv_explorer/pkg/markdown"
)

func main() {
	// Load configuration (including OpenAI API key from .env)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Parse command-line flags for file paths
	queriesFile := flag.String("queries", "queries.json", "Path to JSON queries file")
	dbPath := flag.String("db", "./mydb.sqlite", "Path to SQLite database file")
	flag.Parse()

	// Connect to the SQLite database (creates file if needed)
	database, err := db.Connect(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	err = db.CreateTablesIfNotExist(database)
	if err != nil {
		log.Fatal(err)
	}

	// Read queries from JSON file
	queries, err := arxiv.ReadQueriesFromJSON(*queriesFile)
	if err != nil {
		log.Fatalf("Failed to read queries: %v", err)
	}

	// Process each query (includes ArXiv API call, GPT enrichment, and DB storage)
	for _, query := range queries {
		arxiv.ProcessQuery(database, query, cfg.OpenAIKey)
	}

	fmt.Println("All queries processed.")

	// Optionally, retrieve all entries from the database and save to a markdown file.
	entries, err := db.FetchEntries(database)
	if err != nil {
		log.Fatal("Failed to retrieve data:", err)
	}
	err = markdown.SaveEntriesToMarkdown(entries, "autoresearch.md")
	if err != nil {
		log.Fatal("Failed to save entries to markdown:", err)
	}
	fmt.Println("All rows saved to markdown")
}
