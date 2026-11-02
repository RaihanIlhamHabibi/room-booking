import type { ButtonHTMLAttributes, HTMLAttributes, ReactNode } from "react";

export function Stack({ children, className = "", ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={`stack ${className}`} {...props}>{children}</div>;
}

export function Card({ children, className = "", ...props }: HTMLAttributes<HTMLElement>) {
  return <section className={`card ${className}`} {...props}>{children}</section>;
}

export function Button({ children, className = "", variant = "primary", ...props }: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: "primary" | "quiet" }) {
  return <button className={`button button-${variant} ${className}`} {...props}>{children}</button>;
}

export function Field({ id, label, children, hint, required }: { id: string; label: string; children: ReactNode; hint?: string; required?: boolean }) {
  return <div className="field"><label htmlFor={id}>{label}{required && <span className="required-label"> (required)</span>}</label>{children}{hint && <p className="hint">{hint}</p>}</div>;
}

export function Notice({ children, tone = "neutral", ...props }: HTMLAttributes<HTMLParagraphElement> & { tone?: "neutral" | "error" | "success" }) {
  return <p className={`notice notice-${tone}`} role={tone === "error" ? "alert" : "status"} {...props}>{children}</p>;
}
