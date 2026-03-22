import type { PropsWithChildren, ReactNode } from "react";

import { BrandHomeLink } from "./brand";
import { Badge, GhostButton } from "./ui";

export function WorkspaceLayout({
  title,
  subtitle,
  badge,
  aside,
  topActions,
  children
}: PropsWithChildren<{
  title: string;
  subtitle: string;
  badge?: string;
  aside: ReactNode;
  topActions?: ReactNode;
}>) {
  return (
    <main className="workspace">
      <aside className="workspace__sidebar">{aside}</aside>
      <section className="workspace__main">
        <header className="workspace__header">
          <div>
            {badge ? <Badge>{badge}</Badge> : null}
            <h1>{title}</h1>
            <p>{subtitle}</p>
          </div>
          {topActions ? <div className="workspace__actions">{topActions}</div> : null}
        </header>
        <div className="workspace__content">{children}</div>
      </section>
    </main>
  );
}

export function WorkspaceNav<T extends string>({
  title,
  subtitle,
  value,
  options,
  onChange,
  footer
}: {
  title: string;
  subtitle: string;
  value: T;
  options: Array<{ value: T; label: string; description: string }>;
  onChange: (value: T) => void;
  footer?: ReactNode;
}) {
  return (
    <div className="workspace-nav">
      <div className="workspace-nav__brand">
        <BrandHomeLink compact />
        <p className="eyebrow">ProfDNA</p>
        <strong>{title}</strong>
        <span>{subtitle}</span>
      </div>
      <nav className="workspace-nav__links">
        {options.map((option) => (
          <button
            key={option.value}
            type="button"
            className={`workspace-nav__link ${value === option.value ? "is-active" : ""}`}
            onClick={() => onChange(option.value)}
          >
            <strong>{option.label}</strong>
            <span>{option.description}</span>
          </button>
        ))}
      </nav>
      {footer ? <div className="workspace-nav__footer">{footer}</div> : null}
    </div>
  );
}

export function LogoutAction({ onLogout }: { onLogout: () => void }) {
  return (
    <GhostButton type="button" onClick={onLogout}>
      Выйти
    </GhostButton>
  );
}
