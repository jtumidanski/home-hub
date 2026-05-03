# Recurring Calendar Event End Date — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a series-end control ("On" / "After" / "Never") to the calendar event creation form, compose RFC 5545–compliant `RRULE` strings before submit, and add a backend safety net in `calendar-service` that rejects open-ended recurrence rules.

**Architecture:** Two layers of validation with identical semantics. Frontend gains a pure composition module (`frontend/src/lib/calendar/recurrence.ts`), an extended Zod schema, and an "Ends" sub-form in the event-create dialog. Backend gains a hand-rolled `parseRRULE` scanner plus a `ValidateRecurrence` function wired into both the create and update HTTP handlers (defense-in-depth only — the processor does not propagate the field on update). No HTTP route, request shape, DB schema, or sync-pipeline changes.

**Tech Stack:** TypeScript / React / Zod / react-hook-form / Vitest / @testing-library on the frontend; Go (standard library only) / `testing` package / `gorilla/mux` JSON:API helpers on the backend.

---

## Phase 1 — Frontend pure logic

### Task 1: Create `recurrence.ts` skeleton + `EndsMode` type and `eventStartInstant`

**Files:**
- Create: `frontend/src/lib/calendar/recurrence.ts`
- Create: `frontend/src/components/features/calendar/__tests__/recurrence.test.ts`

- [ ] **Step 1: Write the failing test for `eventStartInstant`**

Add to `frontend/src/components/features/calendar/__tests__/recurrence.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import { eventStartInstant } from "@/lib/calendar/recurrence";

describe("eventStartInstant", () => {
  it("returns the local-midnight Date for an all-day event", () => {
    const d = eventStartInstant("2026-05-06", "10:30", true, "America/New_York");
    expect(d.getFullYear()).toBe(2026);
    expect(d.getMonth()).toBe(4);
    expect(d.getDate()).toBe(6);
  });

  it("returns the parsed local Date for a timed event", () => {
    const d = eventStartInstant("2026-05-06", "10:30", false, "America/New_York");
    expect(d.getFullYear()).toBe(2026);
    expect(d.getMonth()).toBe(4);
    expect(d.getDate()).toBe(6);
    expect(d.getHours()).toBe(10);
    expect(d.getMinutes()).toBe(30);
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `npm --prefix frontend test -- recurrence.test.ts`
Expected: FAIL — module `@/lib/calendar/recurrence` does not exist.

- [ ] **Step 3: Implement the minimal `recurrence.ts`**

Create `frontend/src/lib/calendar/recurrence.ts`:

```ts
export type EndsMode = "on" | "after" | "never";

export function eventStartInstant(
  startDate: string,
  startTime: string,
  allDay: boolean,
  _timeZone: string,
): Date {
  if (allDay) {
    return new Date(`${startDate}T00:00:00`);
  }
  return new Date(`${startDate}T${startTime}`);
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `npm --prefix frontend test -- recurrence.test.ts`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/calendar/recurrence.ts frontend/src/components/features/calendar/__tests__/recurrence.test.ts
git commit -m "feat(calendar): scaffold recurrence module with eventStartInstant"
```

---

### Task 2: `formatUntilUTC` with DST-correct offset

**Files:**
- Modify: `frontend/src/lib/calendar/recurrence.ts`
- Modify: `frontend/src/components/features/calendar/__tests__/recurrence.test.ts`

- [ ] **Step 1: Write the failing tests**

Append to `recurrence.test.ts`:

```ts
import { formatUntilUTC } from "@/lib/calendar/recurrence";

describe("formatUntilUTC", () => {
  it("produces YYYYMMDDTHHMMSSZ for end-of-day in EST (winter)", () => {
    // 2026-01-15 23:59:59 America/New_York is UTC-5 → 2026-01-16 04:59:59Z
    expect(formatUntilUTC("2026-01-15", "America/New_York")).toBe("20260116T045959Z");
  });

  it("produces the correct value in EDT (summer DST)", () => {
    // 2026-07-15 23:59:59 America/New_York is UTC-4 → 2026-07-16 03:59:59Z
    expect(formatUntilUTC("2026-07-15", "America/New_York")).toBe("20260716T035959Z");
  });

  it("produces an unchanged value for UTC", () => {
    expect(formatUntilUTC("2026-06-10", "UTC")).toBe("20260610T235959Z");
  });
});
```

- [ ] **Step 2: Run to verify fail**

Run: `npm --prefix frontend test -- recurrence.test.ts`
Expected: FAIL — `formatUntilUTC` is not exported.

- [ ] **Step 3: Implement `formatUntilUTC`**

Add to `frontend/src/lib/calendar/recurrence.ts`:

```ts
// Returns the UTC instant corresponding to `endsOnDate` at 23:59:59 in
// `timeZone`, formatted as YYYYMMDDTHHMMSSZ. Derives the offset for the
// chosen date specifically via Intl.DateTimeFormat parts so DST transitions
// produce the right offset.
export function formatUntilUTC(endsOnDate: string, timeZone: string): string {
  const [y, m, d] = endsOnDate.split("-").map(Number);
  // Construct the wall-clock end-of-day as if it were UTC so we can compute
  // the offset induced by the named zone.
  const asUTC = Date.UTC(y, m - 1, d, 23, 59, 59);

  const fmt = new Intl.DateTimeFormat("en-US", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });

  // What does that UTC instant LOOK LIKE in the named zone?
  const parts = Object.fromEntries(
    fmt.formatToParts(new Date(asUTC)).map((p) => [p.type, p.value]),
  );
  const localAsUTC = Date.UTC(
    Number(parts.year),
    Number(parts.month) - 1,
    Number(parts.day),
    parts.hour === "24" ? 0 : Number(parts.hour),
    Number(parts.minute),
    Number(parts.second),
  );
  // Offset = how far the named zone's wall clock is from UTC at that moment.
  const offsetMs = localAsUTC - asUTC;
  const utcInstant = new Date(asUTC - offsetMs);

  const pad = (n: number, w = 2) => String(n).padStart(w, "0");
  return (
    `${utcInstant.getUTCFullYear()}` +
    `${pad(utcInstant.getUTCMonth() + 1)}` +
    `${pad(utcInstant.getUTCDate())}` +
    `T${pad(utcInstant.getUTCHours())}` +
    `${pad(utcInstant.getUTCMinutes())}` +
    `${pad(utcInstant.getUTCSeconds())}Z`
  );
}
```

- [ ] **Step 4: Run to verify pass**

Run: `npm --prefix frontend test -- recurrence.test.ts`
Expected: PASS — all 5 tests green.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/calendar/recurrence.ts frontend/src/components/features/calendar/__tests__/recurrence.test.ts
git commit -m "feat(calendar): add DST-aware formatUntilUTC for RRULE composition"
```

