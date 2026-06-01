import Link from "next/link";
import { getTopicList } from "@/lib/api";
import { Pagination } from "@/components/Pagination";
import { StateBlock } from "@/components/StateBlock";

type TopicsPageProps = {
  searchParams: Promise<{ page?: string; sort?: string }>;
};

export default async function TopicsPage({ searchParams }: TopicsPageProps) {
  const params = await searchParams;
  const page = Number(params.page ?? "1") || 1;
  const sort = params.sort === "new" ? "new" : "hot";
  const result = await getTopicList({ page, pageSize: 24, sort });
  const topicList = result.ok ? (result.data.list ?? []) : null;
  const topicTotal = result.ok ? (result.data.total ?? 0) : 0;

  return (
    <main className="mx-auto w-full max-w-6xl px-4 py-8 md:px-8">
      <div className="flex flex-col gap-4 rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-6 shadow-soft md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-3xl font-semibold text-marine-text">话题广场</h1>
          <p className="mt-2 text-sm text-marine-muted">话题详情接口尚未确认，点击话题会进入文章搜索。</p>
        </div>
        <div className="flex gap-2">
          <Link className={`rounded-lg px-3 py-2 text-sm ${sort === "hot" ? "bg-marine-deep text-white" : "bg-marine-bg text-marine-deep"}`} href="/topics?sort=hot">
            热门
          </Link>
          <Link className={`rounded-lg px-3 py-2 text-sm ${sort === "new" ? "bg-marine-deep text-white" : "bg-marine-bg text-marine-deep"}`} href="/topics?sort=new">
            最新
          </Link>
        </div>
      </div>

      <section className="mt-6">
        {!result.ok ? (
          <StateBlock tone="error" title="话题加载失败" description={result.message} />
        ) : topicList.length === 0 ? (
          <StateBlock tone="empty" title="暂无话题" description="还没有可展示的话题。" />
        ) : (
          <>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {topicList.map((topic) => (
                <Link
                  key={topic.id}
                  href={`/search?keyword=${encodeURIComponent(topic.name)}&type=post`}
                  className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft transition hover:-translate-y-0.5 hover:shadow-float"
                >
                  <p className="text-lg font-semibold text-marine-text">#{topic.name}</p>
                  <p className="mt-2 text-sm text-marine-muted">引用 {topic.quoteNum}</p>
                </Link>
              ))}
            </div>
            <div className="mt-5">
              <Pagination page={page} total={topicTotal} pageSize={24} basePath="/topics" query={{ sort }} />
            </div>
          </>
        )}
      </section>
    </main>
  );
}
