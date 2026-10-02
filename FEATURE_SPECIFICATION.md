# Feature Implementation Specification: PDF Bookmarks Import

## Overview
This document specifies the changes implemented to add PDF bookmark import functionality to Gotenberg using the pdfcpu library. The feature allows users to provide bookmark data when converting HTML/Markdown to PDF via the Chromium module, which are then imported into the generated PDF using pdfcpu.

## Core Feature: PDF Bookmarks Import

### 1. PDF Engine Interface Extension

**File**: `pkg/gotenberg/pdfengine.go`

**Change**: Add a new method to the `PdfEngine` interface:

```go
// ImportBookmarks imports bookmarks from a JSON file into a given PDF.
ImportBookmarks(ctx context.Context, logger *zap.Logger, inputPath, inputBookmarksPath, outputPath string) error
```

**Parameters**:
- `inputPath`: Path to the source PDF file
- `inputBookmarksPath`: Path to the JSON file containing bookmark data (in pdfcpu format)
- `outputPath`: Path where the PDF with imported bookmarks will be saved

---

### 2. PdfCpu Implementation

**File**: `pkg/modules/pdfcpu/pdfcpu.go`

**Changes**:

1. **Update module documentation** (`doc.go`):
   - Add "2. Import bookmarks in a PDF file." to the list of capabilities

2. **Implement `ImportBookmarks` method**:

```go
func (engine *PdfCpu) ImportBookmarks(ctx context.Context, logger *zap.Logger, inputPath, inputBookmarksPath, outputPath string) error {
	if inputBookmarksPath == "" {
		return nil
	}

	var args []string
	args = append(args, "bookmarks", "import", inputPath, inputBookmarksPath, outputPath)

	cmd, err := gotenberg.CommandContext(ctx, logger, engine.binPath, args...)
	if err != nil {
		return fmt.Errorf("create command: %w", err)
	}

	_, err = cmd.Exec()
	if err == nil {
		return nil
	}

	return fmt.Errorf("ImportBookmarks PDFs with pdfcpu: %w", err)
}
```

**Logic**:
- If no bookmarks path provided, return nil (no-op)
- Execute pdfcpu command: `pdfcpu bookmarks import <inputPath> <inputBookmarksPath> <outputPath>`
- Handle errors appropriately

---

### 3. Stub Implementations for Other PDF Engines

Add `ImportBookmarks` methods returning `gotenberg.ErrPdfEngineMethodNotSupported` error to:

**Files**:
- `pkg/modules/exiftool/exiftool.go`
- `pkg/modules/libreoffice/pdfengine/pdfengine.go`
- `pkg/modules/pdftk/pdftk.go`
- `pkg/modules/qpdf/qpdf.go`

**Implementation** (same for all):

```go
func (engine *[EngineName]) ImportBookmarks(ctx context.Context, logger *zap.Logger, inputPath, inputBookmarksPath, outputPath string) error {
	return fmt.Errorf("import bookmarks into PDF with [EngineName]: %w", gotenberg.ErrPdfEngineMethodNotSupported)
}
```

---

### 4. Mock Update

**File**: `pkg/gotenberg/mocks.go`

**Changes**:

1. Add `ImportBookmarksMock` field to `PdfEngineMock` struct:

```go
type PdfEngineMock struct {
	MergeMock           func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error
	ConvertMock         func(ctx context.Context, logger *zap.Logger, formats PdfFormats, inputPath, outputPath string) error
	ReadMetadataMock    func(ctx context.Context, logger *zap.Logger, inputPath string) (map[string]interface{}, error)
	WriteMetadataMock   func(ctx context.Context, logger *zap.Logger, metadata map[string]interface{}, inputPath string) error
	ImportBookmarksMock func(ctx context.Context, logger *zap.Logger, inputPath, inputBookmarksPath, outputPath string) error
}
```

2. Implement the mock method:

