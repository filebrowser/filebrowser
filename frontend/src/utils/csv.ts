export interface CsvData {
  headers: string[];
  rows: string[][];
}

/**
 * Parse CSV content into headers and rows
 * Supports quoted fields and handles commas within quotes
 */
export function parseCSV(
  content: string,
  columnSeparator: Array<string>
): CsvData {
  if (!content || content.trim().length === 0) {
    return { headers: [], rows: [] };
  }

  const lines = content.split(/\r?\n/);
  const result: string[][] = [];

  for (const line of lines) {
    if (line.trim().length === 0) continue;

    const row: string[] = [];
    let currentField = "";
    let inQuotes = false;

    for (let i = 0; i < line.length; i++) {
      const char = line[i];
      const nextChar = line[i + 1];

      if (char === '"') {
        if (inQuotes && nextChar === '"') {
          // Escaped quote
          currentField += '"';
          i++; // Skip next quote
        } else {
          // Toggle quote state
          inQuotes = !inQuotes;
        }
      } else if (columnSeparator.includes(char) && !inQuotes) {
        // Field separator
        row.push(currentField);
        currentField = "";
      } else {
        currentField += char;
      }
    }

    // Add the last field
    row.push(currentField);
    result.push(row);
  }

  if (result.length === 0) {
    return { headers: [], rows: [] };
  }

  // First row is headers
  const headers = result[0];
  const rows = result.slice(1);

  return { headers, rows };
}
