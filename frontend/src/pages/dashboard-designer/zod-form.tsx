import { Fragment, type ReactNode } from "react";
import type { UseFormReturn } from "react-hook-form";
import { Controller } from "react-hook-form";
import { z } from "zod";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { cn, toTitleCase } from "@/lib/utils";

const inputClass =
  "h-8 w-full min-w-0 rounded-lg border border-input bg-transparent px-2.5 py-1 text-base outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 md:text-sm";

/**
 * Unwraps Optional / Default / Nullable / Readonly wrappers to reach the
 * underlying schema used for input rendering.
 */
function unwrap(schema: unknown): { inner: unknown; optional: boolean; defaultValue?: unknown } {
  let s = schema as { _def?: { type?: string; innerType?: unknown; defaultValue?: unknown } };
  let optional = false;
  let defaultValue: unknown = undefined;
  const guard = new Set<unknown>();
  while (s && !guard.has(s)) {
    guard.add(s);
    const t = s._def?.type;
    if (t === "optional" || t === "nullable") {
      optional = true;
      s = s._def?.innerType as typeof s;
    } else if (t === "default") {
      defaultValue = s._def?.defaultValue;
      s = s._def?.innerType as typeof s;
    } else if (t === "readonly") {
      s = s._def?.innerType as typeof s;
    } else {
      break;
    }
  }
  return { inner: s, optional, defaultValue };
}

function isZodObject(x: unknown): x is z.ZodObject<Record<string, z.ZodTypeAny>> {
  return (
    typeof x === "object" &&
    x !== null &&
    (x as { _def?: { type?: string } })._def?.type === "object" &&
    typeof (x as { shape?: unknown }).shape === "object"
  );
}

function zodType(x: unknown): string | undefined {
  return (x as { _def?: { type?: string } } | undefined)?._def?.type;
}

function getStringMaxLength(schema: unknown): number | undefined {
  const checks =
    ((schema as { _def?: { checks?: Array<{ _zod?: { def?: { check?: string; maximum?: number } } }> } }).
      _def?.checks) ?? [];
  for (const c of checks) {
    if (c?._zod?.def?.check === "max_length") return c._zod.def.maximum;
  }
  return undefined;
}

function getNumberBounds(schema: unknown): { min?: number; max?: number; int?: boolean } {
  const checks =
    ((schema as {
      _def?: {
        checks?: Array<{
          _zod?: { def?: { check?: string; value?: number; inclusive?: boolean; format?: string } };
        }>;
      };
    }).
      _def?.checks) ?? [];
  const out: { min?: number; max?: number; int?: boolean } = {};
  for (const c of checks) {
    const def = c?._zod?.def;
    if (!def) continue;
    if (def.check === "greater_than" && typeof def.value === "number") {
      out.min = def.inclusive ? def.value : def.value + 1;
    } else if (def.check === "less_than" && typeof def.value === "number") {
      out.max = def.inclusive ? def.value : def.value - 1;
    } else if (def.check === "number_format" && def.format?.includes("int")) {
      out.int = true;
    }
  }
  return out;
}

interface ZodFormFieldsProps<T extends Record<string, unknown>> {
  schema: z.ZodType<T>;
  form: UseFormReturn<T>;
  /** Dot-path prefix used for nested fieldsets */
  path?: string;
}

/**
 * Renders labeled inputs for every property in a zod object schema using
 * the supplied `react-hook-form` instance. Nested objects recurse into
 * `<fieldset>`s; unsupported branches fall back to a JSON textarea.
 */
export function ZodFormFields<T extends Record<string, unknown>>({
  schema,
  form,
  path = "",
}: ZodFormFieldsProps<T>) {
  if (!isZodObject(schema)) {
    return (
      <pre className="rounded-md bg-muted p-2 text-xs">
        Unsupported schema: {zodType(schema) ?? "unknown"}
      </pre>
    );
  }
  const shape = schema.shape;
  const keys = Object.keys(shape);
  return (
    <Fragment>
      {keys.map((key) => {
        const field = shape[key] as unknown;
        const fullPath = path ? `${path}.${key}` : key;
        return (
          <div key={fullPath} className="flex flex-col gap-1">
            {renderField({ name: fullPath, label: toTitleCase(key), schema: field, form })}
          </div>
        );
      })}
    </Fragment>
  );
}