```go
func (engine *PdfEngineMock) ImportBookmarks(ctx context.Context, logger *zap.Logger, inputPath, inputBookmarksPath, outputPath string) error {
	return engine.ImportBookmarksMock(ctx, logger, inputPath, inputBookmarksPath, outputPath)
}
```

---

### 5. Multi PDF Engines Support

**File**: `pkg/modules/pdfengines/multi.go`

**Changes**:

1. Add `importBookmarksEngines` field to `multiPdfEngines` struct:

```go
type multiPdfEngines struct {
	mergeEngines           []gotenberg.PdfEngine
	convertEngines         []gotenberg.PdfEngine
	readMedataEngines      []gotenberg.PdfEngine
	writeMedataEngines     []gotenberg.PdfEngine
	importBookmarksEngines []gotenberg.PdfEngine
}
```

2. Update constructor `newMultiPdfEngines` to accept the new parameter

3. Implement `ImportBookmarks` method with concurrent engine execution pattern (similar to other methods):

```go
func (multi *multiPdfEngines) ImportBookmarks(ctx context.Context, logger *zap.Logger, inputPath, inputBookmarksPath, outputPath string) error {
	var err error
	errChan := make(chan error, 1)

	for _, engine := range multi.importBookmarksEngines {
		go func(engine gotenberg.PdfEngine) {
			errChan <- engine.ImportBookmarks(ctx, logger, inputPath, inputBookmarksPath, outputPath)
		}(engine)

		select {
		case mergeErr := <-errChan:
			errored := multierr.AppendInto(&err, mergeErr)
			if !errored {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("import bookmarks into PDF with multi PDF engines: %w", err)
}
```

**Note**: The logic tries engines in order until one succeeds or all fail.

---

### 6. PDF Engines Module Configuration

**File**: `pkg/modules/pdfengines/pdfengines.go`

**Changes**:

1. Add `importBookmarksNames` field to `PdfEngines` struct:

```go
type PdfEngines struct {
	mergeNames           []string
	convertNames         []string
	readMetadataNames    []string
	writeMedataNames     []string
	importBookmarksNames []string
	engines              []gotenberg.PdfEngine
	disableRoutes        bool
}
```

2. Add flag in `Descriptor()` method:

```go
fs.StringSlice("pdfengines-import-bookmarks-engines", []string{"pdfcpu"}, "Set the PDF engines and their order for the import bookmarks feature - empty means all")
```

**Default**: `["pdfcpu"]`

3. Update `Provision()` to read and assign the flag:

```go
importBookmarksNames := flags.MustStringSlice("pdfengines-import-bookmarks-engines")
// ... later in the method
mod.importBookmarksNames = defaultNames
if len(importBookmarksNames) > 0 {
	mod.importBookmarksNames = importBookmarksNames
}
```

4. Add validation in `Validate()`:

```go
findNonExistingEngines(mod.importBookmarksNames)
```

5. Add system message in `SystemMessages()`:

```go
fmt.Sprintf("import bookmarks engines - %s", strings.Join(mod.importBookmarksNames[:], " "))
```

6. Update `PdfEngine()` method to pass import bookmarks engines to constructor:

```go
return newMultiPdfEngines(
	engines(mod.mergeNames),
	engines(mod.convertNames),
	engines(mod.readMetadataNames),
	engines(mod.writeMedataNames),
	engines(mod.importBookmarksNames),
), nil
```

---

### 7. PDF Engines Routes Helper

**File**: `pkg/modules/pdfengines/routes.go`

**Add**: New stub function `ImportBookmarksStub`:

```go
func ImportBookmarksStub(ctx *api.Context, engine gotenberg.PdfEngine, inputPath string, inputBookmarks []byte, outputPath string) (string, error) {
	if len(inputBookmarks) == 0 {
		fmt.Println("ImportBookmarksStub BM empty")
		return inputPath, nil
	}

	inputBookmarksPath := ctx.GeneratePath(".json")
	err := os.WriteFile(inputBookmarksPath, inputBookmarks, 0o600)
	if err != nil {
		return "", fmt.Errorf("write file %v: %w", inputBookmarksPath, err)
	}
	err = engine.ImportBookmarks(ctx, ctx.Log(), inputPath, inputBookmarksPath, outputPath)
	if err != nil {
		return "", fmt.Errorf("import bookmarks %v: %w", inputPath, err)
	}

	return outputPath, nil
}
```

