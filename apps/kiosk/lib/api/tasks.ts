export interface Task {
  id: string;
  title: string;
  status: 'pending' | 'completed' | 'overdue';
  dueDate: string;
  assignee?: string;
}

/**
 * Mock tasks data - will be replaced with real API call
 */
export async function getTasks(): Promise<Task[]> {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 300));

  const today = new Date().toISOString().split('T')[0];
  const yesterday = new Date(Date.now() - 86400000).toISOString().split('T')[0];

  return [
    {
      id: '1',
      title: 'Buy groceries',
      status: 'pending',
      dueDate: today,
      assignee: 'Mom',
    },
    {
      id: '2',
      title: 'Take out trash',
      status: 'pending',
      dueDate: today,
      assignee: 'Dad',
    },
    {
      id: '3',
      title: 'Call plumber',
      status: 'overdue',
      dueDate: yesterday,
      assignee: 'Mom',
    },
    {
      id: '4',
      title: 'Water plants',
      status: 'completed',
      dueDate: today,
      assignee: 'Kids',
    },
  ];
}
