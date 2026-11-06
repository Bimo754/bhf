import { request } from './api';
import { Host, NetworkService, ADAssessmentResult, NetworkScanResponse } from '../types/network';

export const networkService = {
  listHosts: async (targetId: string): Promise<Host[]> => {
    return request<Host[]>(`/network/target/${targetId}/hosts`);
  },

  listServices: async (targetId: string): Promise<NetworkService[]> => {
    return request<NetworkService[]>(`/network/target/${targetId}/services`);
  },

  scanSubnet: async (targetId: string, cidr?: string, ports?: number[]): Promise<NetworkScanResponse> => {
    return request<NetworkScanResponse>(`/network/target/${targetId}/scan`, {
      method: 'POST',
      body: JSON.stringify({ cidr, ports }),
    });
  },

  runADAudit: async (targetId: string, hostIp?: string): Promise<ADAssessmentResult[]> => {
    return request<ADAssessmentResult[]>(`/network/target/${targetId}/ad-audit`, {
      method: 'POST',
      body: JSON.stringify({ host_ip: hostIp }),
    });
  },
};
