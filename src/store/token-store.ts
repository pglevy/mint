import type {
  TokenCategory,
  DTCGToken,
  DTCGGroup,
  CSSParseResult,
  SAILParseResult,
  SemanticMapping,
} from '../types.js';
import { buildSemanticTokens } from './semantic-mapper.js';

/**
 * Check whether a value is a DTCG leaf token (has $value).
 */
function isDTCGToken(node: unknown): node is DTCGToken {
  return (
    typeof node === 'object' &&
    node !== null &&
    '$value' in node &&
    '$type' in node
  );
}

/**
 * Set a value at a nested path in a DTCGGroup tree.
 * Creates intermediate group objects as needed.
 */
function setNestedToken(root: DTCGGroup, path: string[], token: DTCGToken): void {
  let current: DTCGGroup = root;
  for (let i = 0; i < path.length - 1; i++) {
    const segment = path[i];
    if (!(segment in current) || isDTCGToken(current[segment])) {
      current[segment] = {};
    }
    current = current[segment] as DTCGGroup;
  }
  current[path[path.length - 1]] = token;
}

/**
 * Recursively collect all leaf tokens from a DTCGGroup tree,
 * building dot-notation paths as it walks.
 */
function collectTokens(
  node: DTCGGroup,
  prefix: string,
  category: TokenCategory,
  result: Array<{ path: string; value: string; type: string; category: TokenCategory }>,
): void {
  for (const [key, child] of Object.entries(node)) {
    const path = prefix ? `${prefix}.${key}` : key;
    if (isDTCGToken(child)) {
      result.push({
        path,
        value: child.$value,
        type: child.$type,
        category,
      });
    } else {
      collectTokens(child as DTCGGroup, path, category, result);
    }
  }
}

/**
 * Count leaf tokens in a DTCGGroup tree.
 */
function countTokens(node: DTCGGroup): number {
  let count = 0;
  for (const child of Object.values(node)) {
    if (isDTCGToken(child)) {
      count++;
    } else {
      count += countTokens(child as DTCGGroup);
    }
  }
  return count;
}

/**
 * Deep-clone a DTCGGroup tree (simple JSON round-trip).
 */
function cloneGroup(group: DTCGGroup): DTCGGroup {
  return JSON.parse(JSON.stringify(group));
}

/**
 * In-memory token store that organises tokens into a W3C DTCG-compatible
 * nested tree and provides query, filter, and serialisation methods.
 */
export class TokenStore {
  private tree: Record<TokenCategory, DTCGGroup> = {
    color: {},
    typography: {},
    spacing: {},
  };

  /**
   * Build the DTCG token tree from CSS and SAIL parse results.
   */
  load(css: CSSParseResult, sail: SAILParseResult): void {
    // Reset tree
    this.tree = { color: {}, typography: {}, spacing: {} };

    // ── Colors ────────────────────────────────────────────────
    // Build a lookup of alias names → target names for description
    const aliasMap = new Map(css.aliases.map((a) => [a.tokenName, a.aliasOf]));

    for (const color of css.colors) {
      const aliasTarget = aliasMap.get(color.name);

      if (aliasTarget) {
        // Alias token (step === 0, e.g. "black" → "gray-900")
        // Place directly under color/{name} with DTCG alias syntax
        const [targetFamily, targetStep] = splitLastDash(aliasTarget);
        setNestedToken(this.tree.color, [color.name], {
          $value: `{color.${targetFamily}.${targetStep}}`,
          $type: 'color',
          $description: `Alias of ${aliasTarget}`,
        });
      } else {
        // Concrete color token → color/{family}/{step}
        setNestedToken(this.tree.color, [color.family, String(color.step)], {
          $value: color.value,
          $type: 'color',
          $description: `${capitalize(color.family)} ${color.step}`,
        });
      }
    }

    // ── Typography ────────────────────────────────────────────
    for (const typo of css.typography) {
      const { group, name } = parseTypographyName(typo.name);
      setNestedToken(this.tree.typography, [group, name], {
        $value: typo.value,
        $type: typo.dtcgType,
        $description: `${typo.name}`,
      });
    }

    // ── Spacing ───────────────────────────────────────────────
    for (const sp of css.spacing) {
      const { group, name } = parseSpacingName(sp.name);
      setNestedToken(this.tree.spacing, [group, name], {
        $value: sp.value,
        $type: 'dimension',
        $description: `${sp.name}`,
      });
    }

    // ── Semantic tokens ───────────────────────────────────────
    const semanticMappings = buildSemanticTokens(sail, this.tree as unknown as DTCGGroup);
    this.addSemanticTokens(semanticMappings);

    // ── Report counts ─────────────────────────────────────────
    const colorCount = countTokens(this.tree.color);
    const typoCount = countTokens(this.tree.typography);
    const spacingCount = countTokens(this.tree.spacing);
    const semanticCount = semanticMappings.length;
    console.log(`Loaded ${colorCount} color, ${typoCount} typography, ${spacingCount} spacing tokens (${semanticCount} semantic)`);
  }

