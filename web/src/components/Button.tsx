import React from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
}

export const Button: React.FC<ButtonProps> = ({
  children,
  variant = 'primary',
  size = 'md',
  isLoading = false,
  className = '',
  disabled,
  ...props
}) => {
  const base =
    'inline-flex items-center justify-center font-medium transition-colors rounded-md focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed';

  const sizes = {
    sm: 'px-2.5 py-1 text-xs',
    md: 'px-3 py-1.5 text-xs',
    lg: 'px-4 py-2 text-sm',
  };

  const variants = {
    primary:
      'bg-[#238636] hover:bg-[#2ea043] text-white border border-[#2ea043]/80 font-medium',
    secondary:
      'bg-[#21262d] hover:bg-[#30363d] text-[#c9d1d9] border border-[#30363d]',
    danger:
      'bg-red-950/70 hover:bg-red-900/80 text-red-300 border border-red-800/80',
    ghost:
      'text-zinc-400 hover:text-zinc-100 hover:bg-[#21262d]',
  };

  return (
    <button
      className={`${base} ${sizes[size]} ${variants[variant]} ${className}`}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading ? (
        <span className="flex items-center space-x-1.5">
          <svg className="animate-spin h-3.5 w-3.5 text-current" viewBox="0 0 24 24">
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
              fill="none"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8v8H4z"
            />
          </svg>
          <span>Running...</span>
        </span>
      ) : (
        children
      )}
    </button>
  );
};
