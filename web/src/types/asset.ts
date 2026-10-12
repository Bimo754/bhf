export interface Asset {
  id: string;
  target_id: string;
  scan_id: string;
  host: string;
  url?: string;
  ip?: string;
  port?: number;
  status_code?: number;
  title?: string;
  technologies?: string[];
  created_at: string;
}

export interface PaginatedAssets {
  assets: Asset[];
  total: number;
  page: number;
  limit: number;
}
