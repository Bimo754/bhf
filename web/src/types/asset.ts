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
  priority_score?: number;
  created_at: string;
}

export interface AssetStats {
  total: number;
  live: number;
  code_2xx: number;
  code_3xx: number;
  code_4xx_5xx: number;
  high_priority?: number;
  new_assets?: number;
}

export interface PaginatedAssets {
  assets: Asset[];
  total: number;
  page: number;
  limit: number;
  stats?: AssetStats;
}
