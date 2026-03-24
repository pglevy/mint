/**
 * MCP tool schema command — outputs a static JSON object describing
 * the available Mint tool calls, their parameters, and return types.
 */
export function mcpSchemaCommand(): void {
  const schema = {
    tools: [
      {
        name: 'list_tokens',
        description:
          'List all design tokens, optionally filtered by category and family',
        parameters: {
          type: 'object',
          properties: {
            category: {
              type: 'string',
              enum: ['color', 'typography', 'spacing'],
              description: 'Filter tokens by category',
            },
            family: {
              type: 'string',
              description:
                "Filter tokens by family (e.g., 'red', 'gray'). Only applies when category is specified.",
            },
          },
        },
        returns: {
          type: 'array',
          description:
            'Array of token objects with path, value, type, and category fields',
        },
      },
      {
        name: 'export_tokens',
        description: 'Export design tokens as W3C DTCG JSON',
        parameters: {
          type: 'object',
          properties: {
            category: {
              type: 'string',
              enum: ['color', 'typography', 'spacing'],
              description: 'Filter tokens by category',
            },
            out: {
              type: 'string',
              description:
                'File path to write the export to. If omitted, outputs to stdout.',
            },
          },
        },
        returns: {
          type: 'object',
          description: 'W3C DTCG-formatted JSON token tree',
        },
      },
    ],
  };

  process.stdout.write(JSON.stringify(schema, null, 2) + '\n');
}
