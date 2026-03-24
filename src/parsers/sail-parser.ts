import fs from 'node:fs';
import type {
  SAILParserOptions,
  SAILParseResult,
  SAILEnumMapping,
} from '../types.js';

/**
 * Tailwind unit mapping for SAILMarginSize values.
 * Each SAIL margin enum value maps to a Tailwind spacing unit number.
 */
const MARGIN_SIZE_MAP: Record<string, string> = {
  NONE: '0',
  EVEN_LESS: '1',
  LESS: '2',
  STANDARD: '4',
  MORE: '6',
  EVEN_MORE: '8',
};

/**
 * Parse a Sailwind SAIL types file and extract string literal union types
 * that start with "SAIL". Each type becomes a SAILEnumMapping with its
 * type name and a record of values.
 *
 * For SAILMarginSize, values are mapped to Tailwind unit numbers.
 * For all other types, values map to themselves.
 */
export function parseSAILFile(options: SAILParserOptions): SAILParseResult {
  const { filePath } = options;

  let content: string;
  try {
    content = fs.readFileSync(filePath, 'utf-8');
  } catch {
    console.warn(`Warning: SAIL types file not found at ${filePath}, continuing with CSS-only tokens`);
    return { enums: [] };
  }

  const enums = extractSAILTypes(content);
  return { enums };
}

// ── Internal helpers ────────────────────────────────────────────────

/**
 * Extract all `export type SAIL* = "VALUE1" | "VALUE2" | ...` patterns
 * from the file content and return them as SAILEnumMapping[].
 */
function extractSAILTypes(content: string): SAILEnumMapping[] {
  const mappings: SAILEnumMapping[] = [];

  // Match: export type SAILSomeName = "VAL1" | "VAL2" | "VAL3"
  // The type name must start with SAIL. Values are quoted string literals separated by |.
  const typeRegex = /export\s+type\s+(SAIL\w+)\s*=\s*("[\w]+"(?:\s*\|\s*"[\w]+")*)/g;
  let match: RegExpExecArray | null;

  while ((match = typeRegex.exec(content)) !== null) {
    const typeName = match[1];
    const valuesStr = match[2];

    // Extract individual string literal values from the union
    const valueRegex = /"(\w+)"/g;
    const values: Record<string, string> = {};
    let valueMatch: RegExpExecArray | null;

    while ((valueMatch = valueRegex.exec(valuesStr)) !== null) {
      const rawValue = valueMatch[1];

      if (typeName === 'SAILMarginSize') {
        // Map to Tailwind unit numbers
        values[rawValue] = MARGIN_SIZE_MAP[rawValue] ?? rawValue;
      } else {
        // Store raw value as-is
        values[rawValue] = rawValue;
      }
    }

    mappings.push({ typeName, values });
  }

  return mappings;
}
