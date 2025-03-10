package models

type Feed struct {
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	ID                  string   `xml:"id"`
	Title               string   `xml:"title"`
	Summary             string   `xml:"summary"`
	Authors             []Author `xml:"author"`
	Published           string   `xml:"published"`
	Categories          []string `xml:"category"`
	PDFLink             []Link   `xml:"link"`
	GPTSummary          string
	GPTNewContributions string
	GPTTags             []string
}

type Author struct {
	Name string `xml:"name"`
}

type Link struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type Query struct {
	ID    string `json:"id"`
	Query string `json:"query"`
}

type Queries struct {
	Queries []Query `json:"queries"`
}
