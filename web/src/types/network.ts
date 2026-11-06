export interface Host {
  id: string;
  target_id: string;
  ip: string;
  hostname: string;
  os: string;
  mac: string;
  status: string;
  created_at: string;
}

export interface NetworkService {
  id: string;
  host_id: string;
  target_id: string;
  port: number;
  protocol: string;
  name: string;
  product: string;
  version: string;
  banner: string;
  status: string;
  created_at: string;
}

export interface ADAssessmentResult {
  host_ip: string;
  is_domain_controller: boolean;
  domain_name?: string;
  open_services: string[];
  observations: string[];
}

export interface NetworkScanResponse {
  target_id: string;
  scan_id: string;
  hosts_discovered: number;
  services_open: number;
  hosts: Host[];
  services: NetworkService[];
}