**Logic**:
- Takes bookmark data as JSON bytes
- If empty, returns input path unchanged
- Creates temporary JSON file with bookmark data
- Calls engine's ImportBookmarks method
- Returns output path on success

**Note**: Need to import "os" package.

---

### 8. Chromium Module Integration

**File**: `pkg/modules/chromium/chromium.go`

**Changes**:

1. Import pdfcpu package:
   ```go
   import "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
   ```

2. Add `Bookmarks` field to `PdfOptions` struct:

```go
type PdfOptions struct {
	// ... existing fields ...
	
	// Bookmarks to be inserted unmarshaled
	// as defined in pdfcpu bookmarks export
	Bookmarks pdfcpu.BookmarkTree

	// ... remaining fields ...
}
```

3. Update `DefaultPdfOptions()` to initialize bookmarks:

```go
func DefaultPdfOptions() PdfOptions {
	return PdfOptions{
		// ... existing fields ...
		Bookmarks: pdfcpu.BookmarkTree{},
		// ... remaining fields ...
	}
}
```

---

**File**: `pkg/modules/chromium/routes.go`

**Changes**:

1. Import required packages:
   ```go
   import "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
   ```

2. In `FormDataChromiumPdfOptions` function, add bookmark parsing:

   a. Add variable declaration:
   ```go
   var (
       // ... existing variables ...
       bookmarks pdfcpu.BookmarkTree
   )
   ```

   b. Add custom form field handler:
   ```go
   Custom("bookmarks", func(value string) error {
       if len(value) > 0 {
           err := json.Unmarshal([]byte(value), &bookmarks)
           if err != nil {
               return fmt.Errorf("unmarshal bookmarks: %w", err)
           }
       } else {
           bookmarks = defaultPdfOptions.Bookmarks
       }
       return nil
   })
   ```

   c. Include in returned options:
   ```go
   return formData, PdfOptions{
       // ... existing fields ...
       Bookmarks: bookmarks,
       // ... remaining fields ...
   }
   ```

3. In `convertUrl` function (after PDF generation, before conversion), add bookmark import logic:

```go
if options.GenerateDocumentOutline {
	if len(options.Bookmarks.Bookmarks) > 0 {
		bookmarks, errMarshal := json.Marshal(options.Bookmarks)
		outputBMPath := ctx.GeneratePath(".pdf")

		if errMarshal == nil {
			outputPath, err = pdfengines.ImportBookmarksStub(ctx, engine, outputPath, bookmarks, outputBMPath)
			if err != nil {
				return fmt.Errorf("import bookmarks into PDF err: %w", err)
			}
		} else {
			return fmt.Errorf("import bookmarks into PDF errMarshal : %w", errMarshal)
		}
	}
}
```

**Logic**:
- Only process bookmarks if `GenerateDocumentOutline` is true and bookmarks exist
- Marshal the bookmarks back to JSON
- Generate output path for PDF with bookmarks
- Call `ImportBookmarksStub` helper
- Update `outputPath` to the new path with bookmarks
- This happens **before** the `pdfengines.ConvertStub` call

---

### 9. Test Updates

**File**: `pkg/modules/pdfengines/multi_test.go`

**Changes**: Add `nil` parameter to all `newMultiPdfEngines` calls in tests (for import bookmarks engines).

Example:
```go
newMultiPdfEngines(
	// ... existing parameters ...
	nil, // import bookmarks engines
)
```

**Locations**: All test cases in `TestMultiPdfEngines_*` functions.

---

