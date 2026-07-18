/**
 * Property-based round-trip test for formatValue().
 *
 * Generates random objects (1-5 keys, string/number/nested-object values)
 * and asserts that formatValue() → parseFormattedValue() deep equals the original.
 *
 * **Validates: Requirements 10.4, 11.5**
 */

import { assertEquals } from "@std/assert";
import { formatValue } from "../../src/output/format-value.ts";

// --- Random helpers ---

function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function randomAlphaKey(): string {
  const len = randomInt(3, 8);
  let key = "";
  for (let i = 0; i < len; i++) {
    key += String.fromCharCode(97 + randomInt(0, 25)); // a-z
  }
  return key;
}

function randomSafeString(): string {
  // Start with a letter to ensure the string is never parseable as a number.
  // Use only alphanumeric + spaces (no commas, colons, or braces).
  const firstChars = "abcdefghijklmnopqrstuvwxyz";
  const restChars = "abcdefghijklmnopqrstuvwxyz0123456789 ";
  const len = randomInt(2, 12);
  let str = firstChars[randomInt(0, firstChars.length - 1)];
  for (let i = 1; i < len; i++) {
    str += restChars[randomInt(0, restChars.length - 1)];
  }
  return str.trimEnd() || "a"; // ensure non-empty after trim
}

function randomNumber(): number {
  // integers or simple decimals
  if (Math.random() < 0.5) {
    return randomInt(-999, 999);
  }
  return Math.round((Math.random() * 200 - 100) * 100) / 100;
}

function randomLeafValue(): string | number {
  return Math.random() < 0.5 ? randomSafeString() : randomNumber();
}

function randomNestedObject(): Record<string, string | number> {
  const numKeys = randomInt(1, 3);
  const obj: Record<string, string | number> = {};
  for (let i = 0; i < numKeys; i++) {
    obj[randomAlphaKey()] = randomLeafValue();
  }
  return obj;
}

/**
 * Generate a random object with 1-5 keys.
 * Values can be: string, number, or nested object (with string/number leaves).
 */
function generateRandomObject(): Record<string, unknown> {
  const numKeys = randomInt(1, 5);
  const obj: Record<string, unknown> = {};
  for (let i = 0; i < numKeys; i++) {
    const key = randomAlphaKey();
    const roll = Math.random();
    if (roll < 0.33) {
      obj[key] = randomSafeString();
    } else if (roll < 0.66) {
      obj[key] = randomNumber();
    } else {
      obj[key] = randomNestedObject();
    }
  }
  return obj;
}

// --- Test-only inverse parser ---

/**
 * Parse a formatted value string back into an object.
 * This is the inverse of formatValue() for object inputs.
 *
 * Strategy:
 * - Split on ", " to get key-value pairs (but we need to handle nested JSON)
 * - For each pair, split on ": " (first occurrence) to get key and value
 * - Try parsing value as number, then as JSON object, otherwise keep as string
 */
function parseFormattedValue(
  formatted: string,
): Record<string, unknown> {
  const result: Record<string, unknown> = {};

  // We need a smarter split that respects JSON braces
  const pairs = splitPairs(formatted);

  for (const pair of pairs) {
    const colonIdx = pair.indexOf(": ");
    if (colonIdx === -1) continue;

    const key = pair.substring(0, colonIdx);
    const rawValue = pair.substring(colonIdx + 2);

    result[key] = parseValue(rawValue);
  }

  return result;
}

/**
 * Split formatted string on ", " while respecting JSON brace depth.
 */
function splitPairs(formatted: string): string[] {
  const pairs: string[] = [];
  let current = "";
  let braceDepth = 0;

  for (let i = 0; i < formatted.length; i++) {
    const ch = formatted[i];

    if (ch === "{") {
      braceDepth++;
      current += ch;
    } else if (ch === "}") {
      braceDepth--;
      current += ch;
    } else if (
      ch === "," && formatted[i + 1] === " " && braceDepth === 0
    ) {
      pairs.push(current);
      current = "";
      i++; // skip the space
    } else {
      current += ch;
    }
  }

  if (current) {
    pairs.push(current);
  }

  return pairs;
}

/**
 * Parse a single value string back to its original type.
 */
function parseValue(raw: string): unknown {
  // Try as number first
  const num = Number(raw);
  if (!isNaN(num) && raw === String(num)) {
    return num;
  }

  // Try as JSON (for nested objects)
  if (raw.startsWith("{")) {
    try {
      return JSON.parse(raw);
    } catch {
      // fall through to string
    }
  }

  // Otherwise it's a string
  return raw;
}

// --- Property test ---

Deno.test("Property: formatValue round-trip for random objects (50+ iterations)", () => {
  const iterations = 60;

  for (let i = 0; i < iterations; i++) {
    const original = generateRandomObject();

    // formatValue produces the formatted string
    const formatted = formatValue(original as Record<string, unknown>);

    // parseFormattedValue reconstructs the object
    const reconstructed = parseFormattedValue(formatted);

    // The reconstructed object should deep-equal the original
    // Note: formatValue sorts keys, so we compare with sorted keys
    const sortedOriginal: Record<string, unknown> = {};
    for (const key of Object.keys(original).sort()) {
      sortedOriginal[key] = original[key];
    }

    assertEquals(
      reconstructed,
      sortedOriginal,
      `Round-trip failed on iteration ${i + 1}.\n` +
        `Original: ${JSON.stringify(original)}\n` +
        `Formatted: "${formatted}"\n` +
        `Reconstructed: ${JSON.stringify(reconstructed)}`,
    );
  }
});
