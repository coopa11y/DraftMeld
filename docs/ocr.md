# Local PDF OCR

DraftMeld reads selectable PDF text first. It invokes OCR only when a valid PDF contains no selectable text. OCR runs on the DraftMeld host; the application does not send the PDF or recognized text to a cloud OCR service.

The Docker image includes Poppler, Tesseract, and English recognition data. No extra setup is required when running DraftMeld with Docker.

Native Windows and Linux archives remain small, single-executable application packages. To enable scanned-PDF imports, install Poppler and Tesseract for the current user and make these commands available on `PATH`:

```text
pdftoppm
tesseract
```

On Debian or Ubuntu, the relevant package names are `poppler-utils` and `tesseract-ocr`. On Windows, use a trusted Poppler distribution and the Windows installation guidance linked by the Tesseract project. If a tool is installed outside `PATH`, set its full executable path:

```text
DRAFTMELD_PDFTOPPM_PATH=C:\path\to\pdftoppm.exe
DRAFTMELD_TESSERACT_PATH=C:\path\to\tesseract.exe
```

## Safety and limits

- Uploads must be valid PDFs between 1 byte and 20 MiB.
- Selectable-text extraction supports up to 200 pages; OCR supports up to 25 scanned pages.
- Pages are rendered at 180 DPI and recognized in numerical page order with English language data.
- An OCR attempt has a two-minute deadline, and the server processes one OCR document at a time.
- A rendered page is limited to 50 MiB, all rendered pages to 200 MiB, and recognized text to 10 MiB.
- DraftMeld creates a private temporary working directory and removes the uploaded copy and rendered page images after success or failure.
- OCR-derived imports display a review warning. Users should compare recognized rankings and rules with the original document before relying on them.

If the native application cannot find either command, the import returns an actionable error and normal selectable-text PDF imports continue to work.
