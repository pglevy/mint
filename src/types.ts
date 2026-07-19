export type TokenCategory = "color" | "gradient" | "typography" | "spacing";

export const VALID_CATEGORIES: TokenCategory[] = [
  "color",
  "gradient",
  "typography",
  "spacing",
];

export interface DTCGToken {
  $value: string | number | Record<string, unknown> | unknown[];
  $type: string;
  $description?: string;
}

export interface DTCGGroup {
  [key: string]: DTCGToken | DTCGGroup;
}
