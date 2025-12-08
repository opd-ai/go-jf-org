package organizer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/opd-ai/go-jf-org/internal/detector"
	"github.com/opd-ai/go-jf-org/internal/jellyfin"
	"github.com/opd-ai/go-jf-org/internal/metadata"
	"github.com/opd-ai/go-jf-org/internal/safety"
	"github.com/opd-ai/go-jf-org/pkg/types"
)

// Organizer handles file organization operations
type Organizer struct {
	detector          detector.Detector
	parser            metadata.Parser
	naming            *jellyfin.Naming
	nfoGenerator      *jellyfin.NFOGenerator
	dryRun            bool
	createNFO         bool
	transactionMgr    *safety.TransactionManager
	enableTransactions bool
}

// NewOrganizer creates a new organizer instance
func NewOrganizer(dryRun bool) *Organizer {
	return &Organizer{
		detector:           detector.New(),
		parser:             metadata.NewParser(),
		naming:             jellyfin.NewNaming(),
		nfoGenerator:       jellyfin.NewNFOGenerator(),
		dryRun:             dryRun,
		createNFO:          false,
		enableTransactions: false,
	}
}

// NewOrganizerWithTransactions creates a new organizer with transaction logging
func NewOrganizerWithTransactions(dryRun bool, tm *safety.TransactionManager) *Organizer {
	return &Organizer{
		detector:           detector.New(),
		parser:             metadata.NewParser(),
		naming:             jellyfin.NewNaming(),
		nfoGenerator:       jellyfin.NewNFOGenerator(),
		dryRun:             dryRun,
		createNFO:          false,
		transactionMgr:     tm,
		enableTransactions: tm != nil,
	}
}

// SetCreateNFO enables or disables NFO file creation
func (o *Organizer) SetCreateNFO(create bool) {
	o.createNFO = create
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
				newPath, err := findAvailableName(plan.DestinationPath)
				if err != nil {
					log.Error().Err(err).Str("file", plan.SourcePath).Msg("Failed to find available name")
					continue
				}
				plan.DestinationPath = newPath
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
			
			// Create NFO files after successful move
			nfoOps, err := o.createNFOFiles(plan)
			if err != nil {
				log.Warn().Err(err).Str("file", plan.DestinationPath).Msg("Failed to create NFO files")
			} else if len(nfoOps) > 0 {
				operations = append(operations, nfoOps...)
			}
		}

		operations = append(operations, op)
	}

	return operations, nil
}

