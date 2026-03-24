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
