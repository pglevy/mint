import { describe, it, expect, vi } from 'vitest';
import { TokenStore } from '../../store/token-store.js';
import type { CSSParseResult, SAILParseResult } from '../../types.js';

// ── Test fixtures ───────────────────────────────────────────────────

const emptySail: SAILParseResult = { enums: [] };

function makeCSSResult(overrides?: Partial<CSSParseResult>): CSSParseResult {
  return {
    colors: [
      { name: 'red-50', value: '#FDEDF0', family: 'red', step: 50 },
      { name: 'red-500', value: '#DE0037', family: 'red', step: 500 },
      { name: 'red-900', value: '#9F0019', family: 'red', step: 900 },
      { name: 'gray-500', value: '#6B7280', family: 'gray', step: 500 },
      { name: 'gray-900', value: '#222222', family: 'gray', step: 900 },
      { name: 'black', value: '#222222', family: 'black', step: 0 },
    ],
    typography: [
      { name: 'font-family-base', value: "'Open Sans', sans-serif", dtcgType: 'fontFamily' },
      { name: 'font-weight-base', value: '400', dtcgType: 'fontWeight' },
      { name: 'text-3xl', value: '1.75rem', dtcgType: 'dimension' },
    ],
    spacing: [
      { name: 'spacing-1.5', value: '0.375rem' },
      { name: 'spacing-2.5', value: '0.625rem' },
      { name: 'radius-sm', value: '0.25rem' },
    ],
    aliases: [{ tokenName: 'black', aliasOf: 'gray-900' }],
    ...overrides,
  };
}

function loadedStore(css?: CSSParseResult): TokenStore {
  const store = new TokenStore();
  store.load(css ?? makeCSSResult(), emptySail);
  return store;
}

// ── Tests ───────────────────────────────────────────────────────────

