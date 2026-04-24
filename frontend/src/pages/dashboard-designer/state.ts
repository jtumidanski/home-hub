import type { Dashboard } from "@/types/models/dashboard";
import type { Layout, WidgetInstance } from "@/lib/dashboard/schema";

export type DraftState = {
  name: string;
  layout: Layout;
  dirty: boolean;
  selectedWidgetId: string | null;
  paletteOpen: boolean;
};

export type DraftAction =
  | { type: "rename"; name: string }
  | { type: "move-or-resize"; widgets: WidgetInstance[] }
  | { type: "add"; widget: WidgetInstance }
  | { type: "remove"; id: string }
  | { type: "update-config"; id: string; config: Record<string, unknown> }
  | { type: "select"; id: string | null }
  | { type: "toggle-palette"; open: boolean }
  | { type: "reset"; server: Dashboard }
  | { type: "saved"; server: Dashboard };

export function fromServer(server: Dashboard): DraftState {
  return {
    name: server.attributes.name,
    layout: server.attributes.layout,
    dirty: false,
    selectedWidgetId: null,
    paletteOpen: false,
  };
}

export function draftReducer(state: DraftState, action: DraftAction): DraftState {
  switch (action.type) {
    case "rename":
      return { ...state, name: action.name, dirty: true };
    case "move-or-resize":
      return { ...state, layout: { ...state.layout, widgets: action.widgets }, dirty: true };
    case "add":
      return {
        ...state,
        layout: { ...state.layout, widgets: [...state.layout.widgets, action.widget] },
        dirty: true,
        selectedWidgetId: action.widget.id,
      };
    case "remove":
      return {
        ...state,
        layout: {
          ...state.layout,
          widgets: state.layout.widgets.filter((w) => w.id !== action.id),
        },
        dirty: true,
        selectedWidgetId: state.selectedWidgetId === action.id ? null : state.selectedWidgetId,
      };
    case "update-config":
      return {
        ...state,
        layout: {
          ...state.layout,
          widgets: state.layout.widgets.map((w) =>
            w.id === action.id ? { ...w, config: action.config } : w,
          ),
        },
        dirty: true,
      };
    case "select":
      return { ...state, selectedWidgetId: action.id };
    case "toggle-palette":
      return { ...state, paletteOpen: action.open };
    case "reset":
      return fromServer(action.server);
    case "saved":
      return fromServer(action.server);
  }
}