// ExecuteWithTransaction performs the organization with transaction logging
func (o *Organizer) ExecuteWithTransaction(plans []Plan, conflictStrategy string) (string, []types.Operation, error) {
	if !o.enableTransactions || o.transactionMgr == nil {
		ops, err := o.Execute(plans, conflictStrategy)
		return "", ops, err
	}

	// Begin transaction
	txn, err := o.transactionMgr.Begin()
	if err != nil {
		return "", nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	log.Info().Str("transaction", txn.ID).Int("plans", len(plans)).Msg("Starting transaction")

	operations := make([]types.Operation, 0, len(plans))
	operationIndices := make(map[int]int) // maps operations index to transaction index
	hasErrors := false

	for _, plan := range plans {
		// Handle conflicts
		if plan.Conflict {
			switch conflictStrategy {
			case "skip":
				log.Info().Str("file", plan.SourcePath).Msg("Skipping due to conflict")
				continue
			case "rename":
				// Add suffix to destination
				newPath, err := findAvailableName(plan.DestinationPath)
				if err != nil {
					log.Error().Err(err).Str("file", plan.SourcePath).Msg("Failed to find available name")
					continue
				}
				plan.DestinationPath = newPath
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
			txnIndex := len(txn.Operations)
			o.transactionMgr.AddOperation(txn, op)
			operationIndices[len(operations)-1] = txnIndex
			continue
		}

		// Log operation before executing
		txnIndex := len(txn.Operations)
		o.transactionMgr.AddOperation(txn, op)
		operationIndices[len(operations)] = txnIndex

		// Create destination directory
		destDir := filepath.Dir(plan.DestinationPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			op.Status = types.OperationStatusFailed
			op.Error = fmt.Errorf("failed to create directory: %w", err)
			log.Error().Err(err).Str("dir", destDir).Msg("Failed to create destination directory")
			operations = append(operations, op)
			hasErrors = true
			continue
		}

		// Move file
		log.Info().Str("source", op.Source).Str("dest", op.Destination).Msg("Moving file")
		op.Status = types.OperationStatusInProgress

		if err := os.Rename(op.Source, op.Destination); err != nil {
			op.Status = types.OperationStatusFailed
			op.Error = fmt.Errorf("failed to move file: %w", err)
			log.Error().Err(err).Str("source", op.Source).Str("dest", op.Destination).Msg("Failed to move file")
			hasErrors = true
		} else {
			op.Status = types.OperationStatusCompleted
			log.Info().Str("source", op.Source).Str("dest", op.Destination).Msg("File moved successfully")
			
			// Create NFO files after successful move
			nfoOps, err := o.createNFOFiles(plan)
			if err != nil {
				log.Warn().Err(err).Str("file", plan.DestinationPath).Msg("Failed to create NFO files")
			} else if len(nfoOps) > 0 {
				for _, nfoOp := range nfoOps {
					o.transactionMgr.AddOperation(txn, nfoOp)
					operations = append(operations, nfoOp)
				}
			}
		}

		operations = append(operations, op)
		
		// Update operation status in transaction
		if idx, ok := operationIndices[len(operations)-1]; ok {
			o.transactionMgr.UpdateOperation(txn, idx, op)
		}
	}

	// Complete or fail transaction
	if hasErrors {
		o.transactionMgr.Fail(txn, fmt.Errorf("some operations failed"))
		log.Warn().Str("transaction", txn.ID).Msg("Transaction completed with errors")
	} else {
		o.transactionMgr.Complete(txn)
		log.Info().Str("transaction", txn.ID).Msg("Transaction completed successfully")
	}

	return txn.ID, operations, nil
}

// findAvailableName finds an available filename by adding a suffix
// Returns an error if no available name can be found after 1000 attempts
func findAvailableName(path string) (string, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	for i := 1; i < 1000; i++ {
		newName := fmt.Sprintf("%s-%d%s", name, i, ext)
		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath, nil
		}
	}

	// If we somehow exhaust 1000 tries, return error
	return "", fmt.Errorf("could not find available filename after 1000 attempts for %s", path)
}

// createNFOFiles creates NFO files for the media based on type and metadata
func (o *Organizer) createNFOFiles(plan Plan) ([]types.Operation, error) {
	if !o.createNFO {
		return nil, nil
	}

	operations := make([]types.Operation, 0)
	destDir := filepath.Dir(plan.DestinationPath)

	switch plan.MediaType {
	case types.MediaTypeMovie:
		// Create movie.nfo in the movie directory
		nfoPath := filepath.Join(destDir, "movie.nfo")
		content, err := o.nfoGenerator.GenerateMovieNFO(plan.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to generate movie NFO: %w", err)
		}

		op := types.Operation{
			Type:        types.OperationCreateFile,
			Source:      "",
			Destination: nfoPath,
			Status:      types.OperationStatusPending,
		}

		if !o.dryRun {
			if err := os.WriteFile(nfoPath, []byte(content), 0644); err != nil {
				op.Status = types.OperationStatusFailed
				op.Error = fmt.Errorf("failed to write NFO file: %w", err)
			} else {
				op.Status = types.OperationStatusCompleted
				log.Info().Str("path", nfoPath).Msg("Created movie NFO file")
			}
		} else {
			op.Status = types.OperationStatusCompleted
			log.Info().Str("path", nfoPath).Msg("[DRY-RUN] Would create movie NFO file")
		}

		operations = append(operations, op)

	case types.MediaTypeTV:
		if plan.Metadata.TVMetadata == nil {
			return nil, nil
		}

		tv := plan.Metadata.TVMetadata
		
		// Create tvshow.nfo in the show directory (parent of season directory)
		showDir := filepath.Dir(destDir)
		tvshowNFOPath := filepath.Join(showDir, "tvshow.nfo")
		
		// Only create tvshow.nfo if it doesn't exist (multiple episodes share same show)
		if _, err := os.Stat(tvshowNFOPath); os.IsNotExist(err) {
			content, err := o.nfoGenerator.GenerateTVShowNFO(plan.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to generate TV show NFO: %w", err)
			}

			op := types.Operation{
				Type:        types.OperationCreateFile,
				Source:      "",
				Destination: tvshowNFOPath,
				Status:      types.OperationStatusPending,
			}

			if !o.dryRun {
				if err := os.WriteFile(tvshowNFOPath, []byte(content), 0644); err != nil {
					op.Status = types.OperationStatusFailed
					op.Error = fmt.Errorf("failed to write tvshow NFO: %w", err)
				} else {
					op.Status = types.OperationStatusCompleted
					log.Info().Str("path", tvshowNFOPath).Msg("Created tvshow NFO file")
				}
			} else {
				op.Status = types.OperationStatusCompleted
				log.Info().Str("path", tvshowNFOPath).Msg("[DRY-RUN] Would create tvshow NFO file")
			}

			operations = append(operations, op)
		}

		// Create season.nfo in the season directory
		seasonNFOPath := filepath.Join(destDir, "season.nfo")
		
		// Only create season.nfo if it doesn't exist (multiple episodes share same season)
		if _, err := os.Stat(seasonNFOPath); os.IsNotExist(err) {
			content, err := o.nfoGenerator.GenerateSeasonNFO(tv.Season)
			if err != nil {
				return nil, fmt.Errorf("failed to generate season NFO: %w", err)
			}

			op := types.Operation{
				Type:        types.OperationCreateFile,
				Source:      "",
				Destination: seasonNFOPath,
				Status:      types.OperationStatusPending,
			}

			if !o.dryRun {
				if err := os.WriteFile(seasonNFOPath, []byte(content), 0644); err != nil {
					op.Status = types.OperationStatusFailed
					op.Error = fmt.Errorf("failed to write season NFO: %w", err)
				} else {
					op.Status = types.OperationStatusCompleted
					log.Info().Str("path", seasonNFOPath).Msg("Created season NFO file")
				}
			} else {
				op.Status = types.OperationStatusCompleted
				log.Info().Str("path", seasonNFOPath).Msg("[DRY-RUN] Would create season NFO file")
			}

			operations = append(operations, op)
		}
	}

	return operations, nil
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
