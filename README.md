# ARXIV EXPLORER

Simple console app for exploring arxiv written in go.

### Usecase

Apps allows for parraler searching phrases in arxiv papers, additionaly preparing summary with LLM so that is it easier to choose interesting papers.

App is meant to be run periodicaly, so ids of already scraped papers are stored in .sqlite db and during scraping directory with summaries of each query in markdown is created with its date.

