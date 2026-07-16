import { afterAll, beforeEach, describe, expect, test } from "bun:test";
import { JSDOM } from "jsdom";
import { clearLegacyAuthStorage } from "./storage";

const previousWindow = globalThis.window;
const previousDocument = globalThis.document;
const previousLocalStorage = globalThis.localStorage;
const testDOM = new JSDOM("<!doctype html><html><body></body></html>", { url: "http://localhost/" });

Object.assign(globalThis, {
  document: testDOM.window.document,
  localStorage: testDOM.window.localStorage,
  window: testDOM.window,
});

beforeEach(() => {
  localStorage.clear();
  document.cookie = "fk_token=; Path=/; Max-Age=0";
});

afterAll(() => {
  Object.assign(globalThis, {
    document: previousDocument,
    localStorage: previousLocalStorage,
    window: previousWindow,
  });
  testDOM.window.close();
});

describe("clearLegacyAuthStorage", () => {
  test("removes the legacy JavaScript-readable token", () => {
    localStorage.setItem("fk_token", "legacy-token");
    document.cookie = "fk_token=legacy-token; Path=/";

    clearLegacyAuthStorage();

    expect(localStorage.getItem("fk_token")).toBeNull();
    expect(document.cookie).not.toContain("fk_token=");
  });

  test("does not mutate cookies when no legacy local token exists", () => {
    document.cookie = "preference=compact; Path=/";

    clearLegacyAuthStorage();

    expect(document.cookie).toContain("preference=compact");
  });
});
