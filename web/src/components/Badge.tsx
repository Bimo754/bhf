import React from 'react';
import { ScanStatus } from '../types/scan';

interface BadgeProps {
  status: ScanStatus | string;
}

export const Badge: React.FC<BadgeProps> = ({ status }) => {
  const getStyle = () => {
    switch (status.toUpperCase()) {
      case 'RUNNING':
        return 'bg-blue-950/70 text-blue-400 border-blue-800/80';
      case 'COMPLETED':
        return 'bg-emerald-950/70 text-emerald-400 border-emerald-800/80';
      case 'FAILED':
        return 'bg-red-950/70 text-red-400 border-red-800/80';
      case 'STOPPED':
        return 'bg-amber-950/70 text-amber-400 border-amber-800/80';
      case 'PENDING':
      default:
        return 'bg-zinc-800 text-zinc-400 border-zinc-700';
    }
  };

  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-[11px] font-mono font-medium border ${getStyle()}`}
    >
      {status}
    </span>
  );
};
