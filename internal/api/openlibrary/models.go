package openlibrary

import "time"

// SearchResponse represents the OpenLibrary search API response
type SearchResponse struct {
	NumFound int        `json:"numFound"`
	Start    int        `json:"start"`
	Docs     []BookDoc  `json:"docs"`
}

// BookDoc represents a book document from search results
type BookDoc struct {
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	AuthorKey        []string `json:"author_key"`
	FirstPublishYear int      `json:"first_publish_year"`
	ISBN             []string `json:"isbn"`
	Publisher        []string `json:"publisher"`
	Language         []string `json:"language"`
	CoverI           int      `json:"cover_i"`
	EditionCount     int      `json:"edition_count"`
}

// BookDetails represents detailed book information
type BookDetails struct {
	Key              string        `json:"key"`
	Title            string        `json:"title"`
	Subtitle         string        `json:"subtitle"`
	Description      interface{}   `json:"description"` // Can be string or object
	Authors          []AuthorRef   `json:"authors"`
	Publishers       []string      `json:"publishers"`
	PublishDate      string        `json:"publish_date"`
	ISBN10           []string      `json:"isbn_10"`
	ISBN13           []string      `json:"isbn_13"`
	NumberOfPages    int           `json:"number_of_pages"`
	Subjects         []string      `json:"subjects"`
	Covers           []int         `json:"covers"`
	Works            []WorkRef     `json:"works"`
}

// AuthorRef represents an author reference
type AuthorRef struct {
	Key string `json:"key"`
}

// WorkRef represents a work reference
type WorkRef struct {
	Key string `json:"key"`
}

// AuthorDetails represents detailed author information
type AuthorDetails struct {
	Key           string `json:"key"`
	Name          string `json:"name"`
	BirthDate     string `json:"birth_date"`
	Bio           interface{} `json:"bio"` // Can be string or object
	PersonalName  string `json:"personal_name"`
	AlternateNames []string `json:"alternate_names"`
	Photos        []int  `json:"photos"`
}

// WorkDetails represents detailed work information
type WorkDetails struct {
	Key          string      `json:"key"`
	Title        string      `json:"title"`
	Description  interface{} `json:"description"` // Can be string or object
	Authors      []AuthorRef `json:"authors"`
	Subjects     []string    `json:"subjects"`
	Covers       []int       `json:"covers"`
	FirstPublishDate string  `json:"first_publish_date"`
}

// ISBNResponse represents the OpenLibrary ISBN API response
type ISBNResponse struct {
	Key              string      `json:"key"`
	Title            string      `json:"title"`
	Subtitle         string      `json:"subtitle"`
	Authors          []AuthorRef `json:"authors"`
	Publishers       []string    `json:"publishers"`
	PublishDate      string      `json:"publish_date"`
	ISBN10           []string    `json:"isbn_10"`
	ISBN13           []string    `json:"isbn_13"`
	NumberOfPages    int         `json:"number_of_pages"`
	Subjects         []string    `json:"subjects"`
	Covers           []int       `json:"covers"`
	Works            []WorkRef   `json:"works"`
}

// CachedResponse represents a cached API response
type CachedResponse struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       int         `json:"ttl"` // seconds
}

// ErrorResponse represents an OpenLibrary API error
type ErrorResponse struct {
	Error string `json:"error"`
}
