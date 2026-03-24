export interface JsonApiResponse<T> {
  data: T;
  included?: Array<Record<string, unknown>>;
}

export interface JsonApiListResponse<T> {
  data: T[];
  included?: Array<Record<string, unknown>>;
}

export interface JsonApiError {
  status: string;
  code?: string;
  title: string;
  detail?: string;
}

export interface JsonApiErrorResponse {
  errors: JsonApiError[];
}
