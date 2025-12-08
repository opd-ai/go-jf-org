# Transaction Format Documentation

## Overview

go-jf-org uses a transaction-based safety system to track all file operations. This allows users to review and rollback organization operations if needed.

## Transaction Storage

Transactions are stored as JSON files in the user's home directory:

```
~/.go-jf-org/txn/<transaction-id>.json
```

Each transaction is identified by a unique, cryptographically random 16-character hexadecimal ID.

## Transaction Structure

### Transaction Object

```json
{
  "id": "d8f309ee07381295",
  "timestamp": "2025-12-08T03:48:36Z",
  "operations": [...],
  "status": "completed",
  "completed": "2025-12-08T03:48:36Z",
  "error": ""
}
```

**Fields:**

- `id` (string): Unique transaction identifier
- `timestamp` (string): ISO 8601 timestamp when transaction was created
- `operations` (array): List of file operations in this transaction
- `status` (string): Transaction status (see Status Values below)
- `completed` (string, optional): ISO 8601 timestamp when transaction completed/failed
- `error` (string, optional): Error message if transaction failed

### Operation Object

```json
{
  "type": "move",
  "source": "/media/unsorted/The.Matrix.1999.1080p.mkv",
  "destination": "/media/movies/The Matrix (1999)/The Matrix (1999).mkv",
  "status": "completed",
  "error": null
}
```

**Fields:**

- `type` (string): Operation type (see Operation Types below)
- `source` (string): Source file path (for move/rename operations)
- `destination` (string): Destination file path
- `status` (string): Operation status (see Status Values below)
- `error` (object, optional): Error information if operation failed

### Status Values

**Transaction Status:**
- `pending`: Transaction is in progress
- `completed`: All operations completed successfully
- `failed`: One or more operations failed
- `rolled_back`: Transaction has been reversed

**Operation Status:**
- `pending`: Operation not yet executed
- `in_progress`: Operation is executing
- `completed`: Operation completed successfully
- `failed`: Operation failed
- `rolled_back`: Operation was reversed

### Operation Types

- `move`: Move a file from source to destination
- `rename`: Rename a file (essentially same as move)
- `create_dir`: Create a new directory
- `create_file`: Create a new file (e.g., NFO file)

## Example Transaction

### Complete Transaction

```json
{
  "id": "d8f309ee07381295",
  "timestamp": "2025-12-08T03:48:36.123456789Z",
  "operations": [
    {
      "type": "move",
      "source": "/media/unsorted/The.Matrix.1999.1080p.mkv",
      "destination": "/media/movies/The Matrix (1999)/The Matrix (1999).mkv",
      "status": "completed",
      "error": null
    },
    {
      "type": "move",
      "source": "/media/unsorted/Breaking.Bad.S01E01.720p.mkv",
      "destination": "/media/tv/Breaking Bad/Season 01/Breaking Bad - S01E01.mkv",
      "status": "completed",
      "error": null
    }
  ],
  "status": "completed",
  "completed": "2025-12-08T03:48:36.234567890Z",
  "error": ""
}
```

### Failed Transaction

```json
{
  "id": "abc123def456",
  "timestamp": "2025-12-08T03:48:36.123456789Z",
  "operations": [
    {
      "type": "move",
      "source": "/media/unsorted/file.mkv",
      "destination": "/readonly/dest.mkv",
      "status": "failed",
      "error": {
        "message": "failed to move file: permission denied"
      }
    }
  ],
  "status": "failed",
  "completed": "2025-12-08T03:48:36.234567890Z",
  "error": "some operations failed"
}
```

## Transaction Lifecycle

1. **Begin**: Transaction created with `pending` status, empty operations list
2. **Add Operations**: Operations added with `pending` status before execution
3. **Execute**: Operations executed one by one, status updated to `in_progress` then `completed`/`failed`
4. **Complete/Fail**: Transaction marked as `completed` or `failed` based on results
5. **Rollback** (optional): Transaction status changed to `rolled_back`, files restored

## Rollback Behavior

When rolling back a transaction:

1. Operations are reversed in **reverse order** (last to first)
2. Only `completed` operations are rolled back
3. For `move` operations: File moved back from destination to source
4. For `create_file` operations: Created file is deleted
5. For `create_dir` operations: Directory removed if empty
6. Empty parent directories are recursively cleaned up
7. Transaction status updated to `rolled_back`

## CLI Usage

### View Transactions

```bash
# List all transactions
go-jf-org rollback --list

# Show transaction details
go-jf-org rollback <transaction-id> --show
```

### Rollback

```bash
# Rollback a specific transaction
go-jf-org rollback <transaction-id>
```

### Disable Transactions

```bash
# Organize without transaction logging
go-jf-org organize /media/unsorted --no-transaction
```

## Best Practices

1. **Review Before Rollback**: Always use `--show` to review what will be rolled back
2. **Check Status**: Only `completed` or `failed` transactions can be rolled back
3. **Verify Files**: Ensure files haven't been modified externally before rollback
4. **Backup Important Data**: Transaction system is safe, but backups are always recommended

## Limitations

1. Cannot rollback if:
   - Destination file has been deleted
   - Source location is now occupied by another file
   - Parent directories were removed externally
2. `pending` and `in_progress` transactions cannot be rolled back
3. Already `rolled_back` transactions cannot be rolled back again

## Security Considerations

- Transaction files contain full file paths
- Stored in user home directory with 0644 permissions
- No sensitive data (file contents) is logged
- Transaction IDs are cryptographically random (not guessable)
