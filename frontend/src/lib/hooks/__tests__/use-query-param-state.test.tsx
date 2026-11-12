import { describe, it, expect } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter, useLocation, useNavigationType } from "react-router-dom";
import {
  useQueryParamState,
  withParam,
  type QueryParamCodec,
} from "../use-query-param-state";

const monthCodec: QueryParamCodec<string> = {
  key: "m",
  parse: (raw) => (/^\d{4}-(0[1-9]|1[0-2])$/.test(raw) ? raw : null),
  serialize: (v) => v,
  current: () => "2026-11",
  equals: (a, b) => a === b,
};

function LocationProbe() {
  const loc = useLocation();
  const navType = useNavigationType();
  return (
    <>
      <div data-testid="loc">{loc.pathname + loc.search}</div>
      <div data-testid="nav">{navType}</div>
    </>
  );
}

function Probe({ codec, next }: { codec: QueryParamCodec<string>; next?: string | undefined }) {
  const [value, set] = useQueryParamState(codec);
  return (
    <>
      <div data-testid="value">{value}</div>
      <button onClick={() => next && set(next)}>set</button>
    </>
  );
}

function setup(url: string, codec = monthCodec, next?: string) {
  return render(
    <MemoryRouter initialEntries={[url]}>
      <Probe codec={codec} next={next} />
      <LocationProbe />
    </MemoryRouter>,
  );
}

const loc = () => screen.getByTestId("loc").textContent;
const nav = () => screen.getByTestId("nav").textContent;

describe("useQueryParamState", () => {
  it("reads current when absent", () => {
    setup("/p");
    expect(screen.getByTestId("value").textContent).toBe("2026-11");
    expect(loc()).toBe("/p");
  });

  it("reads a valid param", () => {
    setup("/p?m=2026-01");
    expect(screen.getByTestId("value").textContent).toBe("2026-01");
    expect(loc()).toBe("/p?m=2026-01");
  });

  it("invalid param falls back and is removed with replace", async () => {
    setup("/p?m=2026-13&x=1");
    expect(screen.getByTestId("value").textContent).toBe("2026-11");
    await waitFor(() => expect(loc()).toBe("/p?x=1"));
    expect(nav()).toBe("REPLACE");
  });

  it("empty param is removed", async () => {
    setup("/p?m=");
    await waitFor(() => expect(loc()).toBe("/p"));
  });

  it("current value in URL is removed", async () => {
    setup("/p?m=2026-11");
    await waitFor(() => expect(loc()).toBe("/p"));
  });

  it("set non-current writes param with replace and keeps others", async () => {
    setup("/p?x=1", monthCodec, "2026-02");
    fireEvent.click(screen.getByText("set"));
    await waitFor(() => expect(loc()).toBe("/p?x=1&m=2026-02"));
    expect(nav()).toBe("REPLACE");
  });

  it("set current deletes param", async () => {
    setup("/p?m=2026-02", monthCodec, "2026-11");
    fireEvent.click(screen.getByText("set"));
    await waitFor(() => expect(loc()).toBe("/p"));
  });

  it("normalizing codec rewrites once", async () => {
    let serializeCalls = 0;
    const codec: QueryParamCodec<string> = {
      ...monthCodec,
      parse: (raw) =>
        raw === "2026-03-17" || raw === "2026-03-01" ? "2026-03-01" : monthCodec.parse(raw),
      serialize: (v) => {
        serializeCalls++;
        return v;
      },
    };
    setup("/p?m=2026-03-17", codec);
    await waitFor(() => expect(loc()).toBe("/p?m=2026-03-01"));
    const callsAfterSettle = serializeCalls;
    await Promise.resolve();
    await Promise.resolve();
    expect(loc()).toBe("/p?m=2026-03-01");
    expect(serializeCalls).toBe(callsAfterSettle);
  });
});

describe("withParam", () => {
  it("sets a non-current value, keeping others", () => {
    expect(withParam(new URLSearchParams("x=1"), monthCodec, "2026-02")).toBe("?x=1&m=2026-02");
  });

  it("deletes the param for the current value", () => {
    expect(withParam(new URLSearchParams("m=2026-02"), monthCodec, "2026-11")).toBe("");
  });
});
