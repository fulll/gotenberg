package pdfengines

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/gotenberg/gotenberg/v8/pkg/gotenberg"
)

func TestMultiPdfEngines_Merge(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		engine      *multiPdfEngines
		ctx         context.Context
		expectError bool
	}{
		{
			scenario: "nominal behavior",
			engine: newMultiPdfEngines(
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
							return nil
						},
					},
				},
				nil,
				nil,
				nil,
				nil,
			),
			ctx:         context.Background(),
			expectError: false,
		},
		{
			scenario: "at least one engine does not return an error",
			engine: newMultiPdfEngines(
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
							return errors.New("foo")
						},
					},
					&gotenberg.PdfEngineMock{
						MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
							return nil
						},
					},
				},
				nil,
				nil,
				nil,
				nil,
			),
			ctx:         context.Background(),
			expectError: false,
		},
		{
			scenario: "all engines return an error",
			engine: newMultiPdfEngines(
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
							return errors.New("foo")
						},
					},
					&gotenberg.PdfEngineMock{
						MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
							return errors.New("foo")
						},
					},
				},
				nil,
				nil,
				nil,
				nil,
			),
			ctx:         context.Background(),
			expectError: true,
		},
		{
			scenario: "context expired",
			engine: newMultiPdfEngines(
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						MergeMock: func(ctx context.Context, logger *zap.Logger, inputPaths []string, outputPath string) error {
							return nil
						},
					},
				},
				nil,
				nil,
				nil,
				nil,
			),
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
			expectError: true,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			err := tc.engine.Merge(tc.ctx, zap.NewNop(), nil, "")

			if !tc.expectError && err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}

			if tc.expectError && err == nil {
				t.Fatal("expected error but got none")
			}
		})
	}
}

func TestMultiPdfEngines_Convert(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		engine      *multiPdfEngines
		ctx         context.Context
		expectError bool
	}{
		{
			scenario: "nominal behavior",
			engine: newMultiPdfEngines(
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						ConvertMock: func(ctx context.Context, logger *zap.Logger, formats gotenberg.PdfFormats, inputPath, outputPath string) error {
							return nil
						},
					},
				},
				nil,
				nil,
				nil,
			),
			ctx: context.Background(),
		},
		{
			scenario: "at least one engine does not return an error",
			engine: newMultiPdfEngines(
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						ConvertMock: func(ctx context.Context, logger *zap.Logger, formats gotenberg.PdfFormats, inputPath, outputPath string) error {
							return errors.New("foo")
						},
					},
					&gotenberg.PdfEngineMock{
						ConvertMock: func(ctx context.Context, logger *zap.Logger, formats gotenberg.PdfFormats, inputPath, outputPath string) error {
							return nil
						},
					},
				},
				nil,
				nil,
				nil,
			),
			ctx: context.Background(),
		},
		{
			scenario: "all engines return an error",
			engine: newMultiPdfEngines(
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						ConvertMock: func(ctx context.Context, logger *zap.Logger, formats gotenberg.PdfFormats, inputPath, outputPath string) error {
							return errors.New("foo")
						},
					},
					&gotenberg.PdfEngineMock{
						ConvertMock: func(ctx context.Context, logger *zap.Logger, formats gotenberg.PdfFormats, inputPath, outputPath string) error {
							return errors.New("foo")
						},
					},
				},
				nil,
				nil,
				nil,
			),
			ctx:         context.Background(),
			expectError: true,
		},
		{
			scenario: "context expired",
			engine: newMultiPdfEngines(
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						ConvertMock: func(ctx context.Context, logger *zap.Logger, formats gotenberg.PdfFormats, inputPath, outputPath string) error {
							return nil
						},
					},
				},
				nil,
				nil,
				nil,
			),
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
			expectError: true,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			err := tc.engine.Convert(tc.ctx, zap.NewNop(), gotenberg.PdfFormats{}, "", "")

			if !tc.expectError && err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}

			if tc.expectError && err == nil {
				t.Fatal("expected error but got none")
			}
		})
	}
}

