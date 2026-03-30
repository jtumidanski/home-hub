// Base
export { BaseService } from "./base";
export type { ValidationError } from "./base";

// Singleton instances
export { authService } from "./auth";
export { accountService } from "./account";
export { productivityService } from "./productivity";
export { weatherService } from "./weather";
export { recipeService } from "./recipe";
export { ingredientService } from "./ingredient";
export { mealsService } from "./meals";
export { calendarService } from "./calendar";

// Types re-exported per service
export type { AuthProvider } from "./auth";
