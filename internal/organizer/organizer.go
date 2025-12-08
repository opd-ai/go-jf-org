package organizer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/opd-ai/go-jf-org/internal/detector"
	"github.com/opd-ai/go-jf-org/internal/jellyfin"
	"github.com/opd-ai/go-jf-org/internal/metadata"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

// Organizer handles file organization operations
type Organizer struct {
	detector detector.Detector
	parser   metadata.Parser
	naming   *jellyfin.Naming
	dryRun   bool
}

// NewOrganizer creates a new organizer instance
func NewOrganizer(dryRun bool) *Organizer {
	return &Organizer{
		detector: detector.New(),
		parser:   metadata.NewParser(),
		naming:   jellyfin.NewNaming(),
		dryRun:   dryRun,
	}
}

// Plan represents a planned organization operation
type Plan struct {
	SourcePath      string
	DestinationPath string
	MediaType       types.MediaType
	Metadata        *types.Metadata
	Operation       types.OperationType
	Conflict        bool
	ConflictReason  string
}

// PlanOrganization analyzes files and creates a plan without executing
func (o *Organizer) PlanOrganization(files []string, destRoot string, mediaTypeFilter types.MediaType) ([]Plan, error) {
	plans := make([]Plan, 0, len(files))

	for _, file := range files {
		// Detect media type
		mediaType := o.detector.Detect(filepath.Base(file))

		// Skip if filtering by type and doesn't match
		if mediaTypeFilter != "" && mediaTypeFilter != types.MediaTypeUnknown && mediaType != mediaTypeFilter {
			log.Debug().Str("file", file).Str("type", string(mediaType)).Str("filter", string(mediaTypeFilter)).Msg("Skipping due to type filter")
			continue
		}

		// Skip unknown types
		if mediaType == types.MediaTypeUnknown {
			log.Debug().Str("file", file).Msg("Skipping unknown media type")
			continue
		}

		// Parse metadata
		meta, err := o.parser.Parse(filepath.Base(file), mediaType)
		if err != nil {
			log.Warn().Err(err).Str("file", file).Msg("Failed to parse metadata, skipping")
			continue
		}

		// Build destination path
		ext := filepath.Ext(file)
		destPath := o.naming.BuildFullPath(destRoot, mediaType, meta, ext)
		if destPath == "" {
			log.Warn().Str("file", file).Str("type", string(mediaType)).Msg("Could not build destination path, skipping")
			continue
		}

		plan := Plan{
			SourcePath:      file,
			DestinationPath: destPath,
			MediaType:       mediaType,
			Metadata:        meta,
			Operation:       types.OperationMove,
		}

		// Check for conflicts
		if _, err := os.Stat(destPath); err == nil {
			plan.Conflict = true
			plan.ConflictReason = "destination file already exists"
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

// Execute performs the organization based on the plan
func (o *Organizer) Execute(plans []Plan, conflictStrategy string) ([]types.Operation, error) {
	operations := make([]types.Operation, 0, len(plans))

	for _, plan := range plans {
		// Handle conflicts
		if plan.Conflict {
			switch conflictStrategy {
			case "skip":
				log.Info().Str("file", plan.SourcePath).Msg("Skipping due to conflict")
				continue
			case "rename":
				// Add suffix to destination
				plan.DestinationPath = findAvailableName(plan.DestinationPath)
				log.Info().Str("file", plan.SourcePath).Str("new_dest", plan.DestinationPath).Msg("Renamed due to conflict")
			default:
				log.Warn().Str("file", plan.SourcePath).Msg("Unknown conflict strategy, skipping")
				continue
			}
		}

		op := types.Operation{
			Type:        plan.Operation,
			Source:      plan.SourcePath,
			Destination: plan.DestinationPath,
			Status:      types.OperationStatusPending,
		}

		if o.dryRun {
			log.Info().Str("source", op.Source).Str("dest", op.Destination).Msg("[DRY-RUN] Would move file")
			op.Status = types.OperationStatusCompleted
			operations = append(operations, op)
			continue
		}

		// Create destination directory
		destDir := filepath.Dir(plan.DestinationPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			op.Status = types.OperationStatusFailed
			op.Error = fmt.Errorf("failed to create directory: %w", err)
			log.Error().Err(err).Str("dir", destDir).Msg("Failed to create destination directory")
			operations = append(operations, op)
			continue
		}

		// Move file
		log.Info().Str("source", op.Source).Str("dest", op.Destination).Msg("Moving file")
		op.Status = types.OperationStatusInProgress

		if err := os.Rename(op.Source, op.Destination); err != nil {
			op.Status = types.OperationStatusFailed
			op.Error = fmt.Errorf("failed to move file: %w", err)
			log.Error().Err(err).Str("source", op.Source).Str("dest", op.Destination).Msg("Failed to move file")
		} else {
			op.Status = types.OperationStatusCompleted
			log.Info().Str("source", op.Source).Str("dest", op.Destination).Msg("File moved successfully")
		}

		operations = append(operations, op)
	}

	return operations, nil
}

// findAvailableName finds an available filename by adding a suffix
func findAvailableName(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	for i := 1; i < 1000; i++ {
		newName := fmt.Sprintf("%s-%d%s", name, i, ext)
		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}

	// If we somehow exhaust 1000 tries, return original with timestamp
	return path + ".duplicate"
}

// ValidatePlan checks if a plan can be executed safely
func (o *Organizer) ValidatePlan(plans []Plan) []error {
	errors := make([]error, 0)

	for _, plan := range plans {
		// Check source exists and is readable
		info, err := os.Stat(plan.SourcePath)
		if err != nil {
			errors = append(errors, fmt.Errorf("source file %s: %w", plan.SourcePath, err))
			continue
		}

		if info.IsDir() {
			errors = append(errors, fmt.Errorf("source %s is a directory, not a file", plan.SourcePath))
			continue
		}

		// Check destination directory would be writable
		destDir := filepath.Dir(plan.DestinationPath)
		
		// Check if parent exists
		parentInfo, err := os.Stat(filepath.Dir(destDir))
		if err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("cannot access parent directory of %s: %w", destDir, err))
			continue
		}

		// If parent exists, check if it's writable
		if parentInfo != nil && !parentInfo.IsDir() {
			errors = append(errors, fmt.Errorf("parent of destination %s is not a directory", destDir))
		}
	}

	return errors
}
