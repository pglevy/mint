import fs from 'node:fs';
import type {
  CSSParserOptions,
  CSSParseResult,
  RawColorToken,
  RawTypographyToken,
  RawSpacingToken,
  AliasRecord,
} from '../types.js';

/**
 * Parse a Sailwind CSS file and extract color, typography, and spacing tokens
 * from the @theme block, plus font-family/font-weight from @layer base.
 */
export function parseCSSFile(options: CSSParserOptions): CSSParseResult {
  const { filePath } = options;

  // 1. Read the file — throw if not found
  let content: string;
  try {
    content = fs.readFileSync(filePath, 'utf-8');
  } catch {
    throw new Error(`Error: CSS file not found at ${filePath}`);
  }

  // 2. Extract the @theme block — throw if missing
  const themeBlock = extractThemeBlock(content);
  if (themeBlock === null) {
    throw new Error(`Error: No @theme block found in ${filePath}`);
  }

  // 3. Parse color tokens (concrete hex values)
  const colors = parseColorTokens(themeBlock);

  // 4. Parse alias tokens (var(--color-*) references) and resolve them
  const { resolvedColors, aliases } = parseAndResolveAliases(themeBlock, colors);

  // 5. Parse typography tokens from @theme (--text-*)
  const typography = parseTextTokens(themeBlock);

  // 6. Parse spacing and radius tokens from @theme
  const spacing = parseSpacingTokens(themeBlock);

  // 7. Extract font-family and font-weight from @layer base html rule
  const baseTypography = parseLayerBaseTypography(content);
  typography.push(...baseTypography);

  return {
    colors: [...colors, ...resolvedColors],
    typography,
    spacing,
    aliases,
  };
}


// ── Internal helpers ────────────────────────────────────────────────

/**
 * Extract the content inside the first @theme { ... } block.
 * Returns null if no @theme block is found.
 */
function extractThemeBlock(css: string): string | null {
  const themeStart = css.indexOf('@theme');
  if (themeStart === -1) return null;

  const braceStart = css.indexOf('{', themeStart);
  if (braceStart === -1) return null;

  // Walk forward counting braces to find the matching close
  let depth = 1;
  let i = braceStart + 1;
  while (i < css.length && depth > 0) {
    if (css[i] === '{') depth++;
    if (css[i] === '}') depth--;
    i++;
  }

  if (depth !== 0) return null;

  return css.slice(braceStart + 1, i - 1);
}

/**
 * Parse --color-{family}-{step}: #{hex}; lines into RawColorToken[].
 * Only matches concrete hex values (not var() references).
 */
function parseColorTokens(themeBlock: string): RawColorToken[] {
  const tokens: RawColorToken[] = [];
  const regex = /--color-([a-z]+)-(\d+)\s*:\s*(#[0-9A-Fa-f]{6})/g;
  let match: RegExpExecArray | null;

  while ((match = regex.exec(themeBlock)) !== null) {
    const family = match[1];
    const step = parseInt(match[2], 10);
    const hex = match[3];
    tokens.push({
      name: `${family}-${step}`,
      value: hex,
      family,
      step,
    });
  }

  return tokens;
}

/**
 * Parse --color-{name}: var(--color-{target}); alias lines.
 * Resolves each alias to the concrete hex value from the already-parsed color tokens.
 * Returns both the resolved color tokens and the alias records.
 */
function parseAndResolveAliases(
  themeBlock: string,
  concreteColors: RawColorToken[],
): { resolvedColors: RawColorToken[]; aliases: AliasRecord[] } {
  const resolvedColors: RawColorToken[] = [];
  const aliases: AliasRecord[] = [];

  // Match patterns like: --color-black: var(--color-gray-900)
  const regex = /--color-([a-z]+)\s*:\s*var\(--color-([a-z]+-\d+)\)/g;
  let match: RegExpExecArray | null;

  while ((match = regex.exec(themeBlock)) !== null) {
    const aliasName = match[1];
    const targetName = match[2];

    // Find the concrete color this alias points to
    const target = concreteColors.find((c) => c.name === targetName);
    const resolvedValue = target ? target.value : `var(--color-${targetName})`;

    resolvedColors.push({
      name: aliasName,
      value: resolvedValue,
      family: aliasName,
      step: 0,
    });

    aliases.push({
      tokenName: aliasName,
      aliasOf: targetName,
    });
  }

  return { resolvedColors, aliases };
}

/**
 * Parse --text-{name}: {value}; lines into RawTypographyToken[].
 */
function parseTextTokens(themeBlock: string): RawTypographyToken[] {
  const tokens: RawTypographyToken[] = [];
  const regex = /--text-([a-zA-Z0-9]+)\s*:\s*([^;]+)/g;
  let match: RegExpExecArray | null;

  while ((match = regex.exec(themeBlock)) !== null) {
    tokens.push({
      name: `text-${match[1]}`,
      value: match[2].trim(),
      dtcgType: 'dimension',
    });
  }

  return tokens;
}

/**
 * Parse --spacing-{name} and --radius-{name} lines into RawSpacingToken[].
 * Handles escaped dots in CSS names (e.g., --spacing-1\.5 → "spacing-1.5").
 */
function parseSpacingTokens(themeBlock: string): RawSpacingToken[] {
  const tokens: RawSpacingToken[] = [];

  // Match --spacing-{name}: {value}; (name may contain escaped dots)
  const spacingRegex = /--spacing-([a-zA-Z0-9\\.]+)\s*:\s*([^;]+)/g;
  let match: RegExpExecArray | null;

  while ((match = spacingRegex.exec(themeBlock)) !== null) {
    const rawName = match[1].replace(/\\/g, ''); // Remove CSS escape backslashes
    tokens.push({
      name: `spacing-${rawName}`,
      value: match[2].trim(),
    });
  }

  // Match --radius-{name}: {value};
  const radiusRegex = /--radius-([a-zA-Z0-9\\.]+)\s*:\s*([^;]+)/g;

  while ((match = radiusRegex.exec(themeBlock)) !== null) {
    const rawName = match[1].replace(/\\/g, '');
    tokens.push({
      name: `radius-${rawName}`,
      value: match[2].trim(),
    });
  }

  return tokens;
}

/**
 * Extract font-family and font-weight from the @layer base html { ... } rule.
 */
function parseLayerBaseTypography(css: string): RawTypographyToken[] {
  const tokens: RawTypographyToken[] = [];

  // Find @layer base block
  const layerMatch = css.match(/@layer\s+base\s*\{([\s\S]*?)\n\}/);
  if (!layerMatch) return tokens;

  const layerContent = layerMatch[1];

  // Find the html rule inside @layer base
  const htmlMatch = layerContent.match(/html\s*\{([\s\S]*?)\}/);
  if (!htmlMatch) return tokens;

  const htmlBlock = htmlMatch[1];

  // Extract font-family
  const fontFamilyMatch = htmlBlock.match(/font-family\s*:\s*([^;]+)/);
  if (fontFamilyMatch) {
    tokens.push({
      name: 'font-family-base',
      value: fontFamilyMatch[1].trim(),
      dtcgType: 'fontFamily',
    });
  }

  // Extract font-weight
  const fontWeightMatch = htmlBlock.match(/font-weight\s*:\s*([^;]+)/);
  if (fontWeightMatch) {
    tokens.push({
      name: 'font-weight-base',
      value: fontWeightMatch[1].trim(),
      dtcgType: 'fontWeight',
    });
  }

  return tokens;
}
