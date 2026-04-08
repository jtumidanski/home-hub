export interface TaskAttributes {
  title: string;
  notes?: string;
  status: "pending" | "completed";
  dueOn?: string;
  rolloverEnabled: boolean;
  ownerUserId?: string | null;
  completedAt?: string;
  deletedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Task {
  id: string;
  type: "tasks";
  attributes: TaskAttributes;
}

// --- Create attributes (F14) ---

export interface TaskCreateAttributes {
  title: string;
  notes?: string;
  dueOn?: string;
  rolloverEnabled?: boolean;
  ownerUserId?: string | null;
}

// --- Update attributes (F14) ---

export type TaskUpdateAttributes = Partial<
  Pick<TaskAttributes, "title" | "notes" | "status" | "dueOn" | "rolloverEnabled" | "ownerUserId">
>;

// --- Label map (F15) ---

export const taskStatusLabelMap: Record<TaskAttributes["status"], string> = {
  pending: "Pending",
  completed: "Completed",
};

// --- Helpers (F16) ---

export function isTaskOverdue(task: Task): boolean {
  const { status, dueOn } = task.attributes;
  if (status !== "pending" || !dueOn) {
    return false;
  }
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(dueOn);
  if (!match) {
    return false;
  }
  const [, yearStr, monthStr, dayStr] = match;
  const dueDate = new Date(Number(yearStr), Number(monthStr) - 1, Number(dayStr));
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return dueDate < today;
}

export function isTaskCompleted(task: Task): boolean {
  return task.attributes.status === "completed";
}