**File**: `pkg/modules/pdfengines/pdfengines_test.go`

**Changes**:

1. Add `importBookmarksNames` field initialization in test structs:

```go
mod := PdfEngines{
	mergeNames:           []string{"foo", "bar"},
	convertNames:         []string{"foo", "bar"},
	readMetadataNames:    []string{"foo", "bar"},
	writeMedataNames:     []string{"foo", "bar"},
	importBookmarksNames: []string{"foo", "bar"},
	engines:              // ...
}
```

2. Update expected message count in `TestPdfEngines_SystemMessages`:
   - Change from `4` to `5` messages

3. Add expected message for import bookmarks:

```go
expectedMessages := []string{
	// ... existing messages ...
	fmt.Sprintf("import bookmarks engines - %s", strings.Join(mod.importBookmarksNames[:], " ")),
}
```

**Note**: Some test cases may have commented out assertions for `expectedImportBookmarksPdfEngines` - these should be implemented or left as TODOs based on project conventions.

---

## Dependencies

### Go Module Updates

**File**: `go.mod`

**Changes**:

1. Add pdfcpu dependency in require block:

```go
require (
	github.com/dlclark/regexp2 v1.11.4
	github.com/pdfcpu/pdfcpu v0.9.1
)
```

2. Add indirect dependencies:

```go
require (
	// ... existing ...
	github.com/hhrutter/lzw v1.0.0 // indirect
	github.com/hhrutter/tiff v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/image v0.21.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
```

**File**: `go.sum`

Updated with checksums for all new dependencies and their transitive dependencies.

---

## Build and Deployment Changes

### 1. Dockerfile

**File**: `build/Dockerfile`

**Changes**: Add support for pinning Chrome version via build argument:

1. Add build argument:
   ```dockerfile
   ARG CHROME_VERSION
   ```

2. Modify Chrome installation logic (line ~152) to support conditional installation:

```dockerfile
RUN \
    /bin/bash -c \
    'set -e &&\
    if [[ "$(dpkg --print-architecture)" == "amd64" ]]; then \
      apt-get update -qq &&\
      if [ -z "$CHROME_VERSION" ]; then \
        # Install latest stable version
        curl https://dl.google.com/linux/linux_signing_key.pub | apt-key add - &&\
        echo "deb http://dl.google.com/linux/chrome/deb/ stable main" | tee /etc/apt/sources.list.d/google-chrome.list &&\
        apt-get update -qq &&\
        DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends --allow-unauthenticated google-chrome-stable &&\
        mv /usr/bin/google-chrome-stable /usr/bin/chromium; \
      else \
        # Install specific version
        apt-get update -qq &&\
        curl --output /tmp/chrome.deb "https://dl.google.com/linux/chrome/deb/pool/main/g/google-chrome-stable/google-chrome-stable_${CHROME_VERSION}_amd64.deb" &&\
        DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends /tmp/chrome.deb &&\
        mv /usr/bin/google-chrome-stable /usr/bin/chromium &&\
        rm -rf /tmp/chrome.deb; \  
      fi \
    elif [[ "$(dpkg --print-architecture)" == "armhf" ]]; then \
      # ... existing armhf logic unchanged ...
```

**Logic**:
- If `CHROME_VERSION` is empty/unset: install latest stable version (original behavior)
- If `CHROME_VERSION` is set: download and install specific .deb file from Google's repository

---

### 2. Makefile

**File**: `Makefile`

**Changes**:

1. Update default Docker registry:
   ```makefile
   DOCKER_REGISTRY=ghcr.io/fulll
   ```
   (was: `DOCKER_REGISTRY=gotenberg`)

2. Add `CHROME_VERSION` build argument to `build` target:

```makefile
build:
	# ... existing arguments ...
	--build-arg CHROME_VERSION=$(CHROME_VERSION) \
	# ... rest of command ...
```

3. Add `CHROME_VERSION` to `build-tests` target:

