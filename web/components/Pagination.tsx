"use client";

import Link from "next/link";

type PaginationProps = {
  page: number;
  total: number;
  pageSize: number;
  basePath?: string;
  query?: Record<string, string | number | undefined>;
  onPageChange?: (page: number) => void;
};

export function Pagination({ page, total, pageSize, basePath, query, onPageChange }: PaginationProps) {
  const hasPrevious = page > 1;
  const hasNext = page * pageSize < total;

  function hrefForPage(p: number) {
    if (!basePath) return "#";
    const params = new URLSearchParams();
    Object.entries({ ...query, page: p }).forEach(([key, value]) => {
      if (value !== undefined && String(value) !== "") {
        params.set(key, String(value));
      }
    });
    return `${basePath}?${params.toString()}`;
  }

  const prevButton = hasPrevious ? (
    onPageChange ? (
      <button
        type="button"
        onClick={() => onPageChange(page - 1)}
        className="rounded-lg bg-marine-bg px-3 py-2 text-marine-deep"
      >
        上一页
      </button>
    ) : (
      <Link
        className="rounded-lg bg-marine-bg px-3 py-2 text-marine-deep"
        href={hrefForPage(page - 1)}
      >
        上一页
      </Link>
    )
  ) : (
    <span className="rounded-lg bg-slate-100 px-3 py-2 text-slate-400">上一页</span>
  );

  const nextButton = hasNext ? (
    onPageChange ? (
      <button
        type="button"
        onClick={() => onPageChange(page + 1)}
        className="rounded-lg bg-marine-deep px-3 py-2 text-white"
      >
        下一页
      </button>
    ) : (
      <Link
        className="rounded-lg bg-marine-deep px-3 py-2 text-white"
        href={hrefForPage(page + 1)}
      >
        下一页
      </Link>
    )
  ) : (
    <span className="rounded-lg bg-slate-100 px-3 py-2 text-slate-400">下一页</span>
  );

  return (
    <nav className="flex items-center justify-between rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-3 text-sm shadow-soft">
      <span className="text-marine-muted">第 {page} 页，共 {total} 条</span>
      <div className="flex gap-2">
        {prevButton}
        {nextButton}
      </div>
    </nav>
  );
}
