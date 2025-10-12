# Zine-Layout API Commands

This directory contains individual Glazed commands for interacting with the zine-layout server REST API. Each command is implemented following the Glazed framework patterns and provides structured output with multiple format options.

## Available Commands

### Projects Management

#### `projects-list`
List all projects from the server.

```bash
go run ./cmd/zine-layout api projects-list
go run ./cmd/zine-layout api projects-list --output json
go run ./cmd/zine-layout api projects-list --fields id,name --sort-by name
go run ./cmd/zine-layout api projects-list --output csv --fields id,name,image_count
```

#### `projects-get`
Get details of a specific project by ID.

```bash
go run ./cmd/zine-layout api projects-get --id prj-12345
go run ./cmd/zine-layout api projects-get --id prj-12345 --output json
```

#### `projects-create`
Create a new project.

```bash
go run ./cmd/zine-layout api projects-create --name "My New Project"
go run ./cmd/zine-layout api projects-create --name "Test Project" --output json
```

#### `projects-delete`
Delete a project by ID.

```bash
go run ./cmd/zine-layout api projects-delete --id prj-12345
```

### Images Management

#### `images-list`
List all images in a specific project.

```bash
go run ./cmd/zine-layout api images-list --project-id prj-12345
go run ./cmd/zine-layout api images-list --project-id prj-12345 --output json
go run ./cmd/zine-layout api images-list --project-id prj-12345 --fields name,width,height
```

#### `images-upload`
Upload one or more PNG images to a project.

```bash
go run ./cmd/zine-layout api images-upload --project-id prj-12345 --files image.png
go run ./cmd/zine-layout api images-upload --project-id prj-12345 --files img1.png,img2.png,img3.png
go run ./cmd/zine-layout api images-upload --project-id prj-12345 --files photo.png --output json
```

#### `images-upload-dir`
Upload all PNG images from a directory to a project (bulk upload).

```bash
go run ./cmd/zine-layout api images-upload-dir --project-id prj-12345 --directory ./images
go run ./cmd/zine-layout api images-upload-dir --project-id prj-12345 --directory ./photos --recursive
go run ./cmd/zine-layout api images-upload-dir --project-id prj-12345 --directory ./assets --output json
```

#### `images-sync`
Sync PNG images from a directory to a project (upload missing files only).

```bash
go run ./cmd/zine-layout api images-sync --project-id prj-12345 --directory ./images
go run ./cmd/zine-layout api images-sync --project-id prj-12345 --directory ./images --dry-run
go run ./cmd/zine-layout api images-sync --project-id prj-12345 --directory ./photos --recursive
go run ./cmd/zine-layout api images-sync --project-id prj-12345 --directory ./images --output csv
```

### Presets

#### `presets-list`
List available presets.

```bash
go run ./cmd/zine-layout api presets-list
go run ./cmd/zine-layout api presets-list --output json
```

## Glazed Framework Features

All commands support the full range of Glazed framework features:

### Output Formats
- `--output table` (default) - ASCII table format
- `--output json` - JSON format
- `--output csv` - CSV format
- `--output yaml` - YAML format
- `--output markdown` - Markdown table format

### Field Selection and Filtering
- `--fields id,name,created_at` - Select specific fields
- `--filter image_count` - Remove specific fields
- `--sort-by name` - Sort by field (use `-sort-by=-name` for descending)

### Advanced Options
- `--glazed-limit 10` - Limit number of results
- `--glazed-skip 5` - Skip first N results
- `--select name` - Select single field, output one per line
- `--template '{{.name}}: {{.id}}'` - Custom Go template output

## Architecture

### File Structure
```
api/
├── common.go           # Shared HTTP utilities and command builders
├── commands.go         # Command registration functions
├── projects-list.go    # List projects command
├── projects-get.go     # Get project command
├── projects-create.go  # Create project command
├── projects-delete.go  # Delete project command
├── images-list.go      # List project images command
├── presets-list.go     # List presets command
└── README.md          # This file
```

### Command Pattern
Each command follows the same structure:
1. **Command struct** embedding `*cmds.CommandDescription`
2. **Settings struct** with `glazed.parameter` tags for type-safe parameter parsing
3. **RunIntoGlazeProcessor method** implementing the `GlazeCommand` interface
4. **Constructor function** returning the configured command
5. **Registration function** in `commands.go`

### HTTP Integration
- Uses shared `httpPostJSON` and `httpPostRaw` utilities in `common.go`
- Handles server errors and HTTP status codes appropriately
- Parses JSON responses into structured `types.Row` objects for Glazed processing

## Upload Capabilities

The CLI provides comprehensive image upload functionality:

### Single/Multiple File Upload
```bash
# Upload single file
zine-layout api images-upload --project-id prj-12345 --files image.png

# Upload multiple files
zine-layout api images-upload --project-id prj-12345 --files img1.png,img2.png,img3.png
```

### Bulk Directory Upload  
```bash
# Upload all PNG files from directory
zine-layout api images-upload-dir --project-id prj-12345 --directory ./images

# Recursive directory upload
zine-layout api images-upload-dir --project-id prj-12345 --directory ./photos --recursive
```

### Smart Synchronization
```bash
# Dry run - see what would be uploaded
zine-layout api images-sync --project-id prj-12345 --directory ./images --dry-run

# Upload only missing files
zine-layout api images-sync --project-id prj-12345 --directory ./images

# Recursive sync
zine-layout api images-sync --project-id prj-12345 --directory ./photos --recursive
```

**Key Features:**
- ✅ **Multipart Upload**: Proper HTTP multipart form handling
- ✅ **Batch Processing**: Large directories uploaded in manageable batches  
- ✅ **Smart Sync**: Compares local files with server, uploads missing only
- ✅ **Dry Run**: Preview what would be uploaded without actually uploading
- ✅ **Recursive**: Handle nested directory structures
- ✅ **Error Handling**: Validates file existence and PNG format
- ✅ **Progress Reporting**: Clear status for each file and batch

## Testing

Test the commands with a running zine-layout server:

```bash
# Start the server (in separate terminal)
go run ./cmd/zine-layout serve --addr :8088 --data-root ./data

# Test commands
go run ./cmd/zine-layout api projects-list
go run ./cmd/zine-layout api projects-create --name "Test"
go run ./cmd/zine-layout api projects-get --id <project-id>
go run ./cmd/zine-layout api images-list --project-id <project-id>
go run ./cmd/zine-layout api presets-list
go run ./cmd/zine-layout api projects-delete --id <project-id>

# Test upload functionality
go run ./cmd/zine-layout api images-upload --project-id <id> --files image.png
go run ./cmd/zine-layout api images-upload-dir --project-id <id> --directory ./images
go run ./cmd/zine-layout api images-sync --project-id <id> --directory ./images --dry-run
```

## Extending

To add new commands:
1. Create a new `.go` file following the existing pattern
2. Implement the `GlazeCommand` interface
3. Add a registration function to `commands.go`
4. Call the registration function in `AddAllAPICommands`

Example endpoints that could still be added:
- `images-reorder` - Reorder project images
- `images-delete` - Delete specific images
- `spreads-list` - List project spreads
- `spreads-get/put/delete` - Manage individual spreads
