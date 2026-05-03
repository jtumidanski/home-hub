import { useEffect, useMemo, useRef } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { useCreateEvent, useUpdateEvent } from "@/lib/hooks/api/use-calendar";
import { createErrorFromUnknown } from "@/lib/api/errors";
import {
  eventFormSchema,
  type EventFormData,
  createEventDefaults,
  RECURRENCE_OPTIONS,
} from "@/lib/schemas/calendar-event.schema";
import type { CalendarEvent, CalendarConnection, CalendarSource } from "@/types/models/calendar";

interface EventFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  connections: CalendarConnection[];
  sources: CalendarSource[];
  editEvent?: CalendarEvent | null | undefined;
  editScope?: "single" | "all" | undefined;
  prefilledStart?: Date | undefined;
}

export function EventFormDialog({
  open,
  onOpenChange,
  connections,
  sources,
  editEvent,
  editScope,
  prefilledStart,
}: EventFormDialogProps) {
  const isEdit = !!editEvent;
  const createEvent = useCreateEvent();
  const updateEvent = useUpdateEvent();

  const writeConnections = useMemo(
    () => connections.filter((c) => c.attributes.writeAccess && c.attributes.status === "connected"),
    [connections],
  );
  const defaultConnection = writeConnections[0];

  const defaults = useMemo(() => {
    if (editEvent) {
      const attrs = editEvent.attributes;
      const start = new Date(attrs.startTime);
      const end = new Date(attrs.endTime);
      return {
        title: attrs.title,
        allDay: attrs.allDay,
        startDate: start.toISOString().slice(0, 10),
        startTime: `${String(start.getHours()).padStart(2, "0")}:${String(start.getMinutes()).padStart(2, "0")}`,
        endDate: end.toISOString().slice(0, 10),
        endTime: `${String(end.getHours()).padStart(2, "0")}:${String(end.getMinutes()).padStart(2, "0")}`,
        recurrence: "",
        location: attrs.location ?? "",
        description: attrs.description ?? "",
        calendarId: attrs.sourceId,
        connectionId: attrs.connectionId,
        endsMode: "on" as const,
        endsOnDate: "",
        endsAfterCount: 10,
        endsNeverConfirmed: false,
        endsOnDateUserEdited: false,
      };
    }
    return createEventDefaults(prefilledStart);
  }, [editEvent, prefilledStart]);

  const form = useForm<EventFormData>({
    resolver: zodResolver(eventFormSchema),
    defaultValues: defaults,
  });

  useEffect(() => {
    if (open) {
      form.reset(defaults);
      if (!isEdit && defaultConnection && sources.length > 0) {
        const primary = sources.find((s) => s.attributes.primary);
        form.setValue("connectionId", defaultConnection.id);
        form.setValue("calendarId", primary?.id ?? sources[0]!.id);
      }
    }
  }, [open, defaults, isEdit, defaultConnection, sources, form]);

  const previousRecurrenceRef = useRef(defaults.recurrence);
  // eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
  const allDay = form.watch("allDay");
  // eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
  const recurrence = form.watch("recurrence");
  // eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
  const endsMode = form.watch("endsMode");
  // eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
  const startDate = form.watch("startDate");
  // eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
  const endsOnDateUserEdited = form.watch("endsOnDateUserEdited");

  function addOneYear(yyyymmdd: string): string {
    if (!yyyymmdd) return "";
    const [yStr, mStr, dStr] = yyyymmdd.split("-");
    const y = Number(yStr);
    const m = Number(mStr);
    const d = Number(dStr);
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

  const handleOpenChange = (next: boolean) => {
    if (form.formState.isSubmitting) return;
    onOpenChange(next);
  };

  const onSubmit = async (values: EventFormData) => {
    try {
      const toISO = (date: string, time: string) => {
        const d = new Date(`${date}T${time}`);
        const off = -d.getTimezoneOffset();
        const sign = off >= 0 ? "+" : "-";
        const hh = String(Math.floor(Math.abs(off) / 60)).padStart(2, "0");
        const mm = String(Math.abs(off) % 60).padStart(2, "0");
        return `${date}T${time}:00${sign}${hh}:${mm}`;
      };

      const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

      if (isEdit && editEvent) {
        const startISO = values.allDay
          ? values.startDate
          : toISO(values.startDate, values.startTime);
        const endISO = values.allDay
          ? values.endDate
          : toISO(values.endDate, values.endTime);

        await updateEvent.mutateAsync({
          connectionId: editEvent.attributes.connectionId,
          eventId: editEvent.id,
          data: {
            title: values.title,
            start: startISO,
            end: endISO,
            allDay: values.allDay,
            location: values.location || undefined,
            description: values.description || undefined,
            scope: editScope ?? "single",
            timeZone,
          },
        });
        toast.success("Event updated");
      } else {
        const startISO = values.allDay
          ? values.startDate
          : toISO(values.startDate, values.startTime);
        const endISO = values.allDay
          ? values.endDate
          : toISO(values.endDate, values.endTime);

        await createEvent.mutateAsync({
          connectionId: values.connectionId,
          calendarId: values.calendarId,
          data: {
            title: values.title,
            start: startISO,
            end: endISO,
            allDay: values.allDay,
            location: values.location || undefined,
            description: values.description || undefined,
            recurrence: values.recurrence ? [values.recurrence] : undefined,
            timeZone,
          },
        });
        toast.success("Event created");
      }
      onOpenChange(false);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, `Failed to ${isEdit ? "update" : "create"} event`).message);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Event" : "Create Event"}</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="title"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Title</FormLabel>
                  <FormControl>
                    <Input placeholder="Event title" autoFocus {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="allDay"
              render={({ field }) => (
                <FormItem className="flex items-center gap-2">
                  <FormControl>
                    <input
                      type="checkbox"
                      checked={field.value}
                      onChange={field.onChange}
                      className="h-4 w-4 rounded border-gray-300"
                    />
                  </FormControl>
                  <FormLabel className="!mt-0">All day</FormLabel>
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-3">
              <FormField
                control={form.control}
                name="startDate"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Start date</FormLabel>
                    <FormControl>
                      <Input type="date" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {!allDay && (
                <FormField
                  control={form.control}
                  name="startTime"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Start time</FormLabel>
                      <FormControl>
                        <Input type="time" {...field} />
                      </FormControl>
                    </FormItem>
                  )}
                />
              )}
            </div>

            <div className="grid grid-cols-2 gap-3">
              <FormField
                control={form.control}
                name="endDate"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>End date</FormLabel>
                    <FormControl>
                      <Input type="date" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {!allDay && (
                <FormField
                  control={form.control}
                  name="endTime"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>End time</FormLabel>
                      <FormControl>
                        <Input type="time" {...field} />
                      </FormControl>
                    </FormItem>
                  )}
                />
              )}
            </div>

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

            <FormField
              control={form.control}
              name="location"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Location</FormLabel>
                  <FormControl>
                    <Input placeholder="Add location" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea placeholder="Add description" rows={3} {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {!isEdit && sources.length > 1 && (
              <FormField
                control={form.control}
                name="calendarId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Calendar</FormLabel>
                    <FormControl>
                      <select
                        value={field.value}
                        onChange={field.onChange}
                        className="flex h-8 w-full rounded-lg border border-input bg-popover text-popover-foreground px-2.5 py-1.5 text-sm"
                      >
                        {sources.map((src) => (
                          <option key={src.id} value={src.id}>
                            {src.attributes.name}
                            {src.attributes.primary ? " (primary)" : ""}
                          </option>
                        ))}
                      </select>
                    </FormControl>
                  </FormItem>
                )}
              />
            )}

            <Button type="submit" className="w-full" disabled={form.formState.isSubmitting}>
              {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {isEdit ? "Save Changes" : "Create Event"}
            </Button>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
