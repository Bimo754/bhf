import { request } from './api';
import { PaginatedAssets } from '../types/asset';

export const assetService = {
  listByTarget: async (targetId: string, page = 1, limit = 50, search = '', filter = ''): Promise<PaginatedAssets> => {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
      search: search,
    });
    if (filter && filter !== 'all') {
      params.set('filter', filter);
    }
    return request<PaginatedAssets>(`/assets/target/${targetId}?${params.toString()}`);
  },
};
