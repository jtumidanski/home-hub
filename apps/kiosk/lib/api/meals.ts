export interface DayMeals {
  date: string;
  dayName: string;
  meals: {
    breakfast?: string;
    lunch?: string;
    dinner?: string;
  };
}

export interface MealPlan {
  weekStarting: string;
  days: DayMeals[];
}

/**
 * Mock meal plan data - will be replaced with real API call
 */
export async function getMealPlan(): Promise<MealPlan> {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 300));

  const today = new Date();
  const startOfWeek = new Date(today);
  startOfWeek.setDate(today.getDate() - today.getDay()); // Sunday

  const days: DayMeals[] = [];
  const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  const meals = [
    { breakfast: 'Pancakes', lunch: 'Sandwiches', dinner: 'Spaghetti' },
    { breakfast: 'Cereal', lunch: 'Leftover spaghetti', dinner: 'Grilled chicken' },
    { breakfast: 'Toast & eggs', lunch: 'Salad', dinner: 'Tacos' },
    { breakfast: 'Oatmeal', lunch: 'Soup', dinner: 'Stir fry' },
    { breakfast: 'Yogurt', lunch: 'Pizza', dinner: 'Fish & rice' },
    { breakfast: 'French toast', lunch: 'Burgers', dinner: 'Pasta primavera' },
    { breakfast: 'Bagels', lunch: 'Leftovers', dinner: 'BBQ ribs' },
  ];

  for (let i = 0; i < 7; i++) {
    const date = new Date(startOfWeek);
    date.setDate(startOfWeek.getDate() + i);

    days.push({
      date: date.toISOString().split('T')[0],
      dayName: dayNames[i],
      meals: meals[i],
    });
  }

  return {
    weekStarting: startOfWeek.toISOString().split('T')[0],
    days,
  };
}
