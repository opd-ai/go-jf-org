package tmdb

import "time"

// SearchMovieResponse represents the TMDB movie search API response
type SearchMovieResponse struct {
	Page         int           `json:"page"`
	Results      []MovieResult `json:"results"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
}

// MovieResult represents a single movie result from search
type MovieResult struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Popularity       float64 `json:"popularity"`
	OriginalLanguage string  `json:"original_language"`
	GenreIDs         []int   `json:"genre_ids"`
}

// MovieDetails represents detailed movie information
type MovieDetails struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	Runtime          int     `json:"runtime"`
	Budget           int64   `json:"budget"`
	Revenue          int64   `json:"revenue"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Popularity       float64 `json:"popularity"`
	Status           string  `json:"status"`
	Tagline          string  `json:"tagline"`
	Genres           []Genre `json:"genres"`
	IMDBID           string  `json:"imdb_id"`
	OriginalLanguage string  `json:"original_language"`
}

// SearchTVResponse represents the TMDB TV search API response
type SearchTVResponse struct {
	Page         int        `json:"page"`
	Results      []TVResult `json:"results"`
	TotalPages   int        `json:"total_pages"`
	TotalResults int        `json:"total_results"`
}

// TVResult represents a single TV show result from search
type TVResult struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	OriginalName     string  `json:"original_name"`
	Overview         string  `json:"overview"`
	FirstAirDate     string  `json:"first_air_date"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Popularity       float64 `json:"popularity"`
	OriginalLanguage string  `json:"original_language"`
	GenreIDs         []int   `json:"genre_ids"`
}

// TVDetails represents detailed TV show information
type TVDetails struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	FirstAirDate     string   `json:"first_air_date"`
	LastAirDate      string   `json:"last_air_date"`
	PosterPath       string   `json:"poster_path"`
	BackdropPath     string   `json:"backdrop_path"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	Popularity       float64  `json:"popularity"`
	Status           string   `json:"status"`
	Tagline          string   `json:"tagline"`
	Genres           []Genre  `json:"genres"`
	NumberOfSeasons  int      `json:"number_of_seasons"`
	NumberOfEpisodes int      `json:"number_of_episodes"`
	Seasons          []Season `json:"seasons"`
	OriginalLanguage string   `json:"original_language"`
}

// Genre represents a movie or TV genre
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Season represents a TV show season
type Season struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	AirDate      string `json:"air_date"`
}

// CachedResponse represents a cached API response
type CachedResponse struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       int         `json:"ttl"` // seconds
}

// ErrorResponse represents a TMDB API error
type ErrorResponse struct {
	StatusMessage string `json:"status_message"`
	StatusCode    int    `json:"status_code"`
}
