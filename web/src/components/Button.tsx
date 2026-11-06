import React from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline';
  size?: 'xs' | 'sm' | 'md' | 'lg';
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
    'inline-flex items-center justify-center font-medium transition-all duration-150 rounded-lg focus:outline-none disabled:opacity-40 disabled:cursor-not-allowed select-none active:scale-[0.98]';

  const sizes = {
    xs: 'px-2 py-0.5 text-[11px]',
    sm: 'px-2.5 py-1 text-xs',
    md: 'px-3.5 py-1.5 text-xs',
    lg: 'px-4 py-2 text-sm',
  };

  const variants = {
    primary:
      'bg-indigo-600 hover:bg-indigo-500 text-white shadow-sm ring-1 ring-white/10 font-medium',
    secondary:
      'bg-white/[0.04] hover:bg-white/[0.08] text-slate-200 border border-white/[0.08] font-medium shadow-sm',
    outline:
      'bg-transparent hover:bg-white/[0.04] text-slate-300 border border-white/[0.08]',
    danger:
      'bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-rose-500/20',
    ghost:
      'text-slate-400 hover:text-slate-100 hover:bg-white/[0.04]',
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