  /**
   * Add semantic tokens to the tree. Called by the semantic mapper
   * after load() to wire in alias tokens (task 6.1).
   */
  addSemanticTokens(mappings: SemanticMapping[]): void {
    for (const mapping of mappings) {
      const pathSegments = mapping.name.split('/');
      const category = mapping.category;
      setNestedToken(this.tree[category], pathSegments, {
        $value: mapping.aliasRef,
        $type: category === 'color' ? 'color' : 'dimension',
        $description: mapping.source,
      });
    }
  }

  /**
   * Return the full token tree across all categories.
   */
  getAllTokens(): DTCGGroup {
    return cloneGroup(this.tree as unknown as DTCGGroup);
  }

  /**
   * Return the token sub-tree for a single category.
   */
  getByCategory(category: TokenCategory): DTCGGroup {
    return cloneGroup(this.tree[category]);
  }

  /**
   * Return the token sub-tree for a specific family within a category.
   * For color, family is e.g. "red"; for typography, "font-family"; etc.
   */
  getByFamily(category: TokenCategory, family: string): DTCGGroup {
    const catTree = this.tree[category];
    if (family in catTree && !isDTCGToken(catTree[family])) {
      return cloneGroup(catTree[family] as DTCGGroup);
    }
    // If the family is a leaf token itself, wrap it
    if (family in catTree && isDTCGToken(catTree[family])) {
      return cloneGroup({ [family]: catTree[family] } as DTCGGroup);
    }
    return {};
  }

  /**
   * Return a flat array of all tokens, optionally filtered by category.
   * Each entry has dot-notation path, value, type, and category.
   */
  listTokens(category?: TokenCategory): Array<{
    path: string;
    value: string;
    type: string;
    category: TokenCategory;
  }> {
    const result: Array<{ path: string; value: string; type: string; category: TokenCategory }> = [];
    const categories: TokenCategory[] = category ? [category] : ['color', 'typography', 'spacing'];

    for (const cat of categories) {
      collectTokens(this.tree[cat], cat, cat, result);
    }

    return result;
  }

  /**
   * Serialise the token tree (or a single category) as a JSON string.
   */
  toJSON(category?: TokenCategory): string {
    if (category) {
      return JSON.stringify({ [category]: this.tree[category] }, null, 2);
    }
    return JSON.stringify(this.tree, null, 2);
  }
}

// ── Name-parsing helpers ──────────────────────────────────────────

/**
 * Split a dashed name at the last dash: "gray-900" → ["gray", "900"]
 */
function splitLastDash(name: string): [string, string] {
  const idx = name.lastIndexOf('-');
  if (idx === -1) return [name, name];
  return [name.slice(0, idx), name.slice(idx + 1)];
}

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

/**
 * Parse a typography token name into group + leaf name.
 *   "font-family-base" → { group: "font-family", name: "base" }
 *   "font-weight-base" → { group: "font-weight", name: "base" }
 *   "text-3xl"         → { group: "text-size",   name: "3xl" }
 */
function parseTypographyName(raw: string): { group: string; name: string } {
  if (raw.startsWith('font-family-')) {
    return { group: 'font-family', name: raw.slice('font-family-'.length) };
  }
  if (raw.startsWith('font-weight-')) {
    return { group: 'font-weight', name: raw.slice('font-weight-'.length) };
  }
  if (raw.startsWith('text-')) {
    return { group: 'text-size', name: raw.slice('text-'.length) };
  }
  // Fallback: use the whole name as both group and leaf
  return { group: raw, name: raw };
}

/**
 * Parse a spacing token name into group + leaf name.
 *   "spacing-1.5" → { group: "spacing", name: "1.5" }
 *   "radius-sm"   → { group: "radius",  name: "sm" }
 */
function parseSpacingName(raw: string): { group: string; name: string } {
  if (raw.startsWith('spacing-')) {
    return { group: 'spacing', name: raw.slice('spacing-'.length) };
  }
  if (raw.startsWith('radius-')) {
    return { group: 'radius', name: raw.slice('radius-'.length) };
  }
  // Fallback
  return { group: raw, name: raw };
}
