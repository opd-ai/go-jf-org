package types

// MediaType represents the type of media file
type MediaType string

const (
	// MediaTypeMovie represents a movie file
	MediaTypeMovie MediaType = "movie"
	// MediaTypeTV represents a TV show episode file
	MediaTypeTV MediaType = "tv"
	// MediaTypeMusic represents a music file
	MediaTypeMusic MediaType = "music"
	// MediaTypeBook represents a book file
	MediaTypeBook MediaType = "book"
	// MediaTypeUnknown represents an unknown media type
	MediaTypeUnknown MediaType = "unknown"
)

// MediaFile represents a media file with its metadata
type MediaFile struct {
	// Path is the absolute path to the media file
	Path string
	// Type is the detected media type
	Type MediaType
	// Metadata contains extracted and enriched metadata
	Metadata *Metadata
}

// Metadata contains information about a media file
type Metadata struct {
	// Title is the primary title
	Title string
	// Year is the release year
	Year int
	// Quality contains quality information (1080p, 4K, etc.)
	Quality string
	// Source contains source information (BluRay, WEB-DL, etc.)
	Source string
	// Codec contains codec information (x264, h265, etc.)
	Codec string
	// Additional metadata specific to media type
	MovieMetadata *MovieMetadata
	TVMetadata    *TVMetadata
	MusicMetadata *MusicMetadata
	BookMetadata  *BookMetadata
}

// MovieMetadata contains movie-specific metadata
type MovieMetadata struct {
	OriginalTitle string
	Plot          string
	Director      []string
	Cast          []string
	Genres        []string
	Rating        float64
	TMDBID        int
	IMDBID        string
}

// TVMetadata contains TV show-specific metadata
type TVMetadata struct {
	ShowTitle    string
	Season       int
	Episode      int
	EpisodeTitle string
	Plot         string
	AirDate      string
	TMDBID       int
	TVDBID       int
}

// MusicMetadata contains music-specific metadata
type MusicMetadata struct {
	Artist         string
	Album          string
	AlbumArtist    string
	TrackNumber    int
	DiscNumber     int
	Genre          string
	MusicBrainzID  string
	MusicBrainzRID string
}

// BookMetadata contains book-specific metadata
type BookMetadata struct {
	Author      string
	Publisher   string
	ISBN        string
	Series      string
	SeriesIndex int
	Description string
}

// Operation represents a file operation to be performed
type Operation struct {
	// Type is the operation type (move, rename, create)
	Type OperationType
	// Source is the source path
	Source string
	// Destination is the destination path
	Destination string
	// Status is the current status of the operation
	Status OperationStatus
	// Error contains any error that occurred
	Error error
}

// OperationType represents the type of operation
type OperationType string

const (
	// OperationMove represents a file move operation
	OperationMove OperationType = "move"
	// OperationRename represents a file rename operation
	OperationRename OperationType = "rename"
	// OperationCreateDir represents a directory creation operation
	OperationCreateDir OperationType = "create_dir"
	// OperationCreateFile represents a file creation operation (e.g., NFO)
	OperationCreateFile OperationType = "create_file"
)

// OperationStatus represents the status of an operation
type OperationStatus string

const (
	// OperationStatusPending represents a pending operation
	OperationStatusPending OperationStatus = "pending"
	// OperationStatusInProgress represents an in-progress operation
	OperationStatusInProgress OperationStatus = "in_progress"
	// OperationStatusCompleted represents a completed operation
	OperationStatusCompleted OperationStatus = "completed"
	// OperationStatusFailed represents a failed operation
	OperationStatusFailed OperationStatus = "failed"
	// OperationStatusRolledBack represents a rolled back operation
	OperationStatusRolledBack OperationStatus = "rolled_back"
)
