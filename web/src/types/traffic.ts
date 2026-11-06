export interface TrafficRecord {
  id: string;
  target_id: string;
  scan_id?: string;
  method: string;
  url: string;
  path: string;
  host: string;
  request_headers: string;
  request_body: string;
  response_status_code: number;
  response_headers: string;
  response_body: string;
  content_type: string;
  content_length: number;
  duration_ms: number;
  category: string;
  auth_context: string;
  parameters: string;
  created_at: string;
}

export interface PaginatedTraffic {
  traffic: TrafficRecord[];
  total: number;
  page: number;
  limit: number;
}

export interface TrafficStats {
  target_id: string;
  categories: Record<string, number>;
}
