import { request } from './api';
import { PaginatedAssets } from '../types/asset';

export const assetService = {
  listByTarget: async (targetId: string, page = 1, limit = 50, search = ''): Promise<PaginatedAssets> => {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
      search: search,
    });
    return request<PaginatedAssets>(`/assets/target/${targetId}?${params.toString()}`);
  },
};
