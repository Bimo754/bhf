export type ScanStatus = 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED' | 'STOPPED';

export interface Scan {
  id: string;
  target_id: string;
  tool_name: string;
  status: ScanStatus;
  command: string;
  arguments: string[];
  started_at: string;
  completed_at?: string;
  error_msg?: string;
}