---

### Task 3: `composeRecurrenceRule` covering all three modes + presets

**Files:**
- Modify: `frontend/src/lib/calendar/recurrence.ts`
- Modify: `frontend/src/components/features/calendar/__tests__/recurrence.test.ts`

- [ ] **Step 1: Write the failing tests**

Append to `recurrence.test.ts`:

```ts
import { composeRecurrenceRule } from "@/lib/calendar/recurrence";

describe("composeRecurrenceRule", () => {
  const presets = [
    "RRULE:FREQ=DAILY",
    "RRULE:FREQ=WEEKLY",
    "RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR",
    "RRULE:FREQ=MONTHLY",
    "RRULE:FREQ=YEARLY",
  ];

  it("returns undefined when preset is empty", () => {
    expect(
      composeRecurrenceRule("", "on", "2026-06-10", 10, "2026-05-06", "09:00", "UTC"),
    ).toBeUndefined();
  });

  it("appends UNTIL for mode=on, across all presets", () => {
    for (const preset of presets) {
      const out = composeRecurrenceRule(preset, "on", "2026-06-10", 10, "2026-05-06", "09:00", "UTC");
      expect(out).toEqual([`${preset};UNTIL=20260610T235959Z`]);
    }
  });

  it("appends COUNT for mode=after", () => {
    const out = composeRecurrenceRule(
      "RRULE:FREQ=DAILY", "after", "", 5, "2026-05-06", "09:00", "America/New_York",
    );
    expect(out).toEqual(["RRULE:FREQ=DAILY;COUNT=5"]);
  });

  it("returns the preset unchanged for mode=never", () => {
    const out = composeRecurrenceRule(
      "RRULE:FREQ=WEEKLY", "never", "", 0, "2026-05-06", "09:00", "America/New_York",
    );
    expect(out).toEqual(["RRULE:FREQ=WEEKLY"]);
  });

  it("converts UNTIL to UTC for non-UTC zones (winter)", () => {
    const out = composeRecurrenceRule(
      "RRULE:FREQ=WEEKLY", "on", "2026-01-15", 0, "2026-01-01", "09:00", "America/New_York",
    );
    expect(out).toEqual(["RRULE:FREQ=WEEKLY;UNTIL=20260116T045959Z"]);
  });
});
```

- [ ] **Step 2: Run to verify fail**

Run: `npm --prefix frontend test -- recurrence.test.ts`
Expected: FAIL — `composeRecurrenceRule` not exported.

- [ ] **Step 3: Implement `composeRecurrenceRule`**

Append to `frontend/src/lib/calendar/recurrence.ts`:

```ts
export function composeRecurrenceRule(
  preset: string,
  mode: EndsMode,
  endsOnDate: string,
  endsAfterCount: number,
  _startDate: string,
  _startTime: string,
  timeZone: string,
): string[] | undefined {
  if (preset === "") return undefined;
  switch (mode) {
    case "on":
      return [`${preset};UNTIL=${formatUntilUTC(endsOnDate, timeZone)}`];
    case "after":
      return [`${preset};COUNT=${endsAfterCount}`];
    case "never":
      return [preset];
  }
}
```

- [ ] **Step 4: Run to verify pass**

Run: `npm --prefix frontend test -- recurrence.test.ts`
Expected: PASS — all tests green.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/calendar/recurrence.ts frontend/src/components/features/calendar/__tests__/recurrence.test.ts
git commit -m "feat(calendar): add composeRecurrenceRule for terminated RRULEs"
```

---

## Phase 2 — Frontend schema

### Task 4: Extend `eventFormSchema` with Ends fields and refinements

**Files:**
- Modify: `frontend/src/lib/schemas/calendar-event.schema.ts`
- Create: `frontend/src/lib/schemas/__tests__/calendar-event.schema.test.ts`

- [ ] **Step 1: Write the failing tests**

Create `frontend/src/lib/schemas/__tests__/calendar-event.schema.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import { eventFormSchema, createEventDefaults } from "@/lib/schemas/calendar-event.schema";

function baseValid() {
  return {
    title: "x",
    allDay: false,
    startDate: "2026-05-06",
    startTime: "09:00",
    endDate: "2026-05-06",
    endTime: "10:00",
    recurrence: "RRULE:FREQ=WEEKLY",
    location: "",
    description: "",
    calendarId: "cal-1",
    connectionId: "conn-1",
    endsMode: "on" as const,
    endsOnDate: "2027-05-06",
    endsAfterCount: 10,
    endsNeverConfirmed: false,
    endsOnDateUserEdited: false,
  };
}

describe("eventFormSchema — Ends fields", () => {
  it("createEventDefaults seeds the new fields", () => {
    const d = createEventDefaults();
    expect(d.endsMode).toBe("on");
    expect(d.endsOnDate).toBe("");
    expect(d.endsAfterCount).toBe(10);
    expect(d.endsNeverConfirmed).toBe(false);
    expect(d.endsOnDateUserEdited).toBe(false);
  });

  it("accepts a valid bounded recurring event", () => {
    expect(eventFormSchema.safeParse(baseValid()).success).toBe(true);
  });

  it("ignores Ends fields when recurrence is empty", () => {
    const r = eventFormSchema.safeParse({ ...baseValid(), recurrence: "", endsOnDate: "" });
    expect(r.success).toBe(true);
  });

  it("rejects mode=on with an end date before the start date", () => {
    const r = eventFormSchema.safeParse({ ...baseValid(), endsOnDate: "2026-04-01" });
    expect(r.success).toBe(false);
  });

  it("rejects mode=on with an end date more than 5 years out", () => {
    const r = eventFormSchema.safeParse({ ...baseValid(), endsOnDate: "2031-06-01" });
    expect(r.success).toBe(false);
  });

  it("rejects mode=after with count = 0", () => {
    const r = eventFormSchema.safeParse({ ...baseValid(), endsMode: "after", endsAfterCount: 0 });
    expect(r.success).toBe(false);
  });

  it("rejects mode=after with count = 731", () => {
    const r = eventFormSchema.safeParse({ ...baseValid(), endsMode: "after", endsAfterCount: 731 });
    expect(r.success).toBe(false);
  });

  it("rejects mode=never without confirmation", () => {
    const r = eventFormSchema.safeParse({
      ...baseValid(),
      endsMode: "never",
      endsNeverConfirmed: false,
    });
    expect(r.success).toBe(false);
  });

  it("accepts mode=never when confirmed", () => {
    const r = eventFormSchema.safeParse({
      ...baseValid(),
      endsMode: "never",
      endsNeverConfirmed: true,
    });
    expect(r.success).toBe(true);
  });
});
```

- [ ] **Step 2: Run to verify fail**

Run: `npm --prefix frontend test -- calendar-event.schema.test.ts`
Expected: FAIL — schema lacks the new fields.

- [ ] **Step 3: Modify `calendar-event.schema.ts`**

Replace the entire contents of `frontend/src/lib/schemas/calendar-event.schema.ts` with:

```ts
import { z } from "zod";

