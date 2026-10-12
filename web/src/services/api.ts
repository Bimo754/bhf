interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: string;
}

const API_BASE = '/api';

export async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE}${endpoint}`;
  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  const body: ApiResponse<T> = await response.json();

  if (!response.ok || !body.success) {
    throw new Error(body.error || `HTTP request failed with status ${response.status}`);
  }

  return body.data;
}
