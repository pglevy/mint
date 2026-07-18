/**
 * Formats a token $value into a human-readable string.
 *
 * - Strings pass through unchanged
 * - Numbers are converted to their string representation
 * - Objects produce sorted key: value pairs joined by ", "
 *   (nested objects recurse one level, then fall back to JSON.stringify)
 */
export function formatValue(
  value: string | number | Record<string, unknown>,
): string {
  if (typeof value === "string") {
    return value;
  }

  if (typeof value === "number") {
    return String(value);
  }

  // Special case: object with exactly "value" and "unit" keys → combine as "0.5rem"
  const keys = Object.keys(value);
  if (
    keys.length === 2 &&
    "value" in value &&
    "unit" in value
  ) {
    return `${value["value"]}${value["unit"]}`;
  }

  // Object: sort keys alphabetically and format as "key: value" pairs
  keys.sort();
  const parts = keys.map((key) => {
    const v = value[key];
    const formatted = formatNestedValue(v);
    return `${key}: ${formatted}`;
  });

  return parts.join(", ");
}

/**
 * Format a nested value (one level deep inside an object).
 * Primitives render directly; objects use JSON.stringify.
 */
function formatNestedValue(v: unknown): string {
  if (typeof v === "string") {
    return v;
  }
  if (typeof v === "number") {
    return String(v);
  }
  if (typeof v === "boolean") {
    return String(v);
  }
  if (v === null || v === undefined) {
    return String(v);
  }
  // For nested objects/arrays, fall back to JSON
  return JSON.stringify(v);
}
