"use client";

type StateBlockProps = {
  tone: "loading" | "empty" | "error" | "info";
  title: string;
  description: string;
  actionLabel?: string;
  onAction?: () => void;
};

const toneStyles = {
  loading: "border-marine-blue/40 bg-marine-bg",
  empty: "border-marine-mint/40 bg-white",
  error: "border-red-200 bg-red-50",
  info: "border-marine-warm/60 bg-white"
};

export function StateBlock({
  tone,
  title,
  description,
  actionLabel,
  onAction
}: StateBlockProps) {
  return (
    <section className={`rounded-lg border p-6 text-center shadow-soft ${toneStyles[tone]}`}>
      <h2 className="text-lg font-semibold text-marine-text">{title}</h2>
      <p className="mx-auto mt-2 max-w-xl text-sm leading-6 text-marine-muted">{description}</p>
      {actionLabel && onAction ? (
        <button
          type="button"
          onClick={onAction}
          className="focus-ring mt-4 rounded-lg bg-marine-deep px-4 py-2 text-sm font-semibold text-white"
        >
          {actionLabel}
        </button>
      ) : null}
    </section>
  );
}
