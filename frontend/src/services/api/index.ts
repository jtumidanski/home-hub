// Base
export { BaseService } from "./base";
export type { ValidationError } from "./base";

// Singleton instances
export { authService } from "./auth";
export { accountService } from "./account";
export { productivityService } from "./productivity";

// Types re-exported per service
export type { AuthProvider } from "./auth";