interface RenderFieldArgs<T extends Record<string, unknown>> {
  name: string;
  label: string;
  schema: unknown;
  form: UseFormReturn<T>;
}

function renderField<T extends Record<string, unknown>>({
  name,
  label,
  schema,
  form,
}: RenderFieldArgs<T>): ReactNode {
  const { inner } = unwrap(schema);
  const t = zodType(inner);

  if (t === "object" && isZodObject(inner)) {
    return (
      <fieldset className="rounded-md border border-border p-2">
        <legend className="px-1 text-xs font-medium text-muted-foreground">{label}</legend>
        <div className="flex flex-col gap-2">
          <ZodFormFields schema={inner as z.ZodType<Record<string, unknown>>} form={form as unknown as UseFormReturn<Record<string, unknown>>} path={name} />
        </div>
      </fieldset>
    );
  }

  if (t === "enum") {
    const options =
      ((inner as { options?: string[] }).options) ??
      Object.values(
        ((inner as { _def?: { entries?: Record<string, string> } })._def?.entries) ?? {},
      );
    return (
      <Controller
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        name={name as any}
        control={form.control}
        render={({ field }) => (
          <div className="flex flex-col gap-1" data-testid={`zod-field-${name}`}>
            <Label>{label}</Label>
            <RadioGroup
              value={String(field.value ?? "")}
              onValueChange={(v) => field.onChange(v)}
            >
              {options.map((opt: string) => (
                <RadioGroupItem key={opt} value={opt}>
                  {toTitleCase(opt)}
                </RadioGroupItem>
              ))}
            </RadioGroup>
          </div>
        )}
      />
    );
  }

  if (t === "boolean") {
    return (
      <Controller
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        name={name as any}
        control={form.control}
        render={({ field }) => (
          <label className="flex cursor-pointer items-center gap-2">
            <input
              type="checkbox"
              checked={!!field.value}
              onChange={(e) => field.onChange(e.target.checked)}
              className="h-4 w-4"
              data-testid={`zod-field-${name}`}
              aria-label={label}
            />
            <span className="text-sm">{label}</span>
          </label>
        )}
      />
    );
  }

  if (t === "number") {
    const { min, max, int } = getNumberBounds(inner);
    return (
      <Controller
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        name={name as any}
        control={form.control}
        render={({ field }) => (
          <div className="flex flex-col gap-1">
            <Label>{label}</Label>
            <input
              type="number"
              className={cn(inputClass)}
              min={min}
              max={max}
              step={int ? 1 : undefined}
              value={field.value == null ? "" : String(field.value)}
              onChange={(e) => {
                const v = e.target.value;
                if (v === "") {
                  field.onChange(undefined);
                } else {
                  const n = int ? parseInt(v, 10) : parseFloat(v);
                  field.onChange(Number.isNaN(n) ? undefined : n);
                }
              }}
              data-testid={`zod-field-${name}`}
            />
          </div>
        )}
      />
    );
  }

  if (t === "string") {
    const max = getStringMaxLength(inner);
    return (
      <Controller
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        name={name as any}
        control={form.control}
        render={({ field }) => {
          const value = typeof field.value === "string" ? field.value : "";
          return (
            <div className="flex flex-col gap-1">
              <Label>{label}</Label>
              <input
                type="text"
                className={cn(inputClass)}
                maxLength={max}
                value={value}
                onChange={(e) => field.onChange(e.target.value)}
                data-testid={`zod-field-${name}`}
              />
              {max ? (
                <span className="text-xs text-muted-foreground">
                  {value.length} / {max}
                </span>
              ) : null}
            </div>
          );
        }}
      />
    );
  }

  // Fallback: raw JSON textarea.
  return (
    <Controller
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      name={name as any}
      control={form.control}
      render={({ field }) => (
        <div className="flex flex-col gap-1">
          <Label>{label}</Label>
          <Textarea
            value={
              typeof field.value === "string" ? field.value : JSON.stringify(field.value ?? "")
            }
            onChange={(e) => {
              const v = e.target.value;
              try {
                field.onChange(JSON.parse(v));
              } catch {
                field.onChange(v);
              }
            }}
            data-testid={`zod-field-${name}`}
          />
          <span className="text-xs text-muted-foreground">
            Unsupported field type ({zodType(inner) ?? "unknown"}); edit as JSON.
          </span>
        </div>
      )}
    />
  );
}
