import { Link } from "react-router-dom";

import { cx } from "../lib/utils";

const BENZO_URL = "https://benzo.cloud";

export function LogoMark({ className }: { className?: string }) {
  return (
    <svg
      className={cx("logo-mark", className)}
      viewBox="0 0 160 180"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <defs>
        <linearGradient id="brand-helix" x1="20" y1="10" x2="132" y2="170" gradientUnits="userSpaceOnUse">
          <stop stopColor="#7EE7FF" />
          <stop offset="0.55" stopColor="#43B8FF" />
          <stop offset="1" stopColor="#1865FF" />
        </linearGradient>
      </defs>
      <path
        d="M48 16C102 42 102 68 48 93C-6 118 -6 144 48 168"
        stroke="url(#brand-helix)"
        strokeWidth="12"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M112 16C58 42 58 68 112 93C166 118 166 144 112 168"
        stroke="url(#brand-helix)"
        strokeWidth="12"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path d="M60 36H100" stroke="url(#brand-helix)" strokeWidth="10" strokeLinecap="round" />
      <path d="M56 64L104 92" stroke="url(#brand-helix)" strokeWidth="10" strokeLinecap="round" />
      <path d="M56 92L100 120" stroke="url(#brand-helix)" strokeWidth="10" strokeLinecap="round" />
      <path d="M60 144H100" stroke="url(#brand-helix)" strokeWidth="10" strokeLinecap="round" />
    </svg>
  );
}

export function BrandHomeLink({ compact = false, className = "" }: { compact?: boolean; className?: string }) {
  return (
    <Link className={cx("brand-link", compact && "brand-link--compact", className)} to="/">
      <LogoMark />
      <span>
        <strong>ПрофДНК</strong>
        <em>Benzo.Cloud</em>
      </span>
    </Link>
  );
}

export function BrandAttribution() {
  return (
    <p className="brand-attribution">
      Создан разработчиками из команды{" "}
      <a href={BENZO_URL} target="_blank" rel="noreferrer">
        Benzo.Cloud
      </a>
      .
    </p>
  );
}

export const BRAND_LINKS = {
  benzo: BENZO_URL,
  telegram: "https://t.me/benzocloud"
};
