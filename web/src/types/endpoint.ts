export interface Endpoint {
  id: string;
  target_id: string;
  scan_id?: string;
  host: string;
  url: string;
  path: string;
  method?: string;
  status_code?: number;
  content_length?: number;
  title?: string;
  source_tool?: string;
  created_at: string;
}

export interface PaginatedEndpoints {
  endpoints: Endpoint[];
  total: number;
  page: number;
  limit: number;
}
