import { request } from './api';
import { PaginatedEndpoints, Endpoint } from '../types/endpoint';

export const endpointService = {
  listByTarget: async (targetId: string, page = 1, limit = 50, search = ''): Promise<PaginatedEndpoints> => {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
      search: search,
    });
    return request<PaginatedEndpoints>(`/endpoints/target/${targetId}?${params.toString()}`);
  },

  getAllForTarget: async (targetId: string): Promise<Endpoint[]> => {
    const res = await request<PaginatedEndpoints>(`/endpoints/target/${targetId}?all=true`);
    return res.endpoints || [];
  },
};
