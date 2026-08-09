# Releasing DraftMeld

DraftMeld uses the version in `VERSION` as its single application version. A semantic tag starts publishing only when it points to a commit already on `main`.

## Version policy

Before `1.0.0`, DraftMeld uses minor versions for feature sets and possible breaking changes, patch versions for compatible fixes, and prerelease suffixes for release candidates:

- `0.4.0` — features or initial-development breaking changes;
- `0.4.1` — compatible fixes on the `0.4` line;
- `0.5.0-rc.1` — a release candidate for the next feature line.

Docker tags follow the same model:

- `edge` tracks accepted work on `dev` and may be unstable;
- `0.4.0` identifies one immutable release;
- `0.4` follows compatible patch releases on that minor line;
- `latest` identifies the newest stable tagged container release.

## Public outputs

The public release workflow publishes multi-architecture container images to GitHub Container Registry with build-provenance attestations. The production Dockerfile, Compose configuration, source archives, release notes, and complete application remain public under the AGPL.

The workflow also builds and smoke-tests native Windows and Linux packages as a release-integrity check. It does not upload those convenience packages to public Actions artifacts or GitHub Releases.

## Official supported builds

Signed Windows and Linux packages will be delivered separately when a third-party fulfillment provider and signing process are ready. They must be built from the corresponding public source, include required license/source notices, and use the exact public semantic version. See [community and official distribution](distribution.md).

Never add payment-provider secrets, signing keys, customer records, or entitlement lists to this repository or its public Actions logs.

## Release checklist

1. Merge feature pull requests into `dev`.
2. Open and merge a separate green promotion pull request from `dev` to `main`.
3. Set `VERSION` and every package version to the intended semantic version.
4. Change the matching changelog heading from `Unreleased` to the release date.
5. Run `npm run verify` and `npm run release:check`.
6. Optionally run the release workflow manually from `main` as a non-publishing dry run.
7. Create and push the exact tag from the current `main` commit.
8. Confirm container tags, provenance, source release notes, and the private official-build fulfillment job if applicable.

Never move or reuse a published release tag. Correct a failed published release with a new semantic version.

## Community installation

Copy `deployments/compose.yaml`, optionally copy `.env.example` to `.env`, and run:

```bash
docker compose up -d
```

Open `http://localhost:8080`. The image includes Poppler, Tesseract, and English recognition data for scanned-PDF OCR. To build locally from source, combine `compose.yaml` with `compose.build.yaml`.
