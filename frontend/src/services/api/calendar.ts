import { BaseService } from "./base";
import type {
  CalendarConnection,
  CalendarSource,
  CalendarEvent,
  CreateEventData,
  UpdateEventData,
  AuthorizeResponse,
} from "@/types/models/calendar";

class CalendarService extends BaseService {
  constructor() {
    super("/calendar");
  }

  getConnections(tenant: { id: string }) {
    return this.getList<CalendarConnection>(tenant, "/calendar/connections");
  }

  authorizeGoogle(tenant: { id: string }, redirectUri: string) {
    return this.create<AuthorizeResponse>(tenant, "/calendar/connections/google/authorize", {
      data: {
        type: "calendar-authorization-requests",
        attributes: { redirectUri },
      },
    });
  }

  deleteConnection(tenant: { id: string }, id: string) {
    return this.remove(tenant, `/calendar/connections/${id}`);
  }

  getCalendarSources(tenant: { id: string }, connectionId: string) {
    return this.getList<CalendarSource>(tenant, `/calendar/connections/${connectionId}/calendars`);
  }

  toggleCalendarSource(tenant: { id: string }, connectionId: string, calId: string, visible: boolean) {
    return this.update<CalendarSource>(tenant, `/calendar/connections/${connectionId}/calendars/${calId}`, {
      data: {
        type: "calendar-sources",
        id: calId,
        attributes: { visible },
      },
    });
  }

  triggerSync(tenant: { id: string }, connectionId: string) {
    return this.create<CalendarConnection>(tenant, `/calendar/connections/${connectionId}/sync`, {});
  }

  getEvents(tenant: { id: string }, start: string, end: string) {
    return this.getList<CalendarEvent>(tenant, `/calendar/events?start=${encodeURIComponent(start)}&end=${encodeURIComponent(end)}`);
  }

  reauthorizeGoogle(tenant: { id: string }, redirectUri: string) {
    return this.create<AuthorizeResponse>(tenant, "/calendar/connections/google/authorize", {
      data: {
        type: "calendar-authorization-requests",
        attributes: { redirectUri, reauthorize: true },
      },
    });
  }

  createEvent(tenant: { id: string }, connectionId: string, calendarId: string, data: CreateEventData) {
    return this.create<CalendarEvent>(tenant, `/calendar/connections/${connectionId}/calendars/${calendarId}/events`, {
      data: {
        type: "calendar-events",
        attributes: data,
      },
    });
  }

  updateEvent(tenant: { id: string }, connectionId: string, eventId: string, data: UpdateEventData) {
    return this.update<CalendarEvent>(tenant, `/calendar/connections/${connectionId}/events/${eventId}`, {
      data: {
        type: "calendar-events",
        id: eventId,
        attributes: data,
      },
    });
  }

  deleteEvent(tenant: { id: string }, connectionId: string, eventId: string, scope: "single" | "all" = "single") {
    return this.remove(tenant, `/calendar/connections/${connectionId}/events/${eventId}?scope=${scope}`);
  }
}

export const calendarService = new CalendarService();
