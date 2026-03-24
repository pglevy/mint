import fc from 'fast-check';
import type {
  TokenCategory,
  DTCGToken,
  RawColorToken,
  RawTypographyToken,
  RawSpacingToken,
  CSSParseResult,
  AliasRecord,
} from '../../types.js';

// ── Primitive Arbitraries ───────────────────────────────────────────

/** Random lowercase alpha string (1–12 chars) for token names */
export const tokenNameArb: fc.Arbitrary<string> = fc
  .array(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz'.split('')), {
    minLength: 1,
    maxLength: 12,
  })
  .map((chars) => chars.join(''));

/** Random valid 6-digit hex color string, e.g. "#AB12CD" */
export const hexColorArb: fc.Arbitrary<string> = fc
  .array(fc.constantFrom(...'0123456789ABCDEF'.split('')), {
    minLength: 6,
    maxLength: 6,
  })
  .map((chars) => `#${chars.join('')}`);

/** One of the three valid token categories */
export const tokenCategoryArb: fc.Arbitrary<TokenCategory> = fc.constantFrom(
  'color' as TokenCategory,
  'typography' as TokenCategory,
  'spacing' as TokenCategory,
);

/** Random path segment for DTCG tree nesting (lowercase alpha, 1–8 chars) */
export const tokenPathSegmentArb: fc.Arbitrary<string> = fc
  .array(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz'.split('')), {
    minLength: 1,
    maxLength: 8,
  })
  .map((chars) => chars.join(''));

// ── Token Object Arbitraries ────────────────────────────────────────

/** Random DTCGToken with valid $value, $type, and optional $description */
export const dtcgTokenArb: fc.Arbitrary<DTCGToken> = fc.record({
  $value: hexColorArb,
  $type: fc.constantFrom('color', 'dimension', 'fontFamily', 'fontWeight'),
  $description: fc.option(fc.string({ minLength: 1, maxLength: 40 }), { nil: undefined }),
});

/** Random color step number (common Tailwind scale values) */
const colorStepArb: fc.Arbitrary<number> = fc.constantFrom(
  50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950,
);

/** Random RawColorToken with family and step */
export const colorTokenArb: fc.Arbitrary<RawColorToken> = fc
  .tuple(tokenNameArb, hexColorArb, colorStepArb)
  .map(([family, value, step]) => ({
    name: `${family}-${step}`,
    value,
    family,
    step,
  }));

/** Random RawTypographyToken */
export const typographyTokenArb: fc.Arbitrary<RawTypographyToken> = fc
  .tuple(
    tokenNameArb,
    fc.constantFrom('fontFamily', 'fontWeight', 'dimension') as fc.Arbitrary<
      'fontFamily' | 'fontWeight' | 'dimension'
    >,
  )
  .chain(([name, dtcgType]) => {
    const valueArb =
      dtcgType === 'fontFamily'
        ? fc.constant("'Open Sans', sans-serif")
        : dtcgType === 'fontWeight'
          ? fc.constantFrom('400', '500', '600', '700')
          : fc.constantFrom('0.75rem', '1rem', '1.25rem', '1.5rem', '1.75rem');
    return valueArb.map((value) => ({ name, value, dtcgType }));
  });

/** Random RawSpacingToken */
export const spacingTokenArb: fc.Arbitrary<RawSpacingToken> = fc
  .tuple(
    tokenNameArb,
    fc.constantFrom('0.25rem', '0.375rem', '0.5rem', '0.625rem', '1rem', '1.5rem', '2rem'),
  )
  .map(([name, value]) => ({ name, value }));

/** Random AliasRecord */
export const aliasRecordArb: fc.Arbitrary<AliasRecord> = fc
  .tuple(tokenNameArb, tokenNameArb)
  .filter(([a, b]) => a !== b)
  .map(([tokenName, aliasOf]) => ({ tokenName, aliasOf }));

/** Generates a CSSParseResult with random tokens across all categories */
export const mixedTokenSetArb: fc.Arbitrary<CSSParseResult> = fc.record({
  colors: fc.array(colorTokenArb, { minLength: 1, maxLength: 10 }),
  typography: fc.array(typographyTokenArb, { minLength: 0, maxLength: 5 }),
  spacing: fc.array(spacingTokenArb, { minLength: 0, maxLength: 5 }),
  aliases: fc.array(aliasRecordArb, { minLength: 0, maxLength: 3 }),
});
