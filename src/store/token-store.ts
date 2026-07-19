import type { DTCGGroup, DTCGToken, TokenCategory } from "../types.ts";

/** A single flattened token entry with its resolved path and value. */
export interface TokenEntry {
  path: string;
  value: string | number | Record<string, unknown> | unknown[];
  type: string;
  category: TokenCategory;
}

/** Result returned from `loadFromJSON` with per-category token counts. */
export interface LoadResult {
  counts: Record<TokenCategory, number>;
}

function isDTCGToken(node: unknown): node is DTCGToken {
  return (
    typeof node === "object" &&
    node !== null &&
    "$value" in node &&
    "$type" in node
  );
}

function collectTokens(
  node: DTCGGroup,
  prefix: string,
  category: TokenCategory,
  result: TokenEntry[],
): void {
  for (const [key, child] of Object.entries(node)) {
    const path = prefix ? `${prefix}.${key}` : key;
    if (isDTCGToken(child)) {
      result.push({ path, value: child.$value, type: child.$type, category });
    } else {
      collectTokens(child as DTCGGroup, path, category, result);
    }
  }
}

function countTokens(node: DTCGGroup): number {
  let count = 0;
  for (const child of Object.values(node)) {
    count += isDTCGToken(child) ? 1 : countTokens(child as DTCGGroup);
  }
  return count;
}

function cloneGroup(group: DTCGGroup): DTCGGroup {
  return JSON.parse(JSON.stringify(group));
}

export class TokenStore {
  private tree: Record<TokenCategory, DTCGGroup> = {
    color: {},
    gradient: {},
    typography: {},
    spacing: {},
  };

  /**
   * Load token data from parsed JSON. Returns counts per category
   * instead of logging to console.
   */
  loadFromJSON(json: Record<string, DTCGGroup>): LoadResult {
    this.tree = { color: {}, gradient: {}, typography: {}, spacing: {} };
    for (const cat of ["color", "gradient", "typography", "spacing"] as TokenCategory[]) {
      if (json[cat]) this.tree[cat] = json[cat];
    }
    return {
      counts: {
        color: countTokens(this.tree.color),
        gradient: countTokens(this.tree.gradient),
        typography: countTokens(this.tree.typography),
        spacing: countTokens(this.tree.spacing),
      },
    };
  }

  getAllTokens(): DTCGGroup {
    return cloneGroup(this.tree as unknown as DTCGGroup);
  }

  getByCategory(category: TokenCategory): DTCGGroup {
    return cloneGroup(this.tree[category]);
  }

  getByFamily(category: TokenCategory, family: string): DTCGGroup {
    const catTree = this.tree[category];
    if (!(family in catTree)) return {};
    if (isDTCGToken(catTree[family])) {
      return cloneGroup({ [family]: catTree[family] } as DTCGGroup);
    }
    return cloneGroup(catTree[family] as DTCGGroup);
  }

  /** Returns the top-level family keys for a given category. */
  getFamilies(category: TokenCategory): string[] {
    return Object.keys(this.tree[category]);
  }

  listTokens(category?: TokenCategory): TokenEntry[] {
    const result: TokenEntry[] = [];
    const cats: TokenCategory[] = category
      ? [category]
      : ["color", "gradient", "typography", "spacing"];
    for (const cat of cats) {
      collectTokens(this.tree[cat], cat, cat, result);
    }
    return result;
  }

  toJSON(category?: TokenCategory): string {
    if (category) {
      return JSON.stringify({ [category]: this.tree[category] }, null, 2);
    }
    return JSON.stringify(this.tree, null, 2);
  }
}
