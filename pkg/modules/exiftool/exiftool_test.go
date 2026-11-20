package exiftool

import (
	"errors"
	"testing"

	"github.com/gotenberg/gotenberg/v8/pkg/gotenberg"
)

func TestValidateMetadataEntry(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   any
		wantErr bool
	}{
		{
			name:  "accepts ordinary metadata",
			key:   "Author",
			value: "Gotenberg",
		},
		{
			name:    "rejects case-insensitive filesystem tags",
			key:     "filename",
			value:   "output.pdf",
			wantErr: true,
		},
		{
			name:    "rejects grouped filesystem tags",
			key:     "System:FileName",
			value:   "output.pdf",
			wantErr: true,
		},
		{
			name:    "rejects ExifTool options",
			key:     "csv",
			value:   "metadata.csv",
			wantErr: true,
		},
		{
			name:    "rejects line breaks in string values",
			key:     "Author",
			value:   "author\n-FileName=/tmp/injected",
			wantErr: true,
		},
		{
			name:    "rejects line breaks in string slices",
			key:     "Keywords",
			value:   []string{"safe", "unsafe\rvalue"},
			wantErr: true,
		},
		{
			name:    "rejects line breaks in interface slices",
			key:     "Keywords",
			value:   []any{"safe", "unsafe\nvalue"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMetadataEntry(tc.key, tc.value)
			if tc.wantErr {
				if !errors.Is(err, gotenberg.ErrPdfEngineMetadataValueNotSupported) {
					t.Fatalf("validateMetadataEntry() error = %v, want unsupported metadata value", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateMetadataEntry() error = %v, want nil", err)
			}
		})
	}
}
