import { request } from './api';
import { Target } from '../types/target';

export interface CreateTargetPayload {
  name: string;
  domain: string;
  notes?: string;
}

export const targetService = {
  list: async (): Promise<Target[]> => {
    return request<Target[]>('/targets');
  },

  getById: async (id: string): Promise<Target> => {
    return request<Target>(`/targets/${id}`);
  },

  create: async (payload: CreateTargetPayload): Promise<Target> => {
    return request<Target>('/targets', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  delete: async (id: string): Promise<void> => {
    return request<void>(`/targets/${id}`, {
      method: 'DELETE',
    });
  },
};
