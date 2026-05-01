export interface CalendarConnectionAttributes {
  provider: string;
  status: "connected" | "disconnected" | "syncing" | "error";
  email: string;
  userDisplayName: string;
  userColor: string;
  writeAccess: boolean;
  lastSyncAt: string | null;
  lastSyncAttemptAt: string | null;
  lastSyncEventCount: number;
  errorCode: string | null;
  errorMessage: string | null;
  lastErrorAt: string | null;
  consecutiveFailures: number;
  createdAt: string;
}

export interface CalendarConnection {
  id: string;
  type: "calendar-connections";
  attributes: CalendarConnectionAttributes;
}

export interface CalendarSourceAttributes {
  name: string;
  primary: boolean;
  visible: boolean;
  color: string;
}

export interface CalendarSource {
  id: string;
  type: "calendar-sources";
  attributes: CalendarSourceAttributes;
}

export interface CalendarEventAttributes {
  title: string;
  description: string | null;
  startTime: string;
  endTime: string;
  allDay: boolean;
  location: string | null;
  visibility: string;
  userDisplayName: string;
  userColor: string;
  isOwner: boolean;
  sourceId: string;
  connectionId: string;
  isRecurring: boolean;
}

export interface CalendarEvent {
  id: string;
  type: "calendar-events";
  attributes: CalendarEventAttributes;
}

export interface CreateEventData {
  title: string;
  start: string;
  end: string;
  allDay: boolean;
  location?: string | undefined;
  description?: string | undefined;
  recurrence?: string[] | undefined;
  timeZone?: string | undefined;
}

export interface UpdateEventData {
  title?: string | undefined;
  start?: string | undefined;
  end?: string | undefined;
  allDay?: boolean | undefined;
  location?: string | undefined;
  description?: string | undefined;
  scope?: "single" | "all" | undefined;
  timeZone?: string | undefined;
}

export interface AuthorizeResponseAttributes {
  authorizeUrl: string;
}

export interface AuthorizeResponse {
  id: string;
  type: "calendar-authorization-responses";
  attributes: AuthorizeResponseAttributes;
}
