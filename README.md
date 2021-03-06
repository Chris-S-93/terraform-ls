# Terraform Language Server

Experimental version of [Terraform](https://www.terraform.io) language server.

## What is LSP

Read more about the Language Server Protocol at https://microsoft.github.io/language-server-protocol/

## Current Status

Not all language features (from LSP's or any other perspective) are available
at the time of writing, but this is an active project with the aim of delivering
smaller, incremental updates over time.

We encourage you to [browse existing issues](https://github.com/hashicorp/terraform-ls/issues)
and/or [open new issue](https://github.com/hashicorp/terraform-ls/issues/new/choose)
if you experience a bug or have an idea for a feature.

## Installation

1. [Download for the latest version](https://github.com/hashicorp/terraform-ls/releases/latest)
  of the language server relevant for your operating system and architecture.
2. The language server is distributed as a single binary.
  Install it by unzipping it and moving it to a directory
  included in your system's `PATH`.
3. You can verify integrity by comparing the SHA256 checksums
  which are part of the release (called `terraform-ls_<VERSION>_SHA256SUMS`).
4. Check that you have installed the server correctly via `terraform-ls -v`.
  You should see the latest version printed to your terminal.

## Usage

The most reasonable way you will interact with the language server
is through a client represented by an IDE, or a plugin of an IDE.

Please follow the [relevant guide for your IDE](./docs/USAGE.md).

## Credits

- [Martin Atkins](https://github.com/apparentlymart) - particularly the virtual filesystem
- [Zhe Cheng](https://github.com/njuCZ) - research, design, prototyping assistance
- [Julio Sueiras](https://github.com/juliosueiras) - particularly his [language server implementation](https://github.com/juliosueiras/terraform-lsp)
 

## `terraform-ls` VS `terraform-lsp`

Both HashiCorp and [the maintainer](https://github.com/juliosueiras) of [`terraform-lsp`](https://github.com/juliosueiras/terraform-lsp)
expressed interest in collaborating on a language server and are working
towards a _long-term_ goal of a single stable and feature-complete implementation.

For the time being both projects continue to exist, giving users the choice:

- `terraform-ls` providing
  - overall stability (by relying only on public APIs)
  - compatibility with any provider and any Terraform `>=0.12.0`
  - currently less features
    - due to project being younger and relying on public APIs which may not
      offer the same functionality yet

- `terraform-lsp` providing
  - currently more features
  - compatibility with a single particular Terraform (`0.12.20` at time of writing)
    - configs designed for other `0.12` versions may work, but interpretation may be inaccurate
  - less stability (due to reliance on Terraform's own internal packages)
