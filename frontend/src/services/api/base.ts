import { api } from "@/lib/api/client";
import type { ApiResponse, ApiListResponse } from "@/types/api/responses";

export interface ValidationError {
  field: string;
  message: string;
}

export class BaseService {
  protected basePath: string;

  constructor(basePath: string) {
    this.basePath = basePath;
  }

  protected setTenant(tenant: { id: string }) {
    api.setTenant(tenant);
  }

  protected validate<T>(_data: T): ValidationError[] {
    return [];
  }

  protected transformResponse<T>(data: T): T {
    return data;
  }

  protected getList<T>(tenant: { id: string }, path?: string): Promise<ApiListResponse<T>> {
    this.setTenant(tenant);
    return api.get<ApiListResponse<T>>(path ?? this.basePath);
  }

  protected getOne<T>(tenant: { id: string }, path: string): Promise<ApiResponse<T>> {
    this.setTenant(tenant);
    return api.get<ApiResponse<T>>(path);
  }

  protected create<T>(tenant: { id: string }, path: string, body: unknown): Promise<ApiResponse<T>> {
    this.setTenant(tenant);
    return api.post<ApiResponse<T>>(path, body);
  }

  protected update<T>(tenant: { id: string }, path: string, body: unknown): Promise<ApiResponse<T>> {
    this.setTenant(tenant);
    return api.patch<ApiResponse<T>>(path, body);
  }

  protected remove(tenant: { id: string }, path: string): Promise<void> {
    this.setTenant(tenant);
    return api.delete(path);
  }
}
