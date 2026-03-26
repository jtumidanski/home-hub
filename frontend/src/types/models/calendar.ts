export interface CalendarConnectionAttributes {
  provider: string;
  status: "connected" | "disconnected" | "syncing" | "error";
  email: string;
  userDisplayName: string;
  userColor: string;
  lastSyncAt: string | null;
  lastSyncEventCount: number;
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
}

export interface CalendarEvent {
  id: string;
  type: "calendar-events";
  attributes: CalendarEventAttributes;
}

export interface AuthorizeResponseAttributes {
  authorizeUrl: string;
}

export interface AuthorizeResponse {
  id: string;
  type: "calendar-authorization-responses";
  attributes: AuthorizeResponseAttributes;
}