const FIVE_YEARS_PLUS_CUSHION_MS = 5 * 365 * 24 * 60 * 60 * 1000 + 24 * 60 * 60 * 1000;

export const eventFormSchema = z
  .object({
    title: z.string().min(1, "Title is required").max(1024, "Title must be 1024 characters or fewer"),
    allDay: z.boolean(),
    startDate: z.string().min(1, "Start date is required"),
    startTime: z.string(),
    endDate: z.string(),
    endTime: z.string(),
    recurrence: z.string(),
    location: z.string().max(1024, "Location must be 1024 characters or fewer"),
    description: z.string().max(8192, "Description must be 8192 characters or fewer"),
    calendarId: z.string().min(1, "Calendar is required"),
    connectionId: z.string(),
    endsMode: z.enum(["on", "after", "never"]),
    endsOnDate: z.string(),
    endsAfterCount: z.coerce.number().int(),
    endsNeverConfirmed: z.boolean(),
    endsOnDateUserEdited: z.boolean(),
  })
  .refine(
    (data) => {
      if (data.allDay) {
        return data.endDate >= data.startDate;
      }
      const start = `${data.startDate}T${data.startTime}`;
      const end = `${data.endDate}T${data.endTime}`;
      return end >= start;
    },
    { message: "End must be after start", path: ["endDate"] },
  )
  .refine(
    (data) => {
      if (data.recurrence === "" || data.endsMode !== "on") return true;
      if (!data.endsOnDate) return false;
      const start = new Date(`${data.startDate}T${data.allDay ? "00:00" : data.startTime || "00:00"}`);
      const end = new Date(`${data.endsOnDate}T23:59:59`);
      if (Number.isNaN(end.getTime()) || Number.isNaN(start.getTime())) return false;
      return end.getTime() > start.getTime();
    },
    { message: "End date must be after the event start", path: ["endsOnDate"] },
  )
  .refine(
    (data) => {
      if (data.recurrence === "" || data.endsMode !== "on" || !data.endsOnDate) return true;
      const start = new Date(`${data.startDate}T${data.allDay ? "00:00" : data.startTime || "00:00"}`);
      const end = new Date(`${data.endsOnDate}T23:59:59`);
      if (Number.isNaN(end.getTime()) || Number.isNaN(start.getTime())) return true;
      return end.getTime() - start.getTime() <= FIVE_YEARS_PLUS_CUSHION_MS;
    },
    { message: "End date cannot be more than 5 years out", path: ["endsOnDate"] },
  )
  .refine(
    (data) => {
      if (data.recurrence === "" || data.endsMode !== "after") return true;
      return Number.isInteger(data.endsAfterCount) && data.endsAfterCount >= 1 && data.endsAfterCount <= 730;
    },
    { message: "Must be between 1 and 730 occurrences", path: ["endsAfterCount"] },
  )
  .refine(
    (data) => {
      if (data.recurrence === "" || data.endsMode !== "never") return true;
      return data.endsNeverConfirmed === true;
    },
    { message: "Confirm you understand this event has no end date", path: ["endsNeverConfirmed"] },
  );

export type EventFormData = z.infer<typeof eventFormSchema>;

export const RECURRENCE_OPTIONS = [
  { value: "", label: "Does not repeat" },
  { value: "RRULE:FREQ=DAILY", label: "Daily" },
  { value: "RRULE:FREQ=WEEKLY", label: "Weekly" },
  { value: "RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR", label: "Weekdays (Mon-Fri)" },
  { value: "RRULE:FREQ=MONTHLY", label: "Monthly" },
  { value: "RRULE:FREQ=YEARLY", label: "Yearly" },
] as const;