```makefile
build-tests:
	# ... existing arguments ...
	--build-arg CHROME_VERSION=$(CHROME_VERSION) \
	# ... rest of command ...
```

4. Add `CHROME_VERSION` parameter to `release` target:

```makefile
release:
	$(PDFCPU_VERSION) \
	$(DOCKER_REGISTRY) \
	$(DOCKER_REPOSITORY) \
	$(LINUX_AMD64_RELEASE) \
	$(CHROME_VERSION)  # Add as 11th parameter
```

---

### 3. Release Script

**File**: `scripts/release.sh`

**Changes**:

1. Add `CHROME_VERSION` parameter (11th argument):
   ```bash
   CHROME_VERSION="${11}"
   ```

2. Remove multi-arch platform flag logic, force Linux AMD64 only:
   ```bash
   # Replace conditional logic with:
   PLATFORM_FLAG="--platform linux/amd64"
   ```
   (Note: Original had conditional for AMD64 only vs multi-arch)

3. Add `CHROME_VERSION` build argument to docker buildx command:

```bash
docker buildx build \
  # ... existing arguments ...
  --build-arg CHROME_VERSION="$CHROME_VERSION" \
  # ... rest of command ...
```

---

### 4. GitHub Actions CI/CD

**File**: `.github/workflows/continuous_delivery.yml`

**Changes**:

1. Update Docker registry from Docker Hub to GitHub Container Registry:

```yaml
- name: Log in to Docker Hub Container Registry
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}
```

2. Add Chrome version and release flag to build step:

```yaml
- name: Build and push Docker image for release
  env:
    LINUX_AMD64_RELEASE: "true"
  run: |
    make release CHROME_VERSION=127.0.6533.119-1 GOTENBERG_VERSION=${{ github.event.release.tag_name }} DOCKER_REGISTRY=ghcr.io/fulll DOCKER_REPOSITORY=gotenberg
```

**Specifics**:
- `CHROME_VERSION=127.0.6533.119-1` (pinned version)
- `LINUX_AMD64_RELEASE="true"`
- Registry: `ghcr.io/fulll`
- Repository: `gotenberg`

3. Add AWS ECR deployment steps:

```yaml
- name: generate aws credentials config
  env:
    AWS_CREDENTIALS: ${{ secrets.STAGING_AWS_CREDENTIALS }}
    aws-region: eu-central-1
  run: |
    mkdir -p "${HOME}/.aws"
    echo "${AWS_CREDENTIALS}" > "${HOME}/.aws/credentials"

- name: docker login and push
  run: |
    # Extract tag name and strip first letter
    TAG_NAME=$(echo "${{ github.event.release.tag_name }}" | cut -c 2-)

    docker pull ghcr.io/fulll/gotenberg:${TAG_NAME}-cloudrun
    docker tag ghcr.io/fulll/gotenberg:${TAG_NAME}-cloudrun ${AWS_ECR_REGISTRY}/gotenberg-fulll:${TAG_NAME}-cloudrun
    aws --region eu-central-1 ecr get-login-password | docker login --username AWS --password-stdin ${AWS_ECR_REGISTRY}
    docker tag ${AWS_ECR_REGISTRY}/gotenberg-fulll:${TAG_NAME}-cloudrun ${AWS_ECR_REGISTRY}/gotenberg-fulll:latest
    docker push ${AWS_ECR_REGISTRY}/gotenberg-fulll:${TAG_NAME}-cloudrun
    docker push ${AWS_ECR_REGISTRY}/gotenberg-fulll:latest
```

**Logic**:
- Setup AWS credentials from secrets
- Extract release tag (remove 'v' prefix)
- Pull cloudrun variant from GitHub Container Registry
- Tag for AWS ECR (both versioned and latest)
- Push to ECR in eu-central-1 region

**ECR Details**:
- Account ID: `private_from_secret`
- Region: `eu-central-1`
- Repository: `gotenberg-fulll`

---

## API Usage

### Request Parameters

