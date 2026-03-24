// ── Token Categories ─────────────────────────────────────────────────

export type TokenCategory = 'color' | 'typography' | 'spacing';

export const VALID_CATEGORIES: TokenCategory[] = ['color', 'typography', 'spacing'];

// ── CSS Parser Types ────────────────────────────────────────────────

export interface CSSParserOptions {
  filePath: string;
}

export interface RawColorToken {
  name: string;
  value: string;
  family: string;
  step: number;
}

export interface RawTypographyToken {
  name: string;
  value: string;
  dtcgType: 'fontFamily' | 'fontWeight' | 'dimension';
}

export interface RawSpacingToken {
  name: string;
  value: string;
}

export interface AliasRecord {
  tokenName: string;
  aliasOf: string;
}

export interface CSSParseResult {
  colors: RawColorToken[];
  typography: RawTypographyToken[];
  spacing: RawSpacingToken[];
  aliases: AliasRecord[];
}

// ── SAIL Parser Types ───────────────────────────────────────────────

export interface SAILParserOptions {
  filePath: string;
}

export interface SAILEnumMapping {
  typeName: string;
  values: Record<string, string>;
}

export interface SAILParseResult {
  enums: SAILEnumMapping[];
}

// ── Token Store Types ───────────────────────────────────────────────

export interface DTCGToken {
  $value: string;
  $type: string;
  $description?: string;
}

export interface DTCGGroup {
  [key: string]: DTCGToken | DTCGGroup;
}

// ── Semantic Mapper Types ───────────────────────────────────────────

export interface SemanticMapping {
  name: string;
  aliasRef: string;
  category: TokenCategory;
  source: string;
}
