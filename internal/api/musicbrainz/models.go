package musicbrainz

import "time"

// SearchReleaseResponse represents the MusicBrainz release search API response
type SearchReleaseResponse struct {
	Count    int       `json:"count"`
	Offset   int       `json:"offset"`
	Releases []Release `json:"releases"`
}

// Release represents a MusicBrainz release (album)
type Release struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Status         string         `json:"status"`
	Date           string         `json:"date"`
	Country        string         `json:"country"`
	Score          int            `json:"score"`
	ArtistCredit   []ArtistCredit `json:"artist-credit"`
	ReleaseGroup   ReleaseGroup   `json:"release-group"`
	LabelInfo      []LabelInfo    `json:"label-info"`
	Barcode        string         `json:"barcode"`
	MediaCount     int            `json:"media-count"`
	TrackCount     int            `json:"track-count"`
}

// ReleaseDetails represents detailed release information
type ReleaseDetails struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Status         string         `json:"status"`
	Date           string         `json:"date"`
	Country        string         `json:"country"`
	Barcode        string         `json:"barcode"`
	ArtistCredit   []ArtistCredit `json:"artist-credit"`
	ReleaseGroup   ReleaseGroup   `json:"release-group"`
	LabelInfo      []LabelInfo    `json:"label-info"`
	Media          []Media        `json:"media"`
}

// ReleaseGroup represents a group of releases
type ReleaseGroup struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	PrimaryType    string   `json:"primary-type"`
	SecondaryTypes []string `json:"secondary-types"`
	FirstReleaseDate string `json:"first-release-date"`
}

// ArtistCredit represents an artist credit
type ArtistCredit struct {
	Name   string `json:"name"`
	Artist Artist `json:"artist"`
}

// Artist represents a MusicBrainz artist
type Artist struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	SortName string `json:"sort-name"`
	Type     string `json:"type"`
}

// LabelInfo represents label information
type LabelInfo struct {
	CatalogNumber string `json:"catalog-number"`
	Label         Label  `json:"label"`
}

// Label represents a record label
type Label struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Media represents a medium in a release
type Media struct {
	Format     string  `json:"format"`
	TrackCount int     `json:"track-count"`
	Position   int     `json:"position"`
	Tracks     []Track `json:"tracks"`
}

// Track represents a track/recording
type Track struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Length    int       `json:"length"`
	Position  int       `json:"position"`
	Number    string    `json:"number"`
	Recording Recording `json:"recording"`
}

// Recording represents a MusicBrainz recording
type Recording struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Length       int            `json:"length"`
	ArtistCredit []ArtistCredit `json:"artist-credit"`
}

// SearchArtistResponse represents the MusicBrainz artist search API response
type SearchArtistResponse struct {
	Count   int      `json:"count"`
	Offset  int      `json:"offset"`
	Artists []Artist `json:"artists"`
}

// ArtistDetails represents detailed artist information
type ArtistDetails struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	SortName string `json:"sort-name"`
	Type     string `json:"type"`
	Country  string `json:"country"`
	Gender   string `json:"gender"`
	Aliases  []Alias `json:"aliases"`
}

// Alias represents an artist alias
type Alias struct {
	Name   string `json:"name"`
	Locale string `json:"locale"`
	Type   string `json:"type"`
}

// CachedResponse represents a cached API response
type CachedResponse struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       int         `json:"ttl"` // seconds
}

// ErrorResponse represents a MusicBrainz API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Help    string `json:"help"`
}
