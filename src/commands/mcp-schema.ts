/**
 * MCP tool schema command — outputs a static JSON object describing
 * the available Mint MCP tool calls, their parameters, and return types.
 *
 * No network required — this is pure static output.
 */

import { createColorContext } from "../output/color.ts";
import { createWriter } from "../output/writer.ts";

interface GlobalOptions {
  json: boolean;
  quiet: boolean;
  noColor: boolean;
  debug: boolean;
  help: boolean;
  version: boolean;
}

const MCP_SCHEMA = {
  tools: [
    {
      name: "list_tokens",
      description: "List design tokens with optional category filter",
      inputSchema: {
        type: "object",
        properties: {
          category: {
            type: "string",
            enum: ["color", "typography", "spacing"],
            description: "Filter tokens by category",
          },
        },
        required: [],
      },
      returnType: {
        type: "array",
        items: {
          type: "object",
          properties: {
            path: { type: "string" },
            value: {
              oneOf: [
                { type: "string" },
                { type: "number" },
                { type: "object" },
              ],
            },
            type: { type: "string" },
            category: {
              type: "string",
              enum: ["color", "typography", "spacing"],
            },
          },
        },
        description:
          "Array of token objects with path, value, type, and category",
      },
    },
    {
      name: "export_tokens",
      description: "Export tokens as DTCG JSON",
      inputSchema: {
        type: "object",
        properties: {
          category: {
            type: "string",
            enum: ["color", "typography", "spacing"],
            description: "Filter exported tokens by category",
          },
        },
        required: [],
      },
      returnType: {
        type: "object",
        description: "W3C DTCG-formatted JSON token tree",
      },
    },
  ],
};

export function mcpSchemaCommand(
  _args: string[],
  globals: GlobalOptions,
): void {
  const color = createColorContext({ noColor: globals.noColor });
  const writer = createWriter(
    { quiet: globals.quiet, json: globals.json },
    color,
  );

  writer.json(MCP_SCHEMA);
}
