import { request } from './api';
import { PaginatedFindings } from '../types/finding';

export const findingService = {
  listByTarget: async (targetId: string, page = 1, limit = 50, search = '', severity = ''): Promise<PaginatedFindings> => {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
      search: search,
    });
    if (severity && severity !== 'all') {
      params.set('severity', severity);
    }
    return request<PaginatedFindings>(`/findings/target/${targetId}?${params.toString()}`);
  },

  delete: async (id: string): Promise<void> => {
    return request<void>(`/findings/${id}`, { method: 'DELETE' });
  },
};
