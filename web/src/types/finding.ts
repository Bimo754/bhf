export type FindingSeverity = 'critical' | 'high' | 'medium' | 'low' | 'info';

export interface Finding {
  id: string;
  target_id: string;
  scan_id: string;
  tool_name: string;
  severity: FindingSeverity;
  name: string;
  matched_at: string;
  description?: string;
  extracted_results?: string[];
  curl_command?: string;
  created_at: string;
}

export interface FindingStats {
  total: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  info: number;
}

export interface PaginatedFindings {
  findings: Finding[];
  total: number;
  page: number;
  limit: number;
  stats: FindingStats;
}
