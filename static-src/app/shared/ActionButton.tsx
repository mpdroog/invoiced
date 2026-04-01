import * as React from "react";
import { useState } from "react";

interface ActionButtonProps extends Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, 'onClick'> {
  onClick: (e: React.MouseEvent<HTMLButtonElement>) => Promise<void> | void;
  children: React.ReactNode;
}

/**
 * Button that disables itself while the onClick action is running.
 * Prevents double-clicks and shows a spinner during async operations.
 */
export function ActionButton({ onClick, disabled, children, className, ...props }: ActionButtonProps): React.JSX.Element {
  const [loading, setLoading] = useState(false);

  const handleClick = (e: React.MouseEvent<HTMLButtonElement>): void => {
    e.preventDefault();
    if (loading || disabled) return;
    setLoading(true);
    void Promise.resolve(onClick(e))
      .catch((err: unknown) => {
        handleErr(err);
      })
      .finally(() => {
        setLoading(false);
      });
  };

  return (
    <button
      {...props}
      type="button"
      className={className}
      disabled={disabled === true || loading}
      onClick={handleClick}
    >
      {loading ? <><i className="fas fa-spinner fa-spin" /> Working...</> : children}
    </button>
  );
}

interface ActionLinkProps extends Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'onClick'> {
  onClick: (e: React.MouseEvent<HTMLAnchorElement>) => Promise<void> | void;
  disabled?: boolean;
  children: React.ReactNode;
}

/**
 * Link/anchor that disables itself while the onClick action is running.
 * Prevents double-clicks and shows a spinner during async operations.
 */
export function ActionLink({ onClick, disabled, children, className, ...props }: ActionLinkProps): React.JSX.Element {
  const [loading, setLoading] = useState(false);

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>): void => {
    e.preventDefault();
    if (loading || disabled) return;
    setLoading(true);
    void Promise.resolve(onClick(e))
      .catch((err: unknown) => {
        handleErr(err);
      })
      .finally(() => {
        setLoading(false);
      });
  };

  const disabledClass = (disabled || loading) ? " disabled" : "";

  return (
    <a
      {...props}
      className={`${className ?? ""}${disabledClass}`}
      onClick={handleClick}
    >
      {loading ? <i className="fas fa-spinner fa-spin" /> : children}
    </a>
  );
}
