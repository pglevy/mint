import type { TokenCategory, DTCGToken, DTCGGroup } from '../types.js';

export const TOKENS_URL =
  'https://raw.githubusercontent.com/pglevy/sailwind/main/public/tokens.json';

function isDTCGToken(node: unknown): node is DTCGToken {
  return (
    typeof node === 'object' &&
    node !== null &&
    '$value' in node &&
    '$type' in node
  );
}

function collectTokens(
  node: DTCGGroup,
  prefix: string,
  category: TokenCategory,
  result: Array<{ path: string; value: string; type: string; category: TokenCategory }>,
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
    typography: {},
    spacing: {},
  };

  loadFromJSON(json: Record<string, DTCGGroup>): void {
    this.tree = { color: {}, typography: {}, spacing: {} };
    for (const cat of ['color', 'typography', 'spacing'] as TokenCategory[]) {
      if (json[cat]) this.tree[cat] = json[cat];
    }
    const c = countTokens(this.tree.color);
    const t = countTokens(this.tree.typography);
    const s = countTokens(this.tree.spacing);
    console.log(`Loaded ${c} color, ${t} typography, ${s} spacing tokens`);
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
    if (isDTCGToken(catTree[family])) return cloneGroup({ [family]: catTree[family] } as DTCGGroup);
    return cloneGroup(catTree[family] as DTCGGroup);
  }

  listTokens(category?: TokenCategory): Array<{ path: string; value: string; type: string; category: TokenCategory }> {
    const result: Array<{ path: string; value: string; type: string; category: TokenCategory }> = [];
    const cats: TokenCategory[] = category ? [category] : ['color', 'typography', 'spacing'];
    for (const cat of cats) {
      collectTokens(this.tree[cat], cat, cat, result);
    }
    return result;
  }

  toJSON(category?: TokenCategory): string {
    if (category) return JSON.stringify({ [category]: this.tree[category] }, null, 2);
    return JSON.stringify(this.tree, null, 2);
  }
}