describe('TokenStore', () => {
  describe('load()', () => {
    it('places color tokens under color/{family}/{step}', () => {
      const store = loadedStore();
      const tree = store.getAllTokens();
      const red500 = (tree as any).color.red['500'];
      expect(red500).toEqual({
        $value: '#DE0037',
        $type: 'color',
        $description: 'Red 500',
      });
    });

    it('places alias color tokens with DTCG alias syntax', () => {
      const store = loadedStore();
      const tree = store.getAllTokens();
      const black = (tree as any).color.black;
      expect(black.$value).toBe('{color.gray.900}');
      expect(black.$type).toBe('color');
      expect(black.$description).toBe('Alias of gray-900');
    });

    it('places typography tokens under typography/{group}/{name}', () => {
      const store = loadedStore();
      const tree = store.getAllTokens();
      const fontFamily = (tree as any).typography['font-family'].base;
      expect(fontFamily.$value).toBe("'Open Sans', sans-serif");
      expect(fontFamily.$type).toBe('fontFamily');

      const fontWeight = (tree as any).typography['font-weight'].base;
      expect(fontWeight.$value).toBe('400');
      expect(fontWeight.$type).toBe('fontWeight');

      const textSize = (tree as any).typography['text-size']['3xl'];
      expect(textSize.$value).toBe('1.75rem');
      expect(textSize.$type).toBe('dimension');
    });

    it('places spacing tokens under spacing/{group}/{name}', () => {
      const store = loadedStore();
      const tree = store.getAllTokens();
      const sp = (tree as any).spacing.spacing['1.5'];
      expect(sp.$value).toBe('0.375rem');
      expect(sp.$type).toBe('dimension');

      const radius = (tree as any).spacing.radius.sm;
      expect(radius.$value).toBe('0.25rem');
      expect(radius.$type).toBe('dimension');
    });

    it('logs token counts per category', () => {
      const spy = vi.spyOn(console, 'log').mockImplementation(() => {});
      loadedStore();
      expect(spy).toHaveBeenCalledWith(
        expect.stringMatching(/Loaded \d+ color, \d+ typography, \d+ spacing tokens/),
      );
      spy.mockRestore();
    });
  });

  describe('getAllTokens()', () => {
    it('returns a deep clone (mutations do not affect store)', () => {
      const store = loadedStore();
      const tree1 = store.getAllTokens();
      (tree1 as any).color.red['500'].$value = 'MUTATED';
      const tree2 = store.getAllTokens();
      expect((tree2 as any).color.red['500'].$value).toBe('#DE0037');
    });
  });

  describe('getByCategory()', () => {
    it('returns only color tokens for category "color"', () => {
      const store = loadedStore();
      const colors = store.getByCategory('color');
      expect(colors).toHaveProperty('red');
      expect(colors).toHaveProperty('black');
      expect(colors).not.toHaveProperty('font-family');
      expect(colors).not.toHaveProperty('spacing');
    });

    it('returns only typography tokens for category "typography"', () => {
      const store = loadedStore();
      const typo = store.getByCategory('typography');
      expect(typo).toHaveProperty('font-family');
      expect(typo).toHaveProperty('font-weight');
      expect(typo).toHaveProperty('text-size');
      expect(typo).not.toHaveProperty('red');
    });

    it('returns only spacing tokens for category "spacing"', () => {
      const store = loadedStore();
      const spacing = store.getByCategory('spacing');
      expect(spacing).toHaveProperty('spacing');
      expect(spacing).toHaveProperty('radius');
      expect(spacing).not.toHaveProperty('red');
    });
  });

  describe('getByFamily()', () => {
    it('returns only red family color tokens', () => {
      const store = loadedStore();
      const red = store.getByFamily('color', 'red');
      expect(red).toHaveProperty('50');
      expect(red).toHaveProperty('500');
      expect(red).toHaveProperty('900');
      expect(red).not.toHaveProperty('gray');
    });

    it('returns empty object for non-existent family', () => {
      const store = loadedStore();
      const result = store.getByFamily('color', 'nonexistent');
      expect(result).toEqual({});
    });

    it('wraps a leaf alias token when queried as family', () => {
      const store = loadedStore();
      const black = store.getByFamily('color', 'black');
      expect(black).toHaveProperty('black');
      expect((black as any).black.$value).toBe('{color.gray.900}');
    });
  });

  describe('listTokens()', () => {
    it('returns flat array with dot-notation paths', () => {
      const store = loadedStore();
      const list = store.listTokens();
      const red500 = list.find((t) => t.path === 'color.red.500');
      expect(red500).toBeDefined();
      expect(red500!.value).toBe('#DE0037');
      expect(red500!.type).toBe('color');
      expect(red500!.category).toBe('color');
    });

    it('filters by category when provided', () => {
      const store = loadedStore();
      const spacingTokens = store.listTokens('spacing');
      expect(spacingTokens.every((t) => t.category === 'spacing')).toBe(true);
      expect(spacingTokens.length).toBe(3);
    });

    it('includes all tokens when no category filter', () => {
      const store = loadedStore();
      const all = store.listTokens();
      // 6 colors + 3 typography + 3 spacing = 12
      expect(all.length).toBe(12);
    });
  });

  describe('toJSON()', () => {
    it('produces valid parseable JSON', () => {
      const store = loadedStore();
      const json = store.toJSON();
      expect(() => JSON.parse(json)).not.toThrow();
    });

    it('serializes only the requested category', () => {
      const store = loadedStore();
      const json = store.toJSON('color');
      const parsed = JSON.parse(json);
      expect(parsed).toHaveProperty('color');
      expect(parsed).not.toHaveProperty('typography');
      expect(parsed).not.toHaveProperty('spacing');
    });

    it('serializes all categories when no filter', () => {
      const store = loadedStore();
      const json = store.toJSON();
      const parsed = JSON.parse(json);
      expect(parsed).toHaveProperty('color');
      expect(parsed).toHaveProperty('typography');
      expect(parsed).toHaveProperty('spacing');
    });
  });

  describe('addSemanticTokens()', () => {
    it('adds semantic tokens to the tree', () => {
      const store = loadedStore();
      store.addSemanticTokens([
        {
          name: 'semantic/destructive',
          aliasRef: '{color.red.500}',
          category: 'color',
          source: 'SAILSemanticColor.NEGATIVE',
        },
      ]);
      const tree = store.getAllTokens();
      const destructive = (tree as any).color.semantic.destructive;
      expect(destructive.$value).toBe('{color.red.500}');
      expect(destructive.$type).toBe('color');
      expect(destructive.$description).toBe('SAILSemanticColor.NEGATIVE');
    });
  });

  describe('DTCG format invariants', () => {
    it('every listed token has value, type, and category', () => {
      const store = loadedStore();
      const tokens = store.listTokens();
      for (const t of tokens) {
        expect(typeof t.value).toBe('string');
        expect(t.value.length).toBeGreaterThan(0);
        expect(typeof t.type).toBe('string');
        expect(t.type.length).toBeGreaterThan(0);
        expect(['color', 'typography', 'spacing']).toContain(t.category);
      }
    });

    it('every token in the tree has $value, $type, and $description', () => {
      const store = loadedStore();
      const tokens = store.listTokens();
      // Verify count — all tokens should have descriptions set
      expect(tokens.length).toBe(12);
      // Spot-check a few known tokens in the serialized tree
      const parsed = JSON.parse(store.toJSON());
      expect(parsed.color.red['500'].$description).toBeDefined();
      expect(parsed.typography['font-family'].base.$description).toBeDefined();
      expect(parsed.spacing.spacing['1.5'].$description).toBeDefined();
    });
  });
});
