import type {
  SAILParseResult,
  DTCGGroup,
  SemanticMapping,
  TokenCategory,
} from '../types.js';

/**
 * SAILSemanticColor → primitive color alias mappings.
 * Each entry maps a SAIL enum value to a semantic token name and DTCG alias ref.
 */
const SEMANTIC_COLOR_MAP: Record<string, { name: string; aliasRef: string }> = {
  NEGATIVE: { name: 'semantic/destructive', aliasRef: '{color.red.500}' },
  POSITIVE: { name: 'semantic/positive', aliasRef: '{color.green.500}' },
  ACCENT: { name: 'semantic/accent', aliasRef: '{color.blue.500}' },
  SECONDARY: { name: 'semantic/secondary', aliasRef: '{color.gray.500}' },
  STANDARD: { name: 'semantic/standard', aliasRef: '{color.gray.900}' },
};

/**
 * Tailwind spacing unit → rem value conversion.
 */
const TAILWIND_UNIT_TO_REM: Record<string, string> = {
  '0': '0',
  '1': '0.25rem',
  '2': '0.5rem',
  '4': '1rem',
  '6': '1.5rem',
  '8': '2rem',
};

/**
 * SAILMarginSize enum value → semantic margin token name.
 */
const MARGIN_NAME_MAP: Record<string, string> = {
  NONE: 'none',
  EVEN_LESS: 'even-less',
  LESS: 'less',
  STANDARD: 'standard',
  MORE: 'more',
  EVEN_MORE: 'even-more',
};

/**
 * Build semantic token mappings from SAIL parse results.
 *
 * - SAILSemanticColor values become color alias tokens under `semantic/*`
 * - SAILMarginSize values become concrete dimension tokens under `margin/*`
 *
 * @param sail - Parsed SAIL type definitions
 * @param _primitiveTokens - The primitive token tree (unused but available for validation)
 * @returns Array of SemanticMapping objects to add to the token store
 */
export function buildSemanticTokens(
  sail: SAILParseResult,
  _primitiveTokens: DTCGGroup,
): SemanticMapping[] {
  const mappings: SemanticMapping[] = [];

  for (const enumDef of sail.enums) {
    if (enumDef.typeName === 'SAILSemanticColor') {
      for (const value of Object.keys(enumDef.values)) {
        const mapping = SEMANTIC_COLOR_MAP[value];
        if (mapping) {
          mappings.push({
            name: mapping.name,
            aliasRef: mapping.aliasRef,
            category: 'color' as TokenCategory,
            source: `SAILSemanticColor.${value}`,
          });
        }
      }
    }

    if (enumDef.typeName === 'SAILMarginSize') {
      for (const [value, tailwindUnit] of Object.entries(enumDef.values)) {
        const marginName = MARGIN_NAME_MAP[value];
        const remValue = TAILWIND_UNIT_TO_REM[tailwindUnit];
        if (marginName && remValue !== undefined) {
          mappings.push({
            name: `margin/${marginName}`,
            aliasRef: remValue,
            category: 'spacing' as TokenCategory,
            source: `SAILMarginSize.${value}`,
          });
        }
      }
    }
  }

  return mappings;
}
