export type TokenCategory = 'color' | 'typography' | 'spacing';

export const VALID_CATEGORIES: TokenCategory[] = ['color', 'typography', 'spacing'];

export interface DTCGToken {
  $value: string;
  $type: string;
  $description?: string;
}

export interface DTCGGroup {
  [key: string]: DTCGToken | DTCGGroup;
}
