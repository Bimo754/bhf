import { request } from './api';
import { PaginatedTraffic, TrafficRecord, TrafficStats } from '../types/traffic';

export interface AssessmentRunResponse {
  target_id: string;
  scan_id: string;
  traffic_analyzed: number;
  findings_count: number;
  findings: any[];
}

export interface ProxyStatus {
  running: boolean;
  addr?: string;
  target_id?: string;
}

export interface ExplorerProgress {
  target_id: string;
  current_url: string;
  pages_visited: number;
  elements_clicked: number;
  traffic_captured: number;
  categories_found: string[];
  is_complete: boolean;
  engine_used: string;
  error?: string;
}

export const trafficService = {
  listByTarget: async (
    targetId: string,
    page = 1,
    limit = 50,
    category = '',
    method = '',
    search = ''
  ): Promise<PaginatedTraffic> => {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    });
    if (category && category !== 'all') {
      params.append('category', category);
    }
    if (method && method !== 'all') {
      params.append('method', method);
    }
    if (search) {
      params.append('search', search);
    }
    return request<PaginatedTraffic>(`/traffic/target/${targetId}?${params.toString()}`);
  },

  getById: async (id: string): Promise<TrafficRecord> => {
    return request<TrafficRecord>(`/traffic/${id}`);
  },

  getStats: async (targetId: string): Promise<TrafficStats> => {
    return request<TrafficStats>(`/traffic/target/${targetId}/stats`);
  },

  runAssessment: async (targetId: string): Promise<AssessmentRunResponse> => {
    return request<AssessmentRunResponse>(`/assessment/target/${targetId}/run`, {
      method: 'POST',
    });
  },

  getProxyStatus: async (): Promise<ProxyStatus> => {
    return request<ProxyStatus>('/appintel/proxy/status');
  },

  startProxy: async (targetId: string, port = 8888): Promise<ProxyStatus> => {
    return request<ProxyStatus>('/appintel/proxy/start', {
      method: 'POST',
      body: JSON.stringify({ target_id: targetId, port }),
    });
  },

  stopProxy: async (): Promise<ProxyStatus> => {
    return request<ProxyStatus>('/appintel/proxy/stop', {
      method: 'POST',
    });
  },

  launchBrowser: async (targetId: string, url?: string, headless = false): Promise<any> => {
    return request<any>('/appintel/browser/launch', {
      method: 'POST',
      body: JSON.stringify({ target_id: targetId, url, headless }),
    });
  },

  autoExplore: async (
    targetId: string,
    options?: {
      start_url?: string;
      max_pages?: number;
      max_depth?: number;
      prefer_chrome?: boolean;
    }
  ): Promise<ExplorerProgress> => {
    return request<ExplorerProgress>(`/appintel/target/${targetId}/auto-explore`, {
      method: 'POST',
      body: JSON.stringify(options || {}),
    });
  },

  getAutoExploreStatus: async (
    targetId: string
  ): Promise<{ target_id: string; running: boolean; progress?: ExplorerProgress }> => {
    return request<{ target_id: string; running: boolean; progress?: ExplorerProgress }>(
      `/appintel/target/${targetId}/auto-explore/status`
    );
  },
};
