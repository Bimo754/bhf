export type LogStream = 'stdout' | 'stderr' | 'system';

export interface LogEntry {
  id?: number;
  scan_id: string;
  stream: LogStream;
  line: string;
  timestamp: string;
}

export interface StreamMessage {
  type: 'log' | 'status' | 'complete' | 'error';
  scan_id: string;
  status?: string;
  stream?: LogStream;
  line?: string;
  timestamp: string;
}
