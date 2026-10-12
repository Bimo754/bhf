import React from 'react';
import { Terminal } from '../components/Terminal';
import { LogEntry } from '../types/log';

interface TerminalViewProps {
  logs: LogEntry[];
  isRunning: boolean;
  onStop: () => void;
  onClear: () => void;
}

export const TerminalView: React.FC<TerminalViewProps> = ({
  logs,
  isRunning,
  onStop,
  onClear,
}) => {
  return (
    <div className="h-[calc(100vh-6.5rem)]">
      <Terminal
        logs={logs}
        title="Live Security Process Output Stream"
        isRunning={isRunning}
        onStop={onStop}
        onClear={onClear}
      />
    </div>
  );
};
