package hcl

import (
	"regexp"

	hcllib "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-ls/internal/filesystem"
	"github.com/hashicorp/terraform-ls/internal/source"
)

type File interface {
	BlockAtPosition(filesystem.FilePosition) (*hcllib.Block, hcllib.Pos, error)
}

type file struct {
	filename string
	content  []byte
	f        *hcllib.File
}

func NewFile(f filesystem.File) File {
	return &file{
		filename: f.Filename(),
		content:  []byte(f.Text()),
	}
}

func (f *file) ast() (*hcllib.File, error) {
	if f.f != nil {
		return f.f, nil
	}

	hf, err := hclsyntax.ParseConfig(f.content, f.filename, hcllib.InitialPos)
	f.f = hf

	return f.f, err
}

func (f *file) BlockAtPosition(filePos filesystem.FilePosition) (*hcllib.Block, hcllib.Pos, error) {
	pos := filePos.Position()

	b, err := f.blockAtPosition(pos)
	if err != nil {
		return nil, pos, err
	}

	return b, pos, nil
}

func (f *file) blockAtPosition(pos hcllib.Pos) (*hcllib.Block, error) {
	ast, _ := f.ast()

	if body, ok := ast.Body.(*hclsyntax.Body); ok {
		if body.SrcRange.Empty() && pos != hcllib.InitialPos {
			return nil, &InvalidHclPosErr{pos, body.SrcRange}
		}
		if !body.SrcRange.Empty() {
			if posIsEqual(body.SrcRange.End, pos) {
				return nil, &NoBlockFoundErr{pos}
			}
			if !body.SrcRange.ContainsPos(pos) {
				return nil, &InvalidHclPosErr{pos, body.SrcRange}
			}
		}
	}

	block := ast.OutermostBlockAtPos(pos)
	if block == nil {
		return nil, &NoBlockFoundErr{pos}
	}

	return block, nil
}

func TokenAtPos(lines []source.Line, filePos filesystem.FilePosition) (string, error) {
	return tokenAtPos(lines, filePos.Position())
}

func tokenAtPos(lines []source.Line, pos hcllib.Pos) (string, error) {
	if len(lines) == 0 {
		if pos.Column != 1 || pos.Line != 1 {
			return "", &InvalidHclPosErr{pos, hcllib.Range{}}
		}
		return "", nil
	}
	r := regexp.MustCompile(`\b(\w+)$`)
	for i, srcLine := range lines {
		if i == pos.Line-1 {
			if srcLine.Range().End.Byte < pos.Byte || srcLine.Range().Start.Byte > pos.Byte {
				return "", &InvalidHclPosErr{pos, srcLine.Range()}
			}
			content := string(srcLine.Bytes())
			content = content[:pos.Byte-srcLine.Range().Start.Byte]

			groups := r.FindStringSubmatch(content)
			if len(groups) < 2 {
				return "", nil
			}
			return groups[1], nil
		}
	}
	return "", &InvalidHclPosErr{pos, hcllib.RangeBetween(lines[0].Range(), lines[len(lines)-1].Range())}
}

func posIsEqual(a, b hcllib.Pos) bool {
	return a.Byte == b.Byte &&
		a.Column == b.Column &&
		a.Line == b.Line
}
