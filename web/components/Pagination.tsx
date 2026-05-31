import Link from "next/link";

type PaginationProps = {
  page: number;
  total: number;
  pageSize: number;
  basePath: string;
  query?: Record<string, string | number | undefined>;
};

function hrefFor(
  basePath: string,
  page: number,
  query?: Record<string, string | number | undefined>
) {
  const params = new URLSearchParams();
  Object.entries({ ...query, page }).forEach(([key, value]) => {
    if (value !== undefined && String(value) !== "") {
      params.set(key, String(value));
    }
  });
  return `${basePath}?${params.toString()}`;
}

export function Pagination({ page, total, pageSize, basePath, query }: PaginationProps) {
  const hasPrevious = page > 1;
  const hasNext = page * pageSize < total;

  return (
    <nav className="flex items-center justify-between rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-3 text-sm shadow-soft">
      <span className="text-marine-muted">第 {page} 页，共 {total} 条</span>
      <div className="flex gap-2">
        <Link
          aria-disabled={!hasPrevious}
          className={`rounded-lg px-3 py-2 ${hasPrevious ? "bg-marine-bg text-marine-deep" : "pointer-events-none bg-slate-100 text-slate-400"}`}
          href={hrefFor(basePath, page - 1, query)}
        >
          上一页
        </Link>
        <Link
          aria-disabled={!hasNext}
          className={`rounded-lg px-3 py-2 ${hasNext ? "bg-marine-deep text-white" : "pointer-events-none bg-slate-100 text-slate-400"}`}
          href={hrefFor(basePath, page + 1, query)}
        >
          下一页
        </Link>
      </div>
    </nav>
  );
}
