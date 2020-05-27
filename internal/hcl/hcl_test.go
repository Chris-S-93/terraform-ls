package hcl

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform-ls/internal/filesystem"
	"github.com/hashicorp/terraform-ls/internal/source"
)

func TestFile_BlockAtPosition(t *testing.T) {
	testCases := []struct {
		name string

		content string
		pos     hcl.Pos

		expectedErr   error
		expectedBlock *hcl.Block
	}{
		{
			"invalid config",
			`provider "aws" {`,
			hcl.Pos{
				Line:   1,
				Column: 1,
				Byte:   0,
			},
			nil, // Expect errors to be ignored
			&hcl.Block{
				Type:   "provider",
				Labels: []string{"aws"},
			},
		},
		{
			"valid config and position",
			`provider "aws" {

}
`,
			hcl.Pos{
				Line:   2,
				Column: 1,
				Byte:   17,
			},
			nil,
			&hcl.Block{
				Type:   "provider",
				Labels: []string{"aws"},
			},
		},
		{
			"empty config and valid position",
			``,
			hcl.Pos{
				Line:   1,
				Column: 1,
				Byte:   0,
			},
			&NoBlockFoundErr{AtPos: hcl.Pos{Line: 1, Column: 1, Byte: 0}},
			nil,
		},
		{
			"empty config and out-of-range negative position",
			``,
			hcl.Pos{
				Line:   -42,
				Column: -3,
				Byte:   -46,
			},
			&InvalidHclPosErr{
				Pos:     hcl.Pos{Line: -42, Column: -3, Byte: -46},
				InRange: hcl.Range{Filename: "test.tf", Start: hcl.InitialPos, End: hcl.InitialPos},
			},
			nil,
		},
		{
			"empty config and out-of-range positive position",
			``,
			hcl.Pos{
				Line:   42,
				Column: 3,
				Byte:   46,
			},
			&InvalidHclPosErr{
				Pos:     hcl.Pos{Line: 42, Column: 3, Byte: 46},
				InRange: hcl.Range{Filename: "test.tf", Start: hcl.InitialPos, End: hcl.InitialPos},
			},
			nil,
		},
		{
			"valid config and out-of-range positive position",
			`provider "aws" {

}
`,
			hcl.Pos{
				Line:   42,
				Column: 3,
				Byte:   46,
			},
			&InvalidHclPosErr{
				Pos: hcl.Pos{Line: 42, Column: 3, Byte: 46},
				InRange: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.InitialPos,
					End:      hcl.Pos{Column: 1, Line: 4, Byte: 20},
				},
			},
			nil,
		},
		{
			"valid config and EOF position",
			`provider "aws" {

}
`,
			hcl.Pos{
				Line:   4,
				Column: 1,
				Byte:   20,
			},
			&NoBlockFoundErr{AtPos: hcl.Pos{Line: 4, Column: 1, Byte: 20}},
			nil,
		},
	}

	opts := cmp.Options{
		cmpopts.IgnoreFields(hcl.Block{},
			"Body", "DefRange", "TypeRange", "LabelRanges"),
		cmpopts.IgnoreFields(hcl.Diagnostic{}, "Subject"),
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i+1, tc.name), func(t *testing.T) {
			fsFile := filesystem.NewFile("test.tf", []byte(tc.content))
			f := NewFile(fsFile)
			fp := &testPosition{
				FileHandler: fsFile,
				pos:         tc.pos,
			}

			block, _, err := f.BlockAtPosition(fp)
			if err != nil {
				if tc.expectedErr == nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(tc.expectedErr, err, opts...); diff != "" {
					t.Fatalf("Error mismatch: %s", diff)
				}
				return
			}
			if tc.expectedErr != nil {
				t.Fatalf("Expected error: %s", tc.expectedErr)
			}

			if diff := cmp.Diff(tc.expectedBlock, block, opts...); diff != "" {
				t.Fatalf("Unexpected block difference: %s", diff)
			}

		})
	}
}

func TestFile_TokenAtPos(t *testing.T) {
	testCases := []struct {
		name string

		content string
		pos     hcl.Pos

		expectedErr   bool
		expectedToken string
	}{
		{
			name: "cursor out of range",
			content: `provider "azure" {
 location = var.loc

}`,
			pos: hcl.Pos{
				Line:   5,
				Column: 1,
				Byte:   0,
			},
			expectedErr: true,
		},
		{
			name: "cursor at first",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   1,
				Column: 1,
				Byte:   0,
			},
			expectedErr:   false,
			expectedToken: "",
		},
		{
			name: "cursor in a word",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			expectedErr:   false,
			expectedToken: "pr",
		},
		{
			name: "cursor after a word",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			expectedErr:   false,
			expectedToken: "provider",
		},
		{
			name: "cursor after a quote",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   1,
				Column: 11,
				Byte:   10,
			},
			expectedErr:   false,
			expectedToken: "",
		},
		{
			name: "cursor in a quoted word",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   1,
				Column: 13,
				Byte:   12,
			},
			expectedErr:   false,
			expectedToken: "az",
		},
		{
			name: "cursor after parentheses",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   1,
				Column: 19,
				Byte:   18,
			},
			expectedErr:   false,
			expectedToken: "",
		},
		{
			name: "cursor after space",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   2,
				Column: 3,
				Byte:   21,
			},
			expectedErr:   false,
			expectedToken: "",
		},
		{
			name: "cursor in a field",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   2,
				Column: 5,
				Byte:   23,
			},
			expectedErr:   false,
			expectedToken: "lo",
		},
		{
			name: "cursor after field",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   2,
				Column: 11,
				Byte:   29,
			},
			expectedErr:   false,
			expectedToken: "location",
		},
		{
			name: "cursor after dot",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   2,
				Column: 18,
				Byte:   36,
			},
			expectedErr:   false,
			expectedToken: "",
		},
		{
			name: "cursor after dot word",
			content: `provider "azure" {
  location = var.loc

}`,
			pos: hcl.Pos{
				Line:   2,
				Column: 21,
				Byte:   39,
			},
			expectedErr:   false,
			expectedToken: "loc",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i+1, tc.name), func(t *testing.T) {
			lines := source.MakeSourceLines("test.tf", []byte(tc.content))

			token, err := tokenAtPos(lines, tc.pos)
			if err != nil {
				if !tc.expectedErr {
					t.Fatal(err)
				}

				return
			}
			if tc.expectedErr {
				t.Fatalf("Expected error, but actual token: %q", token)
			}

			if token != tc.expectedToken {
				t.Fatalf("expect token %q but actual %q", tc.expectedToken, token)
			}
		})
	}
}

type testPosition struct {
	filesystem.FileHandler
	pos hcl.Pos
}

func (p *testPosition) Position() hcl.Pos {
	return p.pos
}
