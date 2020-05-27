package lsp

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform-ls/internal/terraform/lang"
	lsp "github.com/sourcegraph/go-lsp"
	"strings"
)

func CompletionList(candidates lang.CompletionCandidates, pos hcl.Pos, caps lsp.TextDocumentClientCapabilities, prefix string) lsp.CompletionList {
	snippetSupport := caps.Completion.CompletionItem.SnippetSupport
	list := lsp.CompletionList{}

	if candidates == nil {
		return list
	}

	cList := candidates.List()

	list.IsIncomplete = !candidates.IsComplete()
	list.Items = make([]lsp.CompletionItem, 0)
	for _, c := range cList {
		item := CompletionItem(c, pos, snippetSupport, prefix)
		if item != nil {
			list.Items = append(list.Items, *item)
		}
	}

	return list
}

func CompletionItem(candidate lang.CompletionCandidate, pos hcl.Pos, snippetSupport bool, prefix string) *lsp.CompletionItem {
	// TODO: deprecated / tags?

	doc := ""
	if c := candidate.Documentation(); c != nil {
		// TODO: markdown handling
		doc = c.Value()
	}

	if snippetSupport {
		pos, newText := candidate.Snippet(pos)

		if !strings.HasPrefix(newText, prefix) {
			return nil
		}
		return &lsp.CompletionItem{
			Label:            candidate.Label(),
			Kind:             lsp.CIKField,
			InsertTextFormat: lsp.ITFSnippet,
			Detail:           candidate.Detail(),
			Documentation:    doc,
			TextEdit: &lsp.TextEdit{
				Range: lsp.Range{
					Start: lsp.Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   lsp.Position{Line: pos.Line - 1, Character: pos.Column - 1},
				},
				NewText: newText[len(prefix):],
			},
		}
	}

	if !strings.HasPrefix(candidate.Label(), prefix) {
		return nil
	}
	return &lsp.CompletionItem{
		Label:            candidate.Label(),
		Kind:             lsp.CIKField,
		InsertTextFormat: lsp.ITFPlainText,
		Detail:           candidate.Detail(),
		Documentation:    doc,
		TextEdit: &lsp.TextEdit{
			Range: lsp.Range{
				Start: lsp.Position{Line: pos.Line - 1, Character: pos.Column - 1},
				End:   lsp.Position{Line: pos.Line - 1, Character: pos.Column - 1},
			},
			NewText: candidate.Label()[len(prefix):],
		},
	}
}
