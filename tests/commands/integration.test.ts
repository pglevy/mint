/**
 * Integration tests for the Mint CLI.
 *
 * Spawns the CLI as a child process using Deno.Command and asserts
 * on stdout, stderr, and exit code.
 *
 * Validates: Requirements 11.4
 */

import { assert, assertEquals, assertStringIncludes } from "@std/assert";

async function runCli(
  args: string[],
): Promise<{ stdout: string; stderr: string; code: number }> {
  const cmd = new Deno.Command("deno", {
    args: [
      "run",
      "--allow-net=raw.githubusercontent.com",
      "--allow-env",
      "main.ts",
      ...args,
    ],
    stdout: "piped",
    stderr: "piped",
    stdin: "null",
  });
  const output = await cmd.output();
  return {
    stdout: new TextDecoder().decode(output.stdout),
    stderr: new TextDecoder().decode(output.stderr),
    code: output.code,
  };
}

Deno.test("mcp-schema success — exit 0, valid JSON with tools array", async () => {
  const result = await runCli(["mcp-schema"]);
  assertEquals(result.code, 0);
  const parsed = JSON.parse(result.stdout);
  assert(Array.isArray(parsed.tools), "Expected 'tools' to be an array");
  assert(parsed.tools.length > 0, "Expected at least one tool in schema");
});

Deno.test("invalid category error — exit 1, stderr mentions the invalid value", async () => {
  const result = await runCli(["list", "--category", "bogus"]);
  assertEquals(result.code, 1);
  assertStringIncludes(result.stderr, "bogus");
});

Deno.test("--json mode — exit 0, stdout is parseable JSON", async () => {
  const result = await runCli(["mcp-schema", "--json"]);
  assertEquals(result.code, 0);
  const parsed = JSON.parse(result.stdout);
  assert(typeof parsed === "object" && parsed !== null, "Expected JSON object");
});

Deno.test("--help output — exit 0, contains mint and Examples", async () => {
  const result = await runCli(["--help"]);
  assertEquals(result.code, 0);
  assertStringIncludes(result.stdout, "Mint");
  assertStringIncludes(result.stdout, "Examples");
});

Deno.test("no-args output — exit 0, concise help text", async () => {
  const result = await runCli([]);
  assertEquals(result.code, 0);
  assertStringIncludes(result.stdout, "mint");
  assertStringIncludes(result.stdout, "--help");
});

Deno.test("unrecognized subcommand — exit 1, stderr contains error", async () => {
  const result = await runCli(["foobar"]);
  assertEquals(result.code, 1);
  assertStringIncludes(result.stderr, "Unknown subcommand");
  assertStringIncludes(result.stderr, "foobar");
});
