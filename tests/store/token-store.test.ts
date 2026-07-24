import { assertEquals, assertExists } from "@std/assert";
import { TokenStore } from "../../src/store/token-store.ts";

const fixtureData = JSON.parse(
  await Deno.readTextFile("tests/fixtures/tokens.json"),
);

Deno.test("TokenStore", async (t) => {
  await t.step("loadFromJSON returns correct counts per category", () => {
    const store = new TokenStore();
    const result = store.loadFromJSON(fixtureData);

    assertEquals(result.counts.color, 4);
    assertEquals(result.counts.typography, 2);
    assertEquals(result.counts.spacing, 2);
  });

  await t.step("getByCategory('color') returns only color tokens", () => {
    const store = new TokenStore();
    store.loadFromJSON(fixtureData);

    const colorGroup = store.getByCategory("color");

    assertExists(colorGroup.red);
    assertExists(colorGroup.blue);
    assertEquals(Object.keys(colorGroup).length, 2);
  });

  await t.step(
    "getByFamily('color', 'red') returns only the red family",
    () => {
      const store = new TokenStore();
      store.loadFromJSON(fixtureData);

      const redFamily = store.getByFamily("color", "red");

      assertExists(redFamily["500"]);
      assertExists(redFamily["700"]);
      assertEquals(Object.keys(redFamily).length, 2);
    },
  );

  await t.step("listTokens() without filter returns all tokens", () => {
    const store = new TokenStore();
    store.loadFromJSON(fixtureData);

    const allTokens = store.listTokens();

    assertEquals(allTokens.length, 8);
  });

  await t.step(
    "listTokens('color') returns only color category tokens",
    () => {
      const store = new TokenStore();
      store.loadFromJSON(fixtureData);

      const colorTokens = store.listTokens("color");

      assertEquals(colorTokens.length, 4);
      for (const token of colorTokens) {
        assertEquals(token.category, "color");
      }
    },
  );

  await t.step("getFamilies('color') returns correct family keys", () => {
    const store = new TokenStore();
    store.loadFromJSON(fixtureData);

    const families = store.getFamilies("color");

    assertEquals(families.sort(), ["blue", "red"]);
  });
});
