"use client";

interface ErrorMessageProps {
  message: string;
  status?: number;
  onDismiss?: () => void;
  onRetry?: () => void;
}

export default function ErrorMessage({ message, status, onDismiss, onRetry }: ErrorMessageProps) {
  return (
    <div className="w-full max-w-xl mx-auto rounded-xl border border-[var(--color-error)]/30 bg-[var(--color-error)]/5 p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-2.5">
          <span className="text-[var(--color-error)] mt-0.5">
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
          </span>
          <div>
            <p className="text-sm text-[var(--color-text)]">{message}</p>
            {status ? (
              <p className="text-xs text-[var(--color-text-tertiary)] font-mono mt-1">HTTP {status}</p>
            ) : null}
            {onRetry && (
              <button onClick={onRetry} className="text-xs text-[var(--color-accent)] hover:underline mt-2 cursor-pointer">
                Try again
              </button>
            )}
          </div>
        </div>
        {onDismiss && (
          <button onClick={onDismiss} className="text-[var(--color-text-tertiary)] hover:text-[var(--color-text-secondary)] text-xs cursor-pointer">✕</button>
        )}
      </div>
    </div>
  );
}