func TestMultiPdfEngines_ReadMetadata(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		engine      *multiPdfEngines
		ctx         context.Context
		expectError bool
	}{
		{
			scenario: "nominal behavior",
			engine: newMultiPdfEngines(
				nil,
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						ReadMetadataMock: func(ctx context.Context, logger *zap.Logger, inputPath string) (map[string]interface{}, error) {
							return make(map[string]interface{}), nil
						},
					},
				},
				nil,
				nil,
			),
			ctx: context.Background(),
		},
		{
			scenario: "at least one engine does not return an error",
			engine: newMultiPdfEngines(
				nil,
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						ReadMetadataMock: func(ctx context.Context, logger *zap.Logger, inputPath string) (map[string]interface{}, error) {
							return nil, errors.New("foo")
						},
					},
					&gotenberg.PdfEngineMock{
						ReadMetadataMock: func(ctx context.Context, logger *zap.Logger, inputPath string) (map[string]interface{}, error) {
							return make(map[string]interface{}), nil
						},
					},
				},
				nil,
				nil,
			),
			ctx: context.Background(),
		},
		{
			scenario: "all engines return an error",
			engine: newMultiPdfEngines(
				nil,
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						ReadMetadataMock: func(ctx context.Context, logger *zap.Logger, inputPath string) (map[string]interface{}, error) {
							return nil, errors.New("foo")
						},
					},
					&gotenberg.PdfEngineMock{
						ReadMetadataMock: func(ctx context.Context, logger *zap.Logger, inputPath string) (map[string]interface{}, error) {
							return nil, errors.New("foo")
						},
					},
				},
				nil,
				nil,
			),
			ctx:         context.Background(),
			expectError: true,
		},
		{
			scenario: "context expired",
			engine: newMultiPdfEngines(
				nil,
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						ReadMetadataMock: func(ctx context.Context, logger *zap.Logger, inputPath string) (map[string]interface{}, error) {
							return make(map[string]interface{}), nil
						},
					},
				},
				nil,
				nil,
			),
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
			expectError: true,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			_, err := tc.engine.ReadMetadata(tc.ctx, zap.NewNop(), "")

			if !tc.expectError && err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}

			if tc.expectError && err == nil {
				t.Fatal("expected error but got none")
			}
		})
	}
}

func TestMultiPdfEngines_WriteMetadata(t *testing.T) {
	for _, tc := range []struct {
		scenario    string
		engine      *multiPdfEngines
		ctx         context.Context
		expectError bool
	}{
		{
			scenario: "nominal behavior",
			engine: newMultiPdfEngines(
				nil,
				nil,
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						WriteMetadataMock: func(ctx context.Context, logger *zap.Logger, metadata map[string]interface{}, inputPath string) error {
							return nil
						},
					},
				},
				nil,
			),
			ctx: context.Background(),
		},
		{
			scenario: "at least one engine does not return an error",
			engine: newMultiPdfEngines(
				nil,
				nil,
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						WriteMetadataMock: func(ctx context.Context, logger *zap.Logger, metadata map[string]interface{}, inputPath string) error {
							return errors.New("foo")
						},
					},
					&gotenberg.PdfEngineMock{
						WriteMetadataMock: func(ctx context.Context, logger *zap.Logger, metadata map[string]interface{}, inputPath string) error {
							return nil
						},
					},
				},
				nil,
			),
			ctx: context.Background(),
		},
		{
			scenario: "all engines return an error",
			engine: newMultiPdfEngines(
				nil,
				nil,
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						WriteMetadataMock: func(ctx context.Context, logger *zap.Logger, metadata map[string]interface{}, inputPath string) error {
							return errors.New("foo")
						},
					},
					&gotenberg.PdfEngineMock{
						WriteMetadataMock: func(ctx context.Context, logger *zap.Logger, metadata map[string]interface{}, inputPath string) error {
							return errors.New("foo")
						},
					},
				},
				nil,
			),
			ctx:         context.Background(),
			expectError: true,
		},
		{
			scenario: "context expired",
			engine: newMultiPdfEngines(
				nil,
				nil,
				nil,
				[]gotenberg.PdfEngine{
					&gotenberg.PdfEngineMock{
						WriteMetadataMock: func(ctx context.Context, logger *zap.Logger, metadata map[string]interface{}, inputPath string) error {
							return nil
						},
					},
				},
				nil,
			),
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
			expectError: true,
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			err := tc.engine.WriteMetadata(tc.ctx, zap.NewNop(), nil, "")

			if !tc.expectError && err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}

			if tc.expectError && err == nil {
				t.Fatal("expected error but got none")
			}
		})
	}
}
