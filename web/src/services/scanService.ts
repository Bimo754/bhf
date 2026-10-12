import { request } from './api';
import { Scan } from '../types/scan';
import { LogEntry } from '../types/log';

export interface StartScanPayload {
  target_id: string;
  tool_name: string;
  custom_args?: string[];
}

export const scanService = {
  start: async (payload: StartScanPayload): Promise<Scan> => {
    return request<Scan>('/scans/start', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  stop: async (scanId: string): Promise<void> => {
    return request<void>(`/scans/${scanId}/stop`, {
      method: 'POST',
    });
  },

  listByTarget: async (targetId: string): Promise<Scan[]> => {
    return request<Scan[]>(`/scans/target/${targetId}`);
  },

  getLogs: async (scanId: string): Promise<LogEntry[]> => {
    return request<LogEntry[]>(`/scans/${scanId}/logs`);
  },
};
