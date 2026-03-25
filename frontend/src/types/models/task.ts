export interface TaskAttributes {
  title: string;
  notes?: string;
  status: "pending" | "completed";
  dueOn?: string;
  rolloverEnabled: boolean;
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
}

// --- Update attributes (F14) ---

export type TaskUpdateAttributes = Partial<
  Pick<TaskAttributes, "title" | "notes" | "status" | "dueOn" | "rolloverEnabled">
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
  return new Date(dueOn) < new Date();
}

export function isTaskCompleted(task: Task): boolean {
  return task.attributes.status === "completed";
}
