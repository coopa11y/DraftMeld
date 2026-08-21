import type { Dispatch, SetStateAction } from "react";
import { importRankingPDF } from "../shared/api/rankings";

export async function runBusy(
  setBusy: Dispatch<SetStateAction<boolean>>,
  setError: Dispatch<SetStateAction<string>>,
  setMessage: Dispatch<SetStateAction<string>>,
  progressMessage: string,
  operation: () => Promise<void>,
  fallbackError: string,
  rethrow = false,
) {
  setBusy(true);
  setError("");
  setMessage(progressMessage);
  try {
    await operation();
  } catch (reason) {
    setError(errorMessage(reason, fallbackError));
    setMessage("");
    if (rethrow) throw reason;
  } finally {
    setBusy(false);
  }
}

export async function runSourceUpdate(
  setBusySourceId: Dispatch<SetStateAction<string>>,
  setError: Dispatch<SetStateAction<string>>,
  setMessage: Dispatch<SetStateAction<string>>,
  source: { id: string; name: string },
  operation: () => Promise<string>,
) {
  setBusySourceId(source.id);
  setError("");
  setMessage(`Updating ${source.name}.`);
  try {
    setMessage(await operation());
  } catch (reason) {
    setError(errorMessage(reason, `Unable to update ${source.name}.`));
    setMessage("");
  } finally {
    setBusySourceId("");
  }
}

export function errorMessage(reason: unknown, fallback: string) {
  return reason instanceof Error ? reason.message : fallback;
}

export function pdfImportMessage(result: Awaited<ReturnType<typeof importRankingPDF>>) {
  return `${result.source.name} imported: ${result.source.recordCount} players from ${result.pageCount} page${result.pageCount === 1 ? "" : "s"}.${
    result.ocrApplied ? " Scanned pages were recognized locally with OCR; review the imported rankings carefully." : ""
  }`;
}
