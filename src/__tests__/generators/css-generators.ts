import fc from 'fast-check';

// ── Helpers ─────────────────────────────────────────────────────────

/** Random lowercase alpha string (1–8 chars) for CSS property name segments */
const cssNameArb: fc.Arbitrary<string> = fc
  .array(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz'.split('')), {
    minLength: 1,
    maxLength: 8,
  })
  .map((chars) => chars.join(''));

/** Random 6-digit hex value (uppercase) */
const hexValueArb: fc.Arbitrary<string> = fc
  .array(fc.constantFrom(...'0123456789ABCDEF'.split('')), {
    minLength: 6,
    maxLength: 6,
  })
  .map((chars) => `#${chars.join('')}`);

/** Random Tailwind-style step number */
const stepArb: fc.Arbitrary<number> = fc.constantFrom(
  50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950,
);

// ── Single Property Arbitraries ─────────────────────────────────────

/** Generates a single --color-{family}-{step}: #{hex}; line */
export const colorPropertyArb: fc.Arbitrary<string> = fc
  .tuple(cssNameArb, stepArb, hexValueArb)
  .map(([family, step, hex]) => `  --color-${family}-${step}: ${hex};`);

/** Generates a single --text-{name}: {value}; line */
const textPropertyArb: fc.Arbitrary<string> = fc
  .tuple(cssNameArb, fc.constantFrom('0.75rem', '1rem', '1.25rem', '1.5rem', '1.75rem'))
  .map(([name, value]) => `  --text-${name}: ${value};`);

/** Generates a single --spacing-{name}: {value}; line */
const spacingPropertyArb: fc.Arbitrary<string> = fc
  .tuple(cssNameArb, fc.constantFrom('0.25rem', '0.375rem', '0.5rem', '1rem', '1.5rem', '2rem'))
  .map(([name, value]) => `  --spacing-${name}: ${value};`);

/** Generates a single --radius-{name}: {value}; line */
const radiusPropertyArb: fc.Arbitrary<string> = fc
  .tuple(cssNameArb, fc.constantFrom('0.125rem', '0.25rem', '0.375rem', '0.5rem', '0.75rem'))
  .map(([name, value]) => `  --radius-${name}: ${value};`);

/** Generates a --color-{name}: var(--color-{target}); alias line */
export const aliasPropertyArb: fc.Arbitrary<string> = fc
  .tuple(cssNameArb, cssNameArb, stepArb)
  .filter(([name, target]) => name !== target)
  .map(([name, target, step]) => `  --color-${name}: var(--color-${target}-${step});`);

// ── Theme Block Arbitraries ─────────────────────────────────────────

/**
 * Generates a CSS @theme block string with random --color-*, --text-*,
 * --spacing-*, and --radius-* properties.
 */
export const themeBlockArb: fc.Arbitrary<string> = fc
  .tuple(
    fc.array(colorPropertyArb, { minLength: 1, maxLength: 6 }),
    fc.array(textPropertyArb, { minLength: 0, maxLength: 3 }),
    fc.array(spacingPropertyArb, { minLength: 0, maxLength: 3 }),
    fc.array(radiusPropertyArb, { minLength: 0, maxLength: 3 }),
  )
  .map(([colors, texts, spacings, radii]) => {
    const lines = [...colors, ...texts, ...spacings, ...radii];
    return `@theme {\n${lines.join('\n')}\n}`;
  });

/**
 * Generates a @theme block that includes both concrete color tokens
 * and var() aliases between them.
 */
export const themeBlockWithAliasesArb: fc.Arbitrary<string> = fc
  .tuple(
    fc.array(colorPropertyArb, { minLength: 2, maxLength: 6 }),
    fc.array(aliasPropertyArb, { minLength: 1, maxLength: 3 }),
  )
  .map(([colors, aliases]) => {
    const lines = [...colors, ...aliases];
    return `@theme {\n${lines.join('\n')}\n}`;
  });
