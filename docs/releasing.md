# Releasing DraftMeld

DraftMeld releases use the version in `VERSION` as the single application version. A semantic tag such as `v0.3.0` starts the release workflow only after the corresponding commit has reached `main`.

## Release outputs

Each tagged release publishes:

- a Windows x64 ZIP;
- Linux x64 and ARM64 tarballs;
- a SHA-256 checksum manifest;
- build-provenance attestations for the downloadable archives;
- versioned Linux x64 and ARM64 images in GitHub Container Registry;
- a GitHub prerelease populated from the matching changelog section.

Manual runs of the release workflow build, package, and smoke-test the native and container outputs without publishing a release or container image.

## Release checklist

1. Merge feature PRs into `dev`, then promote `dev` to `main` through a green PR.
2. Set `VERSION` and every package version to the intended semantic version.
3. Change the matching changelog heading from `Unreleased` to the release date.
4. Run `npm run verify` and `npm run release:check`.
5. Optionally run the release workflow manually from `main` as a non-publishing dry run.
6. Create and push the exact tag from the current `main` commit:

   ```bash
   git tag -a v0.3.0 -m "DraftMeld 0.3.0"
   git push origin v0.3.0
   ```

7. Confirm that the native artifacts, checksums, attestations, container image, and GitHub prerelease were published successfully.

Never move or reuse a published release tag. If a release fails after publication, correct the problem with a new semantic version.

## Installation

Extract the archive for the operating system, then run `draftmeld.exe` on Windows or `./draftmeld` on Linux. DraftMeld stores its SQLite data under `./data` unless `DRAFTMELD_DATA_DIR` is set. Selectable-text PDFs work with no additional software. Scanned-PDF OCR in a native installation requires Poppler's `pdftoppm` and Tesseract on `PATH`; see [local PDF OCR](ocr.md).

The container image is versioned in GitHub Container Registry:

```bash
docker run --rm -p 8080:8080 -v draftmeld-data:/data ghcr.io/coopa11y/draftmeld:0.3.0
```

Open `http://localhost:8080` after the application reports that it has started.

The container includes Poppler, Tesseract, and English recognition data, so scanned-PDF OCR works without installing host tools.
