---
id: task-2gk
title: Add Nix flake-based installation option
status: open
priority: low
type: feature
created_at: 2026-05-30T18:38:57Z
updated_at: 2026-05-30T18:39:02Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - docs
  - promotion
references:
  - README.md
  - Makefile
---

## Description

Add a Nix flake for tl so NixOS/nix users can install with 'nix run github:aholbreich/tl' or add it to their system packages.

Implementation:
1. Create flake.nix in the repo root with:
   - A Go build derivation using buildGoModule (the standard approach for Go projects in nixpkgs).
   - Vendor hash or vendor directory support (check if Go module vendoring is compatible).
   - Outputs: default package (the tl binary), devShell for development (with Go toolchain).

2. Add a 'nix build' target to Makefile (wraps 'nix build .') for CI convenience.

3. Add a 'nix fmt' / 'nix flake check' to CI to keep the flake valid.

4. Update README.md Installation Options section with:
   sh
   nix run github:aholbreich/tl         # run directly
   nix profile install github:aholbreich/tl  # install to user profile
   nix
   inputs.tl.url = "github:aholbreich/tl";
   # then in environment.systemPackages: [ inputs.tl.packages.${system}.default ]
   

Implementation notes:
- Use buildGoModule with vendorHash =  and a vendor directory committed to the repo (or use proxy vendor hash).
- Check Go version compatibility with the Nixpkgs Go toolchain (currently Go 1.22+).
- Consider using flake-parts or keeping it simple with a standalone flake.
- Add a 'nix flake check' step to CI to validate the flake doesn't break.
- The flake should also expose a devShell for tl contributors: 'nix develop' drops into a shell with Go, gotools, etc.

Not in scope:
- NixOS module (not needed — tl is a CLI tool, not a system service)
- Home Manager module (optional future enhancement)
- nixpkgs PR upstreaming (separate effort)

## Notes

- 2026-05-30T18:39:02Z [pi:planning] note: Full Nix task details: Implementation: 1. Create flake.nix with buildGoModule derivation. Outputs: default package (tl binary), devShell (Go toolchain). 2. Add 'nix build' target to Makefile. 3. Add 'nix flake check' to CI. 4. Update README with Nix install section (nix run, nix profile install, system flake input). Not in scope: NixOS module, Home Manager module, nixpkgs PR upstreaming.
