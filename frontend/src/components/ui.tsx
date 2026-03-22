import type {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  PropsWithChildren,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes
} from "react";

import { cx } from "../lib/utils";

export function Button({
  className,
  children,
  ...props
}: PropsWithChildren<ButtonHTMLAttributes<HTMLButtonElement>>) {
  return (
    <button
      className={cx("ui-button", props.disabled && "is-disabled", className)}
      {...props}
    >
      {children}
    </button>
  );
}

export function GhostButton({
  className,
  children,
  ...props
}: PropsWithChildren<ButtonHTMLAttributes<HTMLButtonElement>>) {
  return (
    <button
      className={cx("ui-button ghost", props.disabled && "is-disabled", className)}
      {...props}
    >
      {children}
    </button>
  );
}

export function Badge({ children, className }: PropsWithChildren<{ className?: string }>) {
  return <span className={cx("ui-badge", className)}>{children}</span>;
}

export function Card({ children, className }: PropsWithChildren<{ className?: string }>) {
  return <section className={cx("ui-card", className)}>{children}</section>;
}

export function SectionTitle({
  title,
  description,
  action
}: {
  title: string;
  description?: string;
  action?: ReactNode;
}) {
  return (
    <div className="section-title">
      <div>
        <p className="eyebrow">Workspace</p>
        <h2>{title}</h2>
        {description ? <p>{description}</p> : null}
      </div>
      {action ? <div>{action}</div> : null}
    </div>
  );
}

export function StatCard({
  label,
  value,
  hint
}: {
  label: string;
  value: string | number;
  hint?: string;
}) {
  return (
    <Card className="stat-card">
      <span>{label}</span>
      <strong>{value}</strong>
      {hint ? <small>{hint}</small> : null}
    </Card>
  );
}

export function Field({
  label,
  hint,
  error,
  children
}: PropsWithChildren<{ label: string; hint?: string; error?: string }>) {
  return (
    <label className="field">
      <span className="field__label">{label}</span>
      {children}
      {hint ? <span className="field__hint">{hint}</span> : null}
      {error ? <span className="field__error">{error}</span> : null}
    </label>
  );
}

export function Input(props: InputHTMLAttributes<HTMLInputElement>) {
  return <input className="ui-input" {...props} />;
}

export function TextArea(props: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea className="ui-textarea" {...props} />;
}

export function Select(props: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className="ui-select" {...props} />;
}

export function EmptyState({
  title,
  description,
  action
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <Card className="empty-state">
      <h3>{title}</h3>
      <p>{description}</p>
      {action ? <div>{action}</div> : null}
    </Card>
  );
}

export function Modal({
  isOpen,
  title,
  children,
  onClose
}: PropsWithChildren<{
  isOpen: boolean;
  title: string;
  onClose: () => void;
}>) {
  if (!isOpen) {
    return null;
  }

  return (
    <div className="modal-backdrop" onClick={onClose} role="presentation">
      <div className="modal" onClick={(event) => event.stopPropagation()} role="dialog" aria-modal="true">
        <div className="modal__header">
          <h3>{title}</h3>
          <GhostButton type="button" onClick={onClose}>
            Закрыть
          </GhostButton>
        </div>
        <div className="modal__content">{children}</div>
      </div>
    </div>
  );
}

export function Tabs<T extends string>({
  value,
  options,
  onChange
}: {
  value: T;
  options: Array<{ value: T; label: string }>;
  onChange: (value: T) => void;
}) {
  return (
    <div className="tabs">
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          className={cx("tabs__item", option.value === value && "is-active")}
          onClick={() => onChange(option.value)}
        >
          {option.label}
        </button>
      ))}
    </div>
  );
}

export function LoadingScreen({ title, description }: { title: string; description?: string }) {
  return (
    <div className="loading-screen">
      <div className="loading-screen__pulse" />
      <h2>{title}</h2>
      {description ? <p>{description}</p> : null}
    </div>
  );
}
