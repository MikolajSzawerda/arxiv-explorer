package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"arxiv_explorer/pkg/models"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Connect opens (or creates) a SQLite database.
func Connect(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite3", dbPath)
}

// CreateTablesIfNotExist creates the necessary tables if they do not exist.
func CreateTablesIfNotExist(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS papers (
		paper_id VARCHAR(255) PRIMARY KEY,
		title TEXT,
		summary TEXT,
		categories TEXT,
		pdf_url TEXT,
		authors TEXT,
		publication_date DATETIME,
		query_id VARCHAR(255),
		gpt_summary TEXT,
		gpt_contributions TEXT,
		gpt_tags TEXT
	);
	`
	_, err := db.Exec(createTableSQL)
	return err
}

// GetExistingPaperIDs returns a map of paper IDs that are already stored.
func GetExistingPaperIDs(db *sql.DB, paperIDs []string) (map[string]bool, error) {
	existingPapers := make(map[string]bool)
	placeholders := strings.Repeat("?,", len(paperIDs))
	placeholders = strings.TrimRight(placeholders, ",")
	query := fmt.Sprintf("SELECT paper_id FROM papers WHERE paper_id IN (%s)", placeholders)
	args := make([]interface{}, len(paperIDs))
	for i, id := range paperIDs {
		args[i] = id
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var paperID string
	for rows.Next() {
		if err := rows.Scan(&paperID); err != nil {
			return nil, err
		}
		existingPapers[paperID] = true
	}
	return existingPapers, nil
}

// SaveToDatabaseBatch inserts multiple entries into the database at once.
func SaveToDatabaseBatch(db *sql.DB, entries []models.Entry, queryID string) error {
	if len(entries) == 0 {
		return nil
	}
	query := `INSERT INTO papers (paper_id, title, summary, categories, pdf_url, authors, publication_date, query_id, gpt_summary, gpt_contributions, gpt_tags)
			  VALUES `
	values := make([]interface{}, 0, len(entries)*11)
	for i, entry := range entries {
		if i > 0 {
			query += ","
		}
		query += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		publicationDate, _ := time.Parse(time.RFC3339, entry.Published)
		categories := strings.Join(entry.Categories, ",")
		authors := extractAuthors(entry.Authors)
		pdfURL := ""
		if len(entry.PDFLink) > 0 {
			pdfURL = entry.PDFLink[0].Href
		}
		gptTags := strings.Join(entry.GPTTags, ", ")
		values = append(values, entry.ID, entry.Title, entry.Summary, categories, pdfURL, authors, publicationDate, queryID, entry.GPTSummary, entry.GPTNewContributions, gptTags)
	}
	_, err := db.Exec(query, values...)
	return err
}

func extractAuthors(authors []models.Author) string {
	var names []string
	for _, a := range authors {
		names = append(names, a.Name)
	}
	return strings.Join(names, ", ")
}

// FetchEntries retrieves all stored entries from the database.
func FetchEntries(db *sql.DB) ([]models.Entry, error) {
	query := `SELECT paper_id, title, summary, authors, publication_date, categories, pdf_url, gpt_summary, gpt_contributions, gpt_tags FROM papers ORDER BY publication_date DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.Entry
	for rows.Next() {
		var entry models.Entry
		var authors, categories, pdfURL, tags string
		err := rows.Scan(&entry.ID, &entry.Title, &entry.Summary, &authors, &entry.Published, &categories, &pdfURL, &entry.GPTSummary, &entry.GPTNewContributions, &tags)
		if err != nil {
			return nil, err
		}
		entry.Authors = parseAuthors(authors)
		entry.Categories = strings.Split(categories, ",")
		entry.PDFLink = []models.Link{{Href: pdfURL, Rel: "alternate"}}
		entry.GPTTags = strings.Split(tags, ",")
		entries = append(entries, entry)
	}
	return entries, nil
}

func parseAuthors(authorsStr string) []models.Author {
	var authors []models.Author
	parts := strings.Split(authorsStr, ",")
	for _, part := range parts {
		authors = append(authors, models.Author{Name: strings.TrimSpace(part)})
	}
	return authors
}
