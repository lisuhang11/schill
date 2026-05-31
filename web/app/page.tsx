import Link from "next/link";
import { Edit3, Search } from "lucide-react";
import { getPostList, getTopicList } from "@/lib/api";
import { Pagination } from "@/components/Pagination";
import { PostCard } from "@/components/PostCard";
import { Sidebar } from "@/components/Sidebar";
import { StateBlock } from "@/components/StateBlock";

type HomeProps = {
  searchParams: Promise<{ page?: string }>;
};

export default async function Home({ searchParams }: HomeProps) {
  const params = await searchParams;
  const page = Number(params.page ?? "1") || 1;
  const [postsResult, topicsResult] = await Promise.all([
    getPostList({ page, pageSize: 10 }),
    getTopicList({ page: 1, pageSize: 8, sort: "hot" })
  ]);

  const topics = topicsResult.ok ? topicsResult.data.list : [];

  return (
    <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-8 md:px-8 lg:grid-cols-[minmax(0,1fr)_320px]">
      <section className="space-y-6">
        <div className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-6 shadow-soft">
          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div>
              <p className="text-sm font-medium text-marine-deep">社区动态</p>
              <h1 className="mt-2 text-3xl font-semibold text-marine-text">最新内容</h1>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-marine-muted">
                浏览公开文章、参与评论互动，关注后端已确认的社区能力。
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              <Link
                href="/search"
                className="focus-ring inline-flex items-center gap-2 rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-4 py-2 text-sm font-semibold text-marine-deep"
              >
                <Search size={18} /> 搜索
              </Link>
              <Link
                href="/posts/new"
                className="focus-ring inline-flex items-center gap-2 rounded-lg bg-marine-deep px-4 py-2 text-sm font-semibold text-white"
              >
                <Edit3 size={18} /> 发布
              </Link>
            </div>
          </div>
        </div>

        {!postsResult.ok ? (
          <StateBlock
            tone="error"
            title="内容加载失败"
            description={`无法读取文章列表：${postsResult.message}`}
          />
        ) : postsResult.data.list.length === 0 ? (
          <StateBlock
            tone="empty"
            title="还没有内容"
            description="当前列表为空，可以先发布第一篇纯文本文章。"
          />
        ) : (
          <>
            <div className="space-y-4">
              {postsResult.data.list.map((post) => (
                <PostCard key={post.id} post={post} />
              ))}
            </div>
            <Pagination
              page={page}
              total={postsResult.data.total}
              pageSize={10}
              basePath="/"
            />
          </>
        )}
      </section>

      <Sidebar topics={topics} />
    </main>
  );
}