function padTime(date: Date): string {
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function formatDate(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

export function createEventDefaults(prefilledStart?: Date): EventFormData {
  const now = prefilledStart ?? new Date();
  const roundedMinutes = Math.ceil(now.getMinutes() / 15) * 15;
  const start = new Date(now);
  start.setMinutes(roundedMinutes, 0, 0);

  const end = new Date(start);
  end.setHours(end.getHours() + 1);

  return {
    title: "",
    allDay: false,
    startDate: formatDate(start),
    startTime: padTime(start),
    endDate: formatDate(end),
    endTime: padTime(end),
    recurrence: "",
    location: "",
    description: "",
    calendarId: "",
    connectionId: "",
    endsMode: "on",
    endsOnDate: "",
    endsAfterCount: 10,
    endsNeverConfirmed: false,
    endsOnDateUserEdited: false,
  };
}
```

- [ ] **Step 4: Run to verify pass**

Run: `npm --prefix frontend test -- calendar-event.schema.test.ts`
Expected: PASS — all 9 tests green.

- [ ] **Step 5: Type-check the workspace**

Run: `npm --prefix frontend run build`
Expected: PASS — `tsc -b && vite build` completes without errors. (`event-form-dialog.tsx`'s `defaults` object will currently be missing the new fields; if it errors, fix it now by adding `endsMode: "on", endsOnDate: "", endsAfterCount: 10, endsNeverConfirmed: false, endsOnDateUserEdited: false` to both branches of the `defaults` `useMemo`.)

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/schemas/calendar-event.schema.ts frontend/src/lib/schemas/__tests__/calendar-event.schema.test.ts frontend/src/components/features/calendar/event-form-dialog.tsx
git commit -m "feat(calendar): extend event form schema with Ends fields"
```

---

## Phase 3 — Frontend UI

### Task 5: Render the "Ends" control + warning checkbox

**Files:**
- Modify: `frontend/src/components/features/calendar/event-form-dialog.tsx`

- [ ] **Step 1: Add a `useWatch` for `recurrence` and `endsMode`**

Just after the existing `const allDay = form.watch("allDay");` line, add:

```tsx
// eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
const recurrence = form.watch("recurrence");
// eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
const endsMode = form.watch("endsMode");
```

- [ ] **Step 2: Add the Ends control JSX immediately below the existing Repeats `FormField`**

Inside the `{!isEdit && ( <FormField ... name="recurrence" ... /> )}` block, change the surrounding JSX so the recurrence dropdown is followed by the Ends control. Replace the existing block:

```tsx
{!isEdit && (
  <FormField
    control={form.control}
    name="recurrence"
    render={({ field }) => (
      <FormItem>
        <FormLabel>Repeats</FormLabel>
        <FormControl>
          <select
            value={field.value}
            onChange={field.onChange}
            className="flex h-8 w-full rounded-lg border border-input bg-popover text-popover-foreground px-2.5 py-1.5 text-sm"
          >
            {RECURRENCE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </FormControl>
      </FormItem>
    )}
  />
)}
```

with:

```tsx
{!isEdit && (
  <>
    <FormField
      control={form.control}
      name="recurrence"
      render={({ field }) => (
        <FormItem>
          <FormLabel>Repeats</FormLabel>
          <FormControl>
            <select
              value={field.value}
              onChange={field.onChange}
              className="flex h-8 w-full rounded-lg border border-input bg-popover text-popover-foreground px-2.5 py-1.5 text-sm"
            >
              {RECURRENCE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </FormControl>
        </FormItem>
      )}
    />
    {recurrence !== "" && (
      <div className="space-y-2 rounded-lg border border-input p-3">
        <FormLabel>Ends</FormLabel>
        <FormField
          control={form.control}
          name="endsMode"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <div className="space-y-2">
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="radio"
                      name="endsMode"
                      value="on"
                      checked={field.value === "on"}
                      onChange={() => field.onChange("on")}
                    />
                    On
                    <FormField
                      control={form.control}
                      name="endsOnDate"
                      render={({ field: dateField }) => (
                        <Input
                          type="date"
                          aria-label="End date"
                          disabled={field.value !== "on"}
                          value={dateField.value}
                          onChange={(e) => {
                            form.setValue("endsOnDateUserEdited", true);
                            dateField.onChange(e.target.value);
                          }}
                          className="h-7 w-40"
                        />
                      )}
                    />
                  </label>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="radio"
                      name="endsMode"
                      value="after"
                      checked={field.value === "after"}
                      onChange={() => field.onChange("after")}
                    />
                    After
                    <FormField
                      control={form.control}
                      name="endsAfterCount"
                      render={({ field: countField }) => (
                        <Input
                          type="number"
                          aria-label="Occurrences"
                          min={1}
                          max={730}
                          disabled={field.value !== "after"}
                          value={countField.value}
                          onChange={(e) => countField.onChange(Number(e.target.value))}
                          className="h-7 w-20"
                        />
                      )}
                    />
                    <span className="text-muted-foreground">occurrences</span>
                  </label>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="radio"
                      name="endsMode"
                      value="never"
                      checked={field.value === "never"}
                      onChange={() => field.onChange("never")}
                    />
                    Never
                  </label>
                </div>
              </FormControl>
            </FormItem>
          )}
        />
        {endsMode === "on" && (
          <FormField
            control={form.control}
            name="endsOnDate"
            render={() => <FormMessage />}
          />
        )}
        {endsMode === "after" && (
          <FormField
            control={form.control}
            name="endsAfterCount"
            render={() => <FormMessage />}
          />
        )}
        {endsMode === "never" && (
          <div className="space-y-2 rounded-md bg-yellow-50 p-2 text-sm dark:bg-yellow-950">
            <p>This event will repeat forever. Are you sure?</p>
            <FormField
              control={form.control}
              name="endsNeverConfirmed"
              render={({ field }) => (
                <FormItem>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={field.value}
                      onChange={(e) => field.onChange(e.target.checked)}
                    />
                    I understand this event has no end date.
                  </label>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        )}
      </div>
    )}
  </>
)}
```

- [ ] **Step 3: Type-check & run the existing test suite**

Run: `npm --prefix frontend run build && npm --prefix frontend test`
Expected: PASS — TypeScript clean; existing tests still pass; new schema tests still pass.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/features/calendar/event-form-dialog.tsx
git commit -m "feat(calendar): render Ends control on event create form"
```

---

### Task 6: Add the three auto-update effects

**Files:**
- Modify: `frontend/src/components/features/calendar/event-form-dialog.tsx`

- [ ] **Step 1: Add `useRef` to imports and the three effects**

Update the import line at the top from:

```ts
import { useEffect, useMemo } from "react";
```

to:

```ts
import { useEffect, useMemo, useRef } from "react";
```

Immediately after the existing `useEffect` block that handles `open` (`useEffect(() => { if (open) { form.reset(defaults); ... } ... }, [open, ...])`), insert:

```tsx
const previousRecurrenceRef = useRef(defaults.recurrence);
// eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
const startDate = form.watch("startDate");
// eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
const endsOnDateUserEdited = form.watch("endsOnDateUserEdited");

function addOneYear(yyyymmdd: string): string {
  if (!yyyymmdd) return "";
  const [y, m, d] = yyyymmdd.split("-").map(Number);
  const next = new Date(y + 1, m - 1, d);
  const yy = next.getFullYear();
  const mm = String(next.getMonth() + 1).padStart(2, "0");
  const dd = String(next.getDate()).padStart(2, "0");
  return `${yy}-${mm}-${dd}`;
}

useEffect(() => {
  const prev = previousRecurrenceRef.current;
  if (prev === recurrence) return;
  if (prev === "" && recurrence !== "") {
    form.setValue("endsOnDate", addOneYear(form.getValues("startDate")));
    form.setValue("endsOnDateUserEdited", false);
  } else if (prev !== "" && recurrence === "") {
    form.setValue("endsMode", "on");
    form.setValue("endsOnDate", "");
    form.setValue("endsAfterCount", 10);
    form.setValue("endsNeverConfirmed", false);
    form.setValue("endsOnDateUserEdited", false);
  }
  previousRecurrenceRef.current = recurrence;
}, [recurrence, form]);

useEffect(() => {
  if (recurrence !== "" && endsMode === "on" && !endsOnDateUserEdited) {
    form.setValue("endsOnDate", addOneYear(startDate));
  }
}, [startDate, recurrence, endsMode, endsOnDateUserEdited, form]);
```

- [ ] **Step 2: Type-check & test**

Run: `npm --prefix frontend run build && npm --prefix frontend test`
Expected: PASS — clean.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/features/calendar/event-form-dialog.tsx
git commit -m "feat(calendar): auto-update Ends fields on recurrence/start changes"
```

---

### Task 7: Wire `composeRecurrenceRule` into the submit handler

**Files:**
- Modify: `frontend/src/components/features/calendar/event-form-dialog.tsx`

- [ ] **Step 1: Add the import**

Add to the top imports:

```tsx
import { composeRecurrenceRule } from "@/lib/calendar/recurrence";
```

- [ ] **Step 2: Replace the inline recurrence assembly in `onSubmit`**

In the `else` (create) branch of `onSubmit`, change:

```tsx
recurrence: values.recurrence ? [values.recurrence] : undefined,
```

to:

```tsx
recurrence: composeRecurrenceRule(
  values.recurrence,
  values.endsMode,
  values.endsOnDate,
  values.endsAfterCount,
  values.startDate,
  values.startTime,
  timeZone,
),
```

- [ ] **Step 3: Type-check & test**

Run: `npm --prefix frontend run build && npm --prefix frontend test`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/features/calendar/event-form-dialog.tsx
git commit -m "feat(calendar): submit composed RRULE with UNTIL/COUNT terminator"
```

---

### Task 8: Form-level interaction tests

**Files:**
- Create: `frontend/src/components/features/calendar/__tests__/event-form-dialog.test.tsx`

- [ ] **Step 1: Write the failing tests**

Create the file with:

```tsx
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EventFormDialog } from "../event-form-dialog";
import type { CalendarConnection, CalendarSource } from "@/types/models/calendar";

const mockCreate = vi.fn();
const mockUpdate = vi.fn();

vi.mock("@/lib/hooks/api/use-calendar", () => ({
  useCreateEvent: () => ({ mutateAsync: mockCreate }),
  useUpdateEvent: () => ({ mutateAsync: mockUpdate }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function makeConnection(): CalendarConnection {
  return {
    id: "conn-1",
    type: "calendar-connections",
    attributes: {
      provider: "google",
      providerAccountId: "x",
      providerEmail: "x@example.com",
      status: "connected",
      writeAccess: true,
      lastSyncedAt: null,
      lastError: null,
      createdAt: "2026-01-01T00:00:00Z",
    },
  } as CalendarConnection;
}

function makeSource(): CalendarSource {
  return {
    id: "src-1",
    type: "calendar-sources",
    attributes: {
      connectionId: "conn-1",
      name: "Primary",
      primary: true,
      colorHex: "#000",
      timeZone: "UTC",
      visible: true,
    },
  } as CalendarSource;
}

describe("EventFormDialog — Ends control", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreate.mockResolvedValue({});
  });

  it("hides the Ends control when 'Does not repeat' is selected", () => {
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
      />,
    );
    expect(screen.queryByLabelText("End date")).not.toBeInTheDocument();
  });

  it("shows the Ends control and seeds end date to start + 1y when a recurring preset is chosen", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=WEEKLY");
    const endInput = await screen.findByLabelText("End date");
    expect((endInput as HTMLInputElement).value).toBe("2027-05-06");
  });

  it("auto-updates the end date when start date changes (untouched end-date)", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=WEEKLY");
    const startInput = screen.getByLabelText("Start date");
    await user.clear(startInput);
    await user.type(startInput, "2026-06-10");
    await waitFor(() => {
      expect((screen.getByLabelText("End date") as HTMLInputElement).value).toBe("2027-06-10");
    });
  });

  it("leaves a user-edited end date alone when start date changes later", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=WEEKLY");
    const endInput = screen.getByLabelText("End date");
    await user.clear(endInput);
    await user.type(endInput, "2026-08-01");

    const startInput = screen.getByLabelText("Start date");
    await user.clear(startInput);
    await user.type(startInput, "2026-06-10");
    expect((screen.getByLabelText("End date") as HTMLInputElement).value).toBe("2026-08-01");
  });

  it("blocks submit until the Never confirmation checkbox is checked", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.type(screen.getByPlaceholderText("Event title"), "Standup");
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=WEEKLY");
    await user.click(screen.getByLabelText(/^never$/i));
    await user.click(screen.getByRole("button", { name: /create event/i }));
    expect(mockCreate).not.toHaveBeenCalled();
    await user.click(screen.getByLabelText(/I understand this event has no end date/i));
    await user.click(screen.getByRole("button", { name: /create event/i }));
    await waitFor(() => expect(mockCreate).toHaveBeenCalledTimes(1));
    expect(mockCreate.mock.calls[0]![0].data.recurrence).toEqual(["RRULE:FREQ=WEEKLY"]);
  });

  it("submits an UNTIL-terminated RRULE for mode=on", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.type(screen.getByPlaceholderText("Event title"), "Volleyball");
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=WEEKLY");
    const endInput = screen.getByLabelText("End date");
    await user.clear(endInput);
    await user.type(endInput, "2026-06-10");
    await user.click(screen.getByRole("button", { name: /create event/i }));
    await waitFor(() => expect(mockCreate).toHaveBeenCalledTimes(1));
    const rule = mockCreate.mock.calls[0]![0].data.recurrence[0] as string;
    expect(rule.startsWith("RRULE:FREQ=WEEKLY;UNTIL=20260611T")).toBe(true);
    expect(rule.endsWith("Z")).toBe(true);
  });

  it("submits a COUNT-terminated RRULE for mode=after", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    await user.type(screen.getByPlaceholderText("Event title"), "PT");
    await user.selectOptions(screen.getByLabelText("Repeats"), "RRULE:FREQ=DAILY");
    await user.click(screen.getByLabelText(/^after$/i));
    const countInput = screen.getByLabelText("Occurrences");
    await user.clear(countInput);
    await user.type(countInput, "5");
    await user.click(screen.getByRole("button", { name: /create event/i }));
    await waitFor(() => expect(mockCreate).toHaveBeenCalledTimes(1));
    expect(mockCreate.mock.calls[0]![0].data.recurrence).toEqual(["RRULE:FREQ=DAILY;COUNT=5"]);
  });

  it("resets Ends fields when switching back to 'Does not repeat'", async () => {
    const user = userEvent.setup();
    render(
      <EventFormDialog
        open
        onOpenChange={vi.fn()}
        connections={[makeConnection()]}
        sources={[makeSource()]}
        prefilledStart={new Date("2026-05-06T09:00:00")}
      />,
    );
    const repeats = screen.getByLabelText("Repeats");
    await user.selectOptions(repeats, "RRULE:FREQ=WEEKLY");
    expect(await screen.findByLabelText("End date")).toBeInTheDocument();
    await user.selectOptions(repeats, "");
    expect(screen.queryByLabelText("End date")).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run to verify all tests pass**

Run: `npm --prefix frontend test -- event-form-dialog.test.tsx`
Expected: PASS — 8 tests green.

If a test fails because the existing JSX does not have a `FormLabel` for the Repeats `<select>`, fix `event-form-dialog.tsx` so its `<FormItem>` block already exposes `<FormLabel>Repeats</FormLabel>` (it does in the current source — confirm). The `getByLabelText` calls in the tests rely on the labels rendered by the `FormItem`/`FormLabel` pair tying `<label htmlFor>` to the controls, which `Form` from `@/components/ui/form` already wires up.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/features/calendar/__tests__/event-form-dialog.test.tsx
git commit -m "test(calendar): cover Ends control behavior in event form"
```

---

## Phase 4 — Backend validator

### Task 9: `parseRRULE` (case-insensitive UNTIL/COUNT extraction)

**Files:**
- Create: `services/calendar-service/internal/event/recurrence_validator.go`
- Create: `services/calendar-service/internal/event/recurrence_validator_test.go`

- [ ] **Step 1: Write the failing tests**

Create `recurrence_validator_test.go`:

```go
package event

import (
	"testing"
	"time"
)

func TestParseRRULE(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		hasUntil  bool
		untilUTC  string // RFC3339 UTC
		hasCount  bool
		count     int
		expectErr bool
	}{
		{name: "weekly until date-time", line: "RRULE:FREQ=WEEKLY;UNTIL=20260611T035959Z",
			hasUntil: true, untilUTC: "2026-06-11T03:59:59Z"},
		{name: "weekly count", line: "RRULE:FREQ=WEEKLY;COUNT=10", hasCount: true, count: 10},
		{name: "lowercase tokens", line: "rrule:freq=weekly;until=20260611t035959z",
			hasUntil: true, untilUTC: "2026-06-11T03:59:59Z"},
		{name: "until date-only form", line: "RRULE:FREQ=WEEKLY;UNTIL=20260611",
			hasUntil: true, untilUTC: "2026-06-11T00:00:00Z"},
		{name: "open-ended", line: "RRULE:FREQ=WEEKLY"},
		{name: "malformed until -> err", line: "RRULE:FREQ=WEEKLY;UNTIL=garbage", expectErr: true},
		{name: "malformed count -> err", line: "RRULE:FREQ=WEEKLY;COUNT=abc", expectErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			until, count, hasUntil, hasCount, err := parseRRULE(tc.line)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if hasUntil != tc.hasUntil {
				t.Fatalf("hasUntil = %v, want %v", hasUntil, tc.hasUntil)
			}
			if tc.hasUntil {
				want, _ := time.Parse(time.RFC3339, tc.untilUTC)
				if !until.Equal(want) {
					t.Fatalf("until = %v, want %v", until, want)
				}
			}
			if hasCount != tc.hasCount {
				t.Fatalf("hasCount = %v, want %v", hasCount, tc.hasCount)
			}
			if hasCount && count != tc.count {
				t.Fatalf("count = %d, want %d", count, tc.count)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify fail**

Run: `go test ./services/calendar-service/internal/event/ -run TestParseRRULE`
Expected: FAIL — `parseRRULE` undefined.

- [ ] **Step 3: Implement `recurrence_validator.go` minimally**

Create `services/calendar-service/internal/event/recurrence_validator.go`:

```go
package event

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	maxUntilWindow = 5*365*24*time.Hour + 24*time.Hour
	minOccurrences = 1
	maxOccurrences = 730

	codeUnbounded  = "recurrence_unbounded"
	codeTooLong    = "recurrence_too_long"
	codeCountRange = "recurrence_count_out_of_range"
)

type RecurrenceError struct {
	Code    string
	Detail  string
	RuleRaw string
}

func (e *RecurrenceError) Error() string { return e.Code + ": " + e.Detail }

// parseRRULE extracts UNTIL and COUNT from a single "RRULE:..." line.
// Component names are matched case-insensitively per RFC 5545. UNTIL
// accepts both the date form (YYYYMMDD) and the date-time UTC form
// (YYYYMMDDTHHMMSSZ).
func parseRRULE(line string) (until time.Time, count int, hasUntil, hasCount bool, err error) {
	upper := strings.ToUpper(line)
	if !strings.HasPrefix(upper, "RRULE:") {
		return time.Time{}, 0, false, false, errors.New("not an RRULE line")
	}
	body := line[len("RRULE:"):]
	for _, kv := range strings.Split(body, ";") {
		eq := strings.IndexByte(kv, '=')
		if eq < 0 {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(kv[:eq]))
		val := strings.TrimSpace(kv[eq+1:])
		switch key {
		case "UNTIL":
			t, perr := parseUntil(val)
			if perr != nil {
				return time.Time{}, 0, false, false, perr
			}
			until, hasUntil = t, true
		case "COUNT":
			n, perr := strconv.Atoi(val)
			if perr != nil {
				return time.Time{}, 0, false, false, perr
			}
			count, hasCount = n, true
		}
	}
	return
}

func parseUntil(s string) (time.Time, error) {
	upper := strings.ToUpper(strings.TrimSpace(s))
	if t, err := time.Parse("20060102T150405Z", upper); err == nil {
		return t, nil
	}
	if t, err := time.Parse("20060102", upper); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("invalid UNTIL value: " + s)
}
```

- [ ] **Step 4: Run to verify pass**

Run: `go test ./services/calendar-service/internal/event/ -run TestParseRRULE`
Expected: PASS — all cases green.

- [ ] **Step 5: Commit**

```bash
git add services/calendar-service/internal/event/recurrence_validator.go services/calendar-service/internal/event/recurrence_validator_test.go
git commit -m "feat(calendar-service): scaffold parseRRULE for UNTIL/COUNT extraction"
```

---

### Task 10: `ValidateRecurrence` covering all three error codes

**Files:**
- Modify: `services/calendar-service/internal/event/recurrence_validator.go`
- Modify: `services/calendar-service/internal/event/recurrence_validator_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `recurrence_validator_test.go`:

```go
func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("bad time %q: %v", s, err)
	}
	return v
}

func TestValidateRecurrence(t *testing.T) {
	start := mustTime(t, "2026-05-06T09:00:00Z")

	tests := []struct {
		name    string
		input   []string
		start   time.Time
		wantNil bool
		wantCode string
	}{
		{name: "nil slice", input: nil, start: start, wantNil: true},
		{name: "empty slice", input: []string{}, start: start, wantNil: true},
		{name: "valid until", input: []string{"RRULE:FREQ=WEEKLY;UNTIL=20260611T035959Z"}, start: start, wantNil: true},
		{name: "valid count", input: []string{"RRULE:FREQ=DAILY;COUNT=5"}, start: start, wantNil: true},
		{name: "open-ended", input: []string{"RRULE:FREQ=WEEKLY"}, start: start, wantCode: codeUnbounded},
		{name: "until > 5y", input: []string{"RRULE:FREQ=WEEKLY;UNTIL=20320101T000000Z"}, start: start, wantCode: codeTooLong},
		{name: "count zero", input: []string{"RRULE:FREQ=WEEKLY;COUNT=0"}, start: start, wantCode: codeCountRange},
		{name: "count 731", input: []string{"RRULE:FREQ=WEEKLY;COUNT=731"}, start: start, wantCode: codeCountRange},
		{name: "case-insensitive", input: []string{"rrule:freq=weekly;until=20260601t000000z"}, start: start, wantNil: true},
		{name: "until date-only ok", input: []string{"RRULE:FREQ=WEEKLY;UNTIL=20260601"}, start: start, wantNil: true},
		{name: "EXDATE alongside RRULE only", input: []string{"EXDATE:20260513T090000Z", "RRULE:FREQ=WEEKLY;COUNT=5"}, start: start, wantNil: true},
		{name: "malformed UNTIL -> unbounded", input: []string{"RRULE:FREQ=WEEKLY;UNTIL=garbage"}, start: start, wantCode: codeUnbounded},
		{name: "zero start skips too-long check", input: []string{"RRULE:FREQ=WEEKLY;UNTIL=20990101T000000Z"}, start: time.Time{}, wantNil: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRecurrence(tc.input, tc.start)
			if tc.wantNil {
				if err != nil {
					t.Fatalf("expected nil, got %+v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error code %q, got nil", tc.wantCode)
			}
			if err.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", err.Code, tc.wantCode)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify fail**

Run: `go test ./services/calendar-service/internal/event/ -run TestValidateRecurrence`
Expected: FAIL — `ValidateRecurrence` undefined.

- [ ] **Step 3: Implement `ValidateRecurrence`**

Append to `services/calendar-service/internal/event/recurrence_validator.go`:

```go
// ValidateRecurrence enforces the §4.6 PRD checks on every "RRULE:" entry of
// the slice. Non-RRULE components (EXDATE, RDATE, etc.) are ignored. Returns
// the first failure, or nil if every RRULE line is bounded and within range.
// If eventStart.IsZero(), the 5-year window check is skipped (the count and
// unbounded checks still run); this is used by the update handler where the
// start time may not be supplied.
func ValidateRecurrence(recurrence []string, eventStart time.Time) *RecurrenceError {
	for _, line := range recurrence {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToUpper(trimmed), "RRULE:") {
			continue
		}
		until, count, hasUntil, hasCount, err := parseRRULE(trimmed)
		if err != nil || (!hasUntil && !hasCount) {
			return &RecurrenceError{
				Code:    codeUnbounded,
				Detail:  "Recurring events must specify an end date (UNTIL=) or occurrence count (COUNT=)",
				RuleRaw: trimmed,
			}
		}
		if hasCount && (count < minOccurrences || count > maxOccurrences) {
			return &RecurrenceError{
				Code:    codeCountRange,
				Detail:  "COUNT must be between 1 and 730",
				RuleRaw: trimmed,
			}
		}
		if hasUntil && !eventStart.IsZero() {
			if until.Sub(eventStart) > maxUntilWindow {
				return &RecurrenceError{
					Code:    codeTooLong,
					Detail:  "UNTIL must be no more than 5 years after the event start",
					RuleRaw: trimmed,
				}
			}
		}
	}
	return nil
}
```

- [ ] **Step 4: Run to verify pass**

Run: `go test ./services/calendar-service/internal/event/`
Expected: PASS — all validator tests green; existing event-package tests unchanged.

- [ ] **Step 5: Commit**

```bash
git add services/calendar-service/internal/event/recurrence_validator.go services/calendar-service/internal/event/recurrence_validator_test.go
git commit -m "feat(calendar-service): add ValidateRecurrence safety net"
```

---

## Phase 5 — Backend wiring

### Task 11: Add `Recurrence` field to `UpdateEventRequest`

**Files:**
- Modify: `services/calendar-service/internal/event/rest.go`

- [ ] **Step 1: Modify the struct**

Edit `services/calendar-service/internal/event/rest.go`. Replace the `UpdateEventRequest` declaration (currently lines 35-45):

```go
type UpdateEventRequest struct {
	Id          uuid.UUID `json:"-"`
	Title       *string   `json:"title"`
	Start       *string   `json:"start"`
	End         *string   `json:"end"`
	AllDay      *bool     `json:"allDay"`
	Location    *string   `json:"location"`
	Description *string   `json:"description"`
	Scope       string    `json:"scope"`
	TimeZone    *string   `json:"timeZone"`
}
```

with:

```go
type UpdateEventRequest struct {
	Id          uuid.UUID `json:"-"`
	Title       *string   `json:"title"`
	Start       *string   `json:"start"`
	End         *string   `json:"end"`
	AllDay      *bool     `json:"allDay"`
	Location    *string   `json:"location"`
	Description *string   `json:"description"`
	Scope       string    `json:"scope"`
	TimeZone    *string   `json:"timeZone"`
	Recurrence  *[]string `json:"recurrence"`
}
```

- [ ] **Step 2: Build to verify it compiles**

Run: `go build ./services/calendar-service/...`
Expected: PASS — no compilation errors. The processor does not read this field, so no other code needs to change.

- [ ] **Step 3: Run all calendar-service tests**

Run: `go test ./services/calendar-service/...`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add services/calendar-service/internal/event/rest.go
git commit -m "feat(calendar-service): allow Recurrence on UpdateEventRequest for boundary validation"
```

---

### Task 12: Wire validator into `createEventHandler`

**Files:**
- Modify: `services/calendar-service/internal/event/resource.go`

- [ ] **Step 1: Add the helper and call**

Edit `services/calendar-service/internal/event/resource.go`. Just above `func handleMutationError(...)` near the end of the file, add a private helper:

```go
func validateRecurrenceOrWriteError(d *server.HandlerDependency, w http.ResponseWriter, recurrence []string, eventStart time.Time, connID uuid.UUID) bool {
	if rerr := ValidateRecurrence(recurrence, eventStart); rerr != nil {
		d.Logger().WithFields(logrus.Fields{
			"connID": connID,
			"rule":   rerr.RuleRaw,
		}).Info("rejected recurrence")
		server.WriteJSONAPIError(w, http.StatusUnprocessableEntity, rerr.Code, "Validation Error", rerr.Detail, "")
		return false
	}
	return true
}
```

In `createEventHandler`, immediately after the `if input.Title == "" { ... return }` block (currently around line 109-112), insert:

```go
var eventStart time.Time
if input.Start != "" {
	if t, perr := time.Parse(time.RFC3339, input.Start); perr == nil {
		eventStart = t
	}
}
if !validateRecurrenceOrWriteError(d, w, input.Recurrence, eventStart, connID) {
	return
}
```

- [ ] **Step 2: Write the failing test**

Append to `services/calendar-service/internal/event/recurrence_validator_test.go` a quick smoke test that exercises the helper indirectly via `ValidateRecurrence` with the same payload the handler would produce:

```go
func TestValidateRecurrence_HandlerScenarios(t *testing.T) {
	start := mustTime(t, "2026-05-06T09:00:00-04:00")

	cases := []struct {
		name string
		rule string
		want string
	}{
		{"weekly may6 to jun10 inclusive", "RRULE:FREQ=WEEKLY;UNTIL=20260611T035959Z", ""},
		{"daily count 5", "RRULE:FREQ=DAILY;COUNT=5", ""},
		{"open-ended weekly", "RRULE:FREQ=WEEKLY", codeUnbounded},
		{"until 5y+2d", "RRULE:FREQ=WEEKLY;UNTIL=20310509T000000Z", codeTooLong},
		{"count 731", "RRULE:FREQ=WEEKLY;COUNT=731", codeCountRange},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRecurrence([]string{tc.rule}, start)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("expected nil, got %+v", err)
				}
				return
			}
			if err == nil || err.Code != tc.want {
				t.Fatalf("got %+v, want code %q", err, tc.want)
			}
		})
	}
}
```

- [ ] **Step 3: Build & run all tests**

Run: `go build ./services/calendar-service/... && go test ./services/calendar-service/...`
Expected: PASS — every package compiles; validator test cases all green.

- [ ] **Step 4: Commit**

```bash
git add services/calendar-service/internal/event/resource.go services/calendar-service/internal/event/recurrence_validator_test.go
git commit -m "feat(calendar-service): reject open-ended RRULEs on event create"
```

---

### Task 13: Wire validator into `updateEventHandler`

**Files:**
- Modify: `services/calendar-service/internal/event/resource.go`

- [ ] **Step 1: Add the call in the update branch**

In `updateEventHandler`, immediately after the existing `eventID, err := uuid.Parse(eventIDStr)` block (just before `connProc := connection.NewProcessor(...)`), insert:

```go
if input.Recurrence != nil {
	var eventStart time.Time
	if input.Start != nil && *input.Start != "" {
		if t, perr := time.Parse(time.RFC3339, *input.Start); perr == nil {
			eventStart = t
		}
	}
	if !validateRecurrenceOrWriteError(d, w, *input.Recurrence, eventStart, connID) {
		return
	}
}
```

- [ ] **Step 2: Build & run all tests**

Run: `go build ./services/calendar-service/... && go test ./services/calendar-service/...`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add services/calendar-service/internal/event/resource.go
git commit -m "feat(calendar-service): apply recurrence safety net on event update"
```

---

## Phase 6 — Final verification

### Task 14: Whole-stack verification

**Files:** none (verification only)

- [ ] **Step 1: Frontend build + tests**

Run: `npm --prefix frontend run build && npm --prefix frontend test`
Expected: PASS — TypeScript clean; all tests green (recurrence + schema + form-dialog + pre-existing).

- [ ] **Step 2: Backend tests**

Run: `go test ./services/calendar-service/...`
Expected: PASS — every test in the service.

- [ ] **Step 3: Local Docker build smoke**

Run: `docker build -f services/calendar-service/Dockerfile .`
Expected: PASS — Docker build completes for `calendar-service`. (Per `CLAUDE.md`: always verify Docker builds when changing shared libraries; this task does not touch shared libraries, so this step is just a defensive check.)

- [ ] **Step 4: Acceptance walkthrough (manual, no commit)**

Tick through `prd.md` §10 acceptance criteria against the implemented behavior. Anything failing means a missed task — go back, do not paper over.

- [ ] **Step 5: No commit**

Verification only.

---

## Self-Review Outcome

**Spec coverage** — every PRD §10 acceptance criterion maps to at least one task:
- §10.1 (Ends control appears for recurring) → Task 5 + Task 8 test "hides..." / "shows..."
- §10.2 (default On + start+1y) → Task 6 + Task 8 test "seeds end date"
- §10.3 (auto-update vs user-edited) → Task 6 + Task 8 tests "auto-updates" / "leaves a user-edited..."
- §10.4 (Never warning + checkbox) → Task 5 + Task 8 test "blocks submit until..."
- §10.5–§10.7 (validation messages) → Task 4 schema refinements
- §10.8 (Weekly May 6 → June 10 finite) → Task 8 test "submits an UNTIL-terminated RRULE"
- §10.9 (Daily After 5) → Task 8 test "submits a COUNT-terminated RRULE"
- §10.10 (UTC conversion for non-UTC zones) → Task 2 + Task 3
- §10.11–§10.13 (backend 422 codes) → Task 10 validator tests + Task 12 + Task 13 wiring
- §10.14 (FE composition coverage) → Task 3
- §10.15 (BE handler tests for 422 paths) → Task 10 + Task 12's TestValidateRecurrence_HandlerScenarios (per design §3.2 note, no httptest harness exists; validator-level coverage is the substitute)
- §10.16 (no migration) → no migration tasks
- §10.17 (edit dialog unchanged) → Task 5 explicitly gates new UI on `!isEdit`

**Placeholder scan** — no TBDs, no "implement later", every code step has the actual code, every command has expected output.

**Type consistency** — `composeRecurrenceRule` signature in Task 3 matches the call in Task 7; `ValidateRecurrence` signature in Task 10 matches the calls in Task 12 / Task 13; `RecurrenceError` fields used in `validateRecurrenceOrWriteError` (Task 12) match the struct definition in Task 9.
