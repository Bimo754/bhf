import { request } from './api';
import { ToolInfo } from '../types/tool';

export const toolService = {
  list: async (): Promise<ToolInfo[]> => {
    return request<ToolInfo[]>('/tools');
  },
};
