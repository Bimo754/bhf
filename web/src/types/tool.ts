export type ToolCategory =
  | 'Subdomain Enumeration'
  | 'HTTP Probing'
  | 'URL Discovery'
  | 'Crawling & Spidering'
  | 'Vulnerability Scanning'
  | 'Fuzzing & Content Discovery';

export interface ToolInfo {
  name: string;
  binary_name: string;
  category: ToolCategory;
  description: string;
  is_installed: boolean;
}
