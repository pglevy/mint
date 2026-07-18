import { assertEquals, assertNotMatch } from "@std/assert";
import { formatValue } from "../../src/output/format-value.ts";

Deno.test("formatValue - string passthrough", () => {
  assertEquals(formatValue("#ff0000"), "#ff0000");
  assertEquals(formatValue("16px"), "16px");
  assertEquals(formatValue(""), "");
});

Deno.test("formatValue - number to string", () => {
  assertEquals(formatValue(16), "16");
  assertEquals(formatValue(0), "0");
  assertEquals(formatValue(3.14), "3.14");
});

Deno.test("formatValue - object with sorted keys", () => {
  const result = formatValue({ fontSize: "16px", color: "#000" });
  assertEquals(result, "color: #000, fontSize: 16px");
});

Deno.test("formatValue - nested objects are JSON stringified", () => {
  const result = formatValue({ shadow: { x: 0, y: 4 } });
  // The nested object should be JSON.stringified, not rendered as [object Object]
  assertEquals(result, 'shadow: {"x":0,"y":4}');
});

Deno.test("formatValue - real Sailwind composite token produces readable output", () => {
  const typography = {
    fontSize: "1.5rem",
    lineHeight: "2",
    fontWeight: "700",
  };
  const result = formatValue(typography);

  // Must NOT contain [object Object]
  assertNotMatch(result, /\[object Object\]/);

  // Should contain all key-value pairs in sorted order
  assertEquals(result, "fontSize: 1.5rem, fontWeight: 700, lineHeight: 2");
});

Deno.test("formatValue - boolean and null nested values", () => {
  const result = formatValue({ active: true, removed: false, ref: null });
  assertEquals(result, "active: true, ref: null, removed: false");
});
