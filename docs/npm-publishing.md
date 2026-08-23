# Publishing the npm collections

The npm organization is `golangskills`. The public package scope is `@golangskills`; `golangskills.com` remains the project website and GitHub repository identity.

## Packages

| Collection | Package |
| --- | --- |
| Engineering Skills for Go | `@golangskills/engineering` |
| Distributed Systems Skills for Go | `@golangskills/distributed-systems` |
| Fintech Skills for Go | `@golangskills/fintech` |

All three packages use the same signed release version.

## One-time npm setup

1. Confirm that the `ashwingopalsamy` npm account has write access to the `golangskills` organization.
2. Enable two-factor authentication for publishing.
3. Keep the package names and public visibility unchanged after the first release. A published name and version cannot be reused.

The first publication creates the package records. npm staged publishing cannot create a brand-new package, so bootstrap with an interactive two-factor-authenticated pre-release. Do not put an npm token in this repository or in a chat message.

## Local preflight

From the repository root:

```sh
VERSION=0.3.0-rc.1
go run ./cmd/skillctl check
go run ./cmd/skillctl generate
go run ./cmd/skillctl npm package -version "$VERSION"
go run ./cmd/skillctl npm check -version "$VERSION"
```

Review each package without publishing:

```sh
for package_dir in dist/npm/*; do
  (cd "$package_dir" && npm pack --dry-run)
done
```

The staged packages are data-only. They have no lifecycle scripts, dependencies, or network-capable installer. npm installation does not modify an agent configuration.

## Bootstrap publish

After authenticating locally with npm, publish each staged package with public access. Use the exact version selected for the signed release:

```sh
VERSION=0.3.0-rc.1
go run ./cmd/skillctl npm package -version "$VERSION"
go run ./cmd/skillctl npm check -version "$VERSION"
for package_dir in dist/npm/*; do
  (cd "$package_dir" && npm publish --access public --tag bootstrap)
done
```

The `bootstrap` dist-tag keeps the pre-release away from the normal `latest` channel. After the package records exist, configure Trusted Publishing for each package, regenerate the stable `0.3.0` staging, and push the signed `v0.3.0` tag. The GitHub workflow then publishes the stable version through OIDC.

## Trusted publishing for later releases

After each package exists, configure npm Trusted Publishing with these exact values:

- Provider: GitHub Actions
- Organization or user: `ashwingopalsamy`
- Repository: `golangskills.com`
- Workflow filename: `publish-npm.yml`
- Environment: `npm-publish`
- Allowed action: `npm publish`

The workflow uses GitHub OIDC and publishes all three packages from a signed `v*` tag. It grants `id-token: write`, uses a GitHub-hosted runner, and upgrades npm to the version required by trusted publishing. No long-lived npm publish token is required.

## Discovery and installation

The npm packages are versioned artifacts. The open Skills CLI continues to install from the canonical GitHub source:

```sh
npx skills add ashwingopalsamy/golangskills.com
```

See [docs/install.md](install.md) for the generated Codex, Claude Code, Cursor, and OpenCode layouts.

## References

- [npm organization-scoped packages](https://docs.npmjs.com/creating-and-publishing-an-organization-scoped-package/)
- [npm package file selection](https://docs.npmjs.com/cli/commands/npm-publish/)
- [npm Trusted Publishing](https://docs.npmjs.com/trusted-publishers/)
- [npm provenance](https://docs.npmjs.com/generating-provenance-statements/)
