package lang

import (
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-ls/internal/mdplain"
	"github.com/hashicorp/terraform-ls/internal/terraform/schema"
)

// Parser implements a parser which can turn raw HCL block
// into ConfigBlock with the help of a schema reader
type Parser interface {
	SetLogger(*log.Logger)
	SetSchemaReader(schema.Reader)
	BlockTypeCandidates() CompletionCandidates
	ParseBlockFromHCL(*hcl.Block) (ConfigBlock, error)
}

// ConfigBlock implements an abstraction above HCL block
// which provides any LSP capabilities (e.g. completion)
type ConfigBlock interface {
	CompletionCandidatesAtPos(pos hcl.Pos) (CompletionCandidates, error)
	Name() string
	BlockType() string
	Labels() []*ParsedLabel
}

// Block represents a decoded HCL block (by a Parser)
// which keeps track of the related schema
type Block interface {
	BlockAtPos(pos hcl.Pos) (Block, bool)
	LabelAtPos(pos hcl.Pos) (*ParsedLabel, bool)
	Range() hcl.Range
	PosInLabels(pos hcl.Pos) bool
	PosInBody(pos hcl.Pos) bool
	PosInAttribute(pos hcl.Pos) bool
	Attributes() map[string]*Attribute
	BlockTypes() map[string]*BlockType
}

type LabelSchema []Label

type Label struct {
	Name          string
	IsCompletable bool
}

type ParsedLabel struct {
	Name  string
	Value string
}

type BlockType struct {
	BlockList []Block
	schema    *tfjson.SchemaBlockType
}

type Attribute struct {
	schema       *tfjson.SchemaAttribute
	hclAttribute *hcl.Attribute
}

// CompletionCandidates represents a list of candidates
// for completion loosely reflecting lsp.CompletionList
type CompletionCandidates interface {
	List() []CompletionCandidate
	Len() int
	IsComplete() bool
}

// CompletionCandidate represents a candidate for completion
// loosely reflecting lsp.CompletionItem
type CompletionCandidate interface {
	Label() string
	Detail() string
	Documentation() MarkupContent
	Snippet(pos hcl.Pos) (hcl.Pos, string)
}

// MarkupContent reflects lsp.MarkupContent
type MarkupContent interface {
	// TODO: eventually will need to propapate Kind here once the LSP
	// protocol types we use support it
	Value() string
}

// PlainText represents plain text markup content for the LSP.
type PlainText string

// Value returns the content itself for the LSP protocol.
func (m PlainText) Value() string {
	return string(m)
}

// Markdown represents markdown formatted markup content for the LSP.
type Markdown string

// Value returns the content itself for the LSP protocol.
func (m Markdown) Value() string {
	// This currently returns plaintext for Markdown, but should be changed once
	// the protocol types support markdown.
	return mdplain.Clean(string(m))
}
