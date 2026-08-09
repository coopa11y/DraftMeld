import { FormField } from "./FormField";

export interface CsvColumnDefinition {
  key: string;
  label: string;
  required?: boolean;
  aliases?: string[];
}

interface CsvColumnMapperProps {
  busy: boolean;
  columns: CsvColumnDefinition[];
  headers: string[];
  mapping: Record<string, string>;
  optionalLabel: string;
  onChange: (mapping: Record<string, string>) => void;
}

export function CsvColumnMapper({ busy, columns, headers, mapping, optionalLabel, onChange }: CsvColumnMapperProps) {
  const required = columns.filter((column) => column.required);
  const optional = columns.filter((column) => !column.required);

  function control(column: CsvColumnDefinition) {
    return (
      <FormField key={column.key} label={column.label}>
        <select
          required={column.required}
          value={mapping[column.key] ?? ""}
          onChange={(event) => onChange({ ...mapping, [column.key]: event.target.value })}
          disabled={busy}
        >
          <option value="">{column.required ? "Choose a CSV column" : "Not included"}</option>
          {headers.map((header) => (
            <option key={header} value={header}>
              {header}
            </option>
          ))}
        </select>
      </FormField>
    );
  }

  return (
    <fieldset className="column-mapper">
      <legend>Match required columns</legend>
      {required.map(control)}
      {optional.length > 0 ? (
        <details>
          <summary>{optionalLabel}</summary>
          <div className="column-mapper-grid">{optional.map(control)}</div>
        </details>
      ) : null}
    </fieldset>
  );
}

export async function inspectCsvHeaders(file: File): Promise<string[]> {
  const firstLine = (await file.slice(0, 16_384).text()).split(/\r?\n/, 1)[0] ?? "";
  return parseCsvHeader(firstLine);
}

export function autoMapCsvHeaders(headers: string[], columns: CsvColumnDefinition[]): Record<string, string> {
  const normalized = new Map(headers.map((header) => [normalizeHeader(header), header]));
  const mapping: Record<string, string> = {};
  for (const column of columns) {
    const aliases = column.aliases ?? [column.key];
    const match = aliases.map((alias) => normalized.get(normalizeHeader(alias))).find(Boolean);
    if (match) mapping[column.key] = match;
  }
  return mapping;
}

function normalizeHeader(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "");
}

function parseCsvHeader(line: string): string[] {
  const fields: string[] = [];
  let current = "";
  let quoted = false;
  for (let index = 0; index < line.length; index++) {
    const character = line[index];
    if (character === '"' && quoted && line[index + 1] === '"') {
      current += '"';
      index++;
    } else if (character === '"') {
      quoted = !quoted;
    } else if (character === "," && !quoted) {
      fields.push(current.trim());
      current = "";
    } else {
      current += character;
    }
  }
  fields.push(current.trim());
  return fields.filter(Boolean);
}
