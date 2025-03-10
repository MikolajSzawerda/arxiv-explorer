package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"arxiv_explorer/pkg/models"
)

// SaveEntriesToMarkdown converts entries to Markdown and writes them to the specified file.
func SaveEntriesToMarkdown(entries []models.Entry, fileName string) error {
	var sb strings.Builder

	sb.WriteString("# Research Papers Summary\n\n")
	for _, entry := range entries {
		title := strings.ReplaceAll(entry.Title, "\n", "")
		title = strings.ReplaceAll(title, "\t", "")
		sb.WriteString(fmt.Sprintf("## %s\n\n", title))
		sb.WriteString(fmt.Sprintf("- **ID**: %s\n", entry.ID))
		sb.WriteString(fmt.Sprintf("- **Published**: %s\n", entry.Published))
		sb.WriteString("- **Authors**: ")
		var authorNames []string
		for _, author := range entry.Authors {
			authorNames = append(authorNames, author.Name)
		}
		sb.WriteString(strings.Join(authorNames, ", "))
		sb.WriteString("\n")
		sb.WriteString("- **Categories**: " + strings.Join(entry.Categories, ", ") + "\n")
		sb.WriteString(fmt.Sprintf("\n### GPT Summary\n%s\n", entry.GPTSummary))
		sb.WriteString(fmt.Sprintf("\n### New Contributions\n%s\n", entry.GPTNewContributions))
		sb.WriteString("\n### Tags\n" + strings.Join(entry.GPTTags, ", ") + "\n")
		if len(entry.PDFLink) > 0 {
			sb.WriteString("\n### PDF Link\n")
			sb.WriteString(fmt.Sprintf("[Link](%s)\n", entry.PDFLink[0].Href))
		}
		sb.WriteString("\n---\n\n")
	}
	// Ensure the target directory exists.
	err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(sb.String())
	return err
}
