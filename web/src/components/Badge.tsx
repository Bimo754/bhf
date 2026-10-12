import React from 'react';
import { ScanStatus } from '../types/scan';

interface BadgeProps {
  status: ScanStatus | string;
}

export const Badge: React.FC<BadgeProps> = ({ status }) => {
  const getStyle = () => {
    switch (status.toUpperCase()) {
      case 'RUNNING':
        return 'bg-cyber-cyan/10 text-cyber-cyan border-cyber-cyan/30 animate-pulse';
      case 'COMPLETED':
        return 'bg-cyber-green/10 text-cyber-green border-cyber-green/30';
      case 'FAILED':
        return 'bg-cyber-rose/10 text-cyber-rose border-cyber-rose/30';
      case 'STOPPED':
        return 'bg-cyber-amber/10 text-cyber-amber border-cyber-amber/30';
      case 'PENDING':
      default:
        return 'bg-gray-800 text-gray-400 border-gray-700';
    }
  };

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-mono font-medium border ${getStyle()}`}
    >
      {status}
    </span>
  );
};
