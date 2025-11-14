'use client';

import React from 'react';
import { Card, CardSection } from '@/app/components/ui/Card';
import { MealPlan } from '@/lib/api/meals';
import { UtensilsCrossed } from 'lucide-react';

interface MealsCardProps {
  mealPlan?: MealPlan | null;
  loading?: boolean;
  error?: string | null;
}

export function MealsCard({ mealPlan, loading, error }: MealsCardProps) {
  if (error) {
    return (
      <Card>
        <div className="flex items-center justify-between mb-2">
          <h3 className="font-semibold text-gray-900 dark:text-white">Meals</h3>
          <span className="text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 px-2 py-1 rounded">
            Preview
          </span>
        </div>
        <div className="text-center py-4">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      </Card>
    );
  }

  if (loading || !mealPlan) {
    return <Card loading={true}>{null}</Card>;
  }

  const today = new Date().toISOString().split('T')[0];
  const todayMeals = mealPlan.days.find(d => d.date === today);
  const upcomingDays = mealPlan.days.filter(d => d.date >= today).slice(0, 3);

  return (
    <Card>
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-semibold text-gray-900 dark:text-white">Meals</h3>
        <span className="text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 px-2 py-1 rounded">
          Preview
        </span>
      </div>

      <div className="space-y-4">
        {/* Today's Meals */}
        {todayMeals && (
          <CardSection title="Today">
            <div className="space-y-2">
              {todayMeals.meals.breakfast && (
                <MealItem label="Breakfast" meal={todayMeals.meals.breakfast} />
              )}
              {todayMeals.meals.lunch && (
                <MealItem label="Lunch" meal={todayMeals.meals.lunch} />
              )}
              {todayMeals.meals.dinner && (
                <MealItem label="Dinner" meal={todayMeals.meals.dinner} />
              )}
            </div>
          </CardSection>
        )}

        {/* Upcoming Meals */}
        <CardSection title="This Week">
          <div className="space-y-3">
            {upcomingDays.slice(1).map(day => (
              <div
                key={day.date}
                className="pb-3 border-b border-gray-200 dark:border-gray-700 last:border-0"
              >
                <div className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {day.dayName}
                </div>
                <div className="text-sm text-gray-600 dark:text-gray-400 space-y-0.5">
                  {day.meals.dinner && (
                    <div className="flex items-center gap-2">
                      <UtensilsCrossed className="h-3 w-3" />
                      <span>{day.meals.dinner}</span>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </CardSection>
      </div>
    </Card>
  );
}

function MealItem({ label, meal }: { label: string; meal: string }) {
  return (
    <div className="flex items-center justify-between py-2">
      <span className="text-sm text-gray-600 dark:text-gray-400">{label}</span>
      <span className="text-sm font-medium text-gray-900 dark:text-white">{meal}</span>
    </div>
  );
}