Users can now provide bookmarks when converting HTML/Markdown to PDF via Chromium routes:

**Form Field**: `bookmarks` (string, JSON format)

**Format**: JSON string matching pdfcpu BookmarkTree structure

**Example**:
```json
{
  "Bookmarks": [
    {
      "Title": "Chapter 1",
      "PageFrom": 1,
      "PageThru": -1,
      "Kids": [
        {
          "Title": "Section 1.1",
          "PageFrom": 2,
          "PageThru": -1
        }
      ]
    }
  ]
}
```

**Behavior**:
- Bookmarks are only imported if `generateDocumentOutline` is `true`
- If bookmarks field is empty/missing, no bookmarks are added
- Invalid JSON returns error to user

---

## Implementation Notes and Clarifications

1. **Test Coverage**: 
   - In `pdfengines_test.go`, the commented-out assertions for `expectedImportBookmarksPdfEngines` are intentional
   - No additional test implementation is required beyond what's shown
   - Keep the commented code as-is

2. **Debug Logging**:
   - The `ImportBookmarksStub` function includes: `fmt.Println("ImportBookmarksStub BM empty")`
   - **Keep this logging statement** - it's intentional for debugging purposes

3. **Bookmark Validation**:
   - No additional validation of bookmark structure is needed beyond JSON unmarshaling
   - pdfcpu handles its own validation
   - Keep the current simple approach

4. **Implementation Approach**:
   - The current approach (marshal to JSON → write temp file → call pdfcpu CLI) is intentional
   - **Keep this approach** - do not refactor to use pdfcpu's Go API directly
   - This maintains consistency with how other PDF operations are handled

5. **Multi-Architecture Support**:
   - **Linux AMD64 only** is intentional and required
   - The project is customized for specific deployment needs
   - Do not attempt to restore multi-arch support

6. **AWS ECR Deployment**:
   - AWS ECR push steps are **required and must be kept**
   - This is for the project's specific deployment pipeline
   - All AWS-related configuration should be preserved as-is

7. **Chrome Version Pinning**:
   - Chrome version **must be pinned** to a specific version for reproducible builds
   - This allows control over Chrome updates in case new versions introduce breaking changes
   - When reimplementing, update to the latest available stable Chrome version at that time, but keep it fixed (not "latest")
   - Example: If current version is `127.0.6533.119-1`, find the latest stable version and pin to that specific version number
   - Check https://dl.google.com/linux/chrome/deb/dists/stable/main/binary-amd64/Packages for available versions

---

## Implementation Checklist

When reimplementing on a newer version:

- [ ] Add pdfcpu dependency to go.mod
- [ ] Extend PdfEngine interface with ImportBookmarks method
- [ ] Implement ImportBookmarks in pdfcpu module
- [ ] Add stub implementations in other PDF engines
- [ ] Update mock implementations
- [ ] Add multi-engine support for import bookmarks
- [ ] Add configuration flag for import bookmarks engines
- [ ] Update PdfEngines module to handle import bookmarks
- [ ] Add ImportBookmarksStub helper function
- [ ] Add Bookmarks field to Chromium PdfOptions
- [ ] Add bookmarks form field parsing in Chromium routes
- [ ] Integrate bookmark import in convertUrl function
- [ ] Update all test files with new parameters
- [ ] Add Chrome version build argument to Dockerfile
- [ ] Update Makefile with CHROME_VERSION support
- [ ] Update release script
- [ ] (Optional) Update CI/CD for specific deployment needs
- [ ] Test bookmark import with sample pdfcpu bookmark JSON
- [ ] Verify all PDF engines return appropriate errors
- [ ] Validate multi-engine fallback behavior

---

## References

- pdfcpu documentation: https://github.com/pdfcpu/pdfcpu
- pdfcpu bookmark format: See pdfcpu CLI documentation for `bookmarks export` command output format
- Original commit: `67c02e41cc185765ca4775a82556d55aaf882e8f`
