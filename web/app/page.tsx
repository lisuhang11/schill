import Link from "next/link";
import { ArrowRight, BarChart3, Building2, Edit3, MessageCircle, Search, ShieldCheck, Sparkles, Users } from "lucide-react";
import { getPostList, getTopicList } from "@/lib/api";
import { formatCount } from "@/lib/format";
import { Pagination } from "@/components/Pagination";
import { PostCard } from "@/components/PostCard";
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

  const topics = topicsResult.ok ? (topicsResult.data.list ?? []) : [];

  // Backend may return data as empty object {} when no records exist
  const postList = postsResult.ok ? (postsResult.data.list ?? []) : [];
  const postTotal = postsResult.ok ? (postsResult.data.total ?? 0) : 0;
  const commentTotal = postList.reduce((sum, post) => sum + (post.commentCount ?? 0), 0);
  const collectionTotal = postList.reduce((sum, post) => sum + (post.collectionCount ?? 0), 0);
  const featuredPosts = postList.slice(0, 3);

  return (
    <main className="mx-auto w-full max-w-6xl px-4 py-8 md:px-8">
      <section className="overflow-hidden rounded-lg border border-[rgba(77,100,124,0.16)] bg-white shadow-soft">
        <div className="grid gap-0 lg:grid-cols-[minmax(0,1fr)_360px]">
          <div className="p-6 md:p-8">
            <div className="flex flex-wrap items-center gap-2 text-xs font-semibold text-marine-deep">
              <span className="inline-flex items-center gap-1 rounded-lg bg-marine-bg px-2.5 py-1">
                <Building2 size={14} /> Schill Enterprise
              </span>
              <span className="rounded-lg bg-marine-warm/45 px-2.5 py-1 text-marine-text">海盐蓝白工作台</span>
            </div>
            <h1 className="mt-5 max-w-3xl text-4xl font-semibold leading-tight text-marine-text md:text-5xl">
              企业级内容协作与社区运营门户
            </h1>
            <p className="mt-4 max-w-2xl text-sm leading-7 text-marine-muted md:text-base">
              聚合内容发布、话题洞察、互动数据和团队知识流，让公开文章与社区讨论保持清爽、可扫描、可持续运营。
            </p>
            <div className="mt-6 flex flex-wrap gap-3">
              <Link
                href="/posts/new"
                className="focus-ring inline-flex h-11 items-center gap-2 rounded-lg bg-marine-deep px-4 text-sm font-semibold text-white shadow-sm"
              >
                <Edit3 size={18} /> 发布内容
              </Link>
              <Link
                href="/search"
                className="focus-ring inline-flex h-11 items-center gap-2 rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-4 text-sm font-semibold text-marine-deep"
              >
                <Search size={18} /> 全站检索
              </Link>
            </div>
          </div>
          <div className="border-t border-[rgba(77,100,124,0.12)] bg-marine-bg/70 p-6 lg:border-l lg:border-t-0">
            <div className="rounded-lg bg-white p-5 shadow-soft">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <p className="text-sm font-semibold text-marine-text">运营概览</p>
                  <p className="mt-1 text-xs text-marine-muted">基于当前列表实时汇总</p>
                </div>
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-marine-mint/25 text-marine-deep">
                  <BarChart3 size={20} />
                </span>
              </div>
              <div className="mt-5 grid grid-cols-2 gap-3">
                <Metric label="内容" value={formatCount(postTotal)} />
                <Metric label="互动评论" value={formatCount(commentTotal)} />
                <Metric label="收藏" value={formatCount(collectionTotal)} />
                <Metric label="热门话题" value={formatCount(topics.length)} />
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="mt-6 grid gap-4 md:grid-cols-3">
        <FeatureTile
          icon={ShieldCheck}
          title="可信内容治理"
          description="公开、粉丝、互关等可见性状态在列表中清晰呈现。"
        />
        <FeatureTile
          icon={Users}
          title="团队知识沉淀"
          description="话题与标签帮助内容快速归档，适合长期运营。"
        />
        <FeatureTile
          icon={Sparkles}
          title="轻量社区氛围"
          description="保留海盐蓝白的明亮感，避免后台系统的压迫感。"
        />
      </section>

      <section className="mt-8 grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
        <div className="space-y-5">
          <div className="flex flex-col gap-3 border-b border-[rgba(77,100,124,0.12)] pb-4 md:flex-row md:items-end md:justify-between">
            <div>
              <p className="text-sm font-medium text-marine-deep">内容中心</p>
              <h2 className="mt-1 text-2xl font-semibold text-marine-text">最新发布</h2>
            </div>
            <Link href="/feed" className="inline-flex items-center gap-1 text-sm font-semibold text-marine-deep hover:text-marine-text">
              查看动态 <ArrowRight size={16} />
            </Link>
          </div>
        {!postsResult.ok ? (
          <StateBlock
            tone="error"
            title="内容加载失败"
            description={`无法读取文章列表：${postsResult.message}`}
          />
        ) : postList.length === 0 ? (
          <StateBlock
            tone="empty"
            title="还没有内容"
            description="当前列表为空，可以先发布第一篇纯文本文章。"
          />
        ) : (
          <>
            <div className="space-y-4">
              {postList.map((post) => (
                <PostCard key={post.id} post={post} />
              ))}
            </div>
            <Pagination
              page={page}
              total={postTotal}
              pageSize={10}
              basePath="/"
            />
          </>
        )}
        </div>

        <aside className="space-y-5">
          <section className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h2 className="text-base font-semibold text-marine-text">热门话题</h2>
                <p className="mt-1 text-xs text-marine-muted">运营推荐入口</p>
              </div>
              <Sparkles size={18} className="text-marine-warm" />
            </div>
            <div className="mt-4 flex flex-wrap gap-2">
              {topics.length ? (
                topics.slice(0, 10).map((topic) => (
                  <Link
                    href={`/search?keyword=${encodeURIComponent(topic.name)}&type=post`}
                    key={topic.id}
                    className="rounded-lg bg-marine-bg px-3 py-2 text-xs font-medium text-marine-deep hover:bg-marine-blue/20"
                  >
                    #{topic.name}
                  </Link>
                ))
              ) : (
                <p className="text-sm leading-6 text-marine-muted">暂无话题数据。</p>
              )}
            </div>
          </section>

          <section className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft">
            <h2 className="text-base font-semibold text-marine-text">精选内容</h2>
            <div className="mt-4 space-y-3">
              {featuredPosts.length ? (
                featuredPosts.map((post) => (
                  <Link key={post.id} href={`/posts/${post.id}`} className="block rounded-lg border border-[rgba(77,100,124,0.12)] p-3 hover:bg-marine-bg/70">
                    <p className="line-clamp-1 text-sm font-semibold text-marine-text">{post.title}</p>
                    <p className="mt-2 inline-flex items-center gap-1 text-xs text-marine-muted">
                      <MessageCircle size={13} /> {formatCount(post.commentCount)} 条讨论
                    </p>
                  </Link>
                ))
              ) : (
                <p className="text-sm leading-6 text-marine-muted">暂无精选内容。</p>
              )}
            </div>
          </section>
        </aside>
      </section>
    </main>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-[rgba(77,100,124,0.12)] bg-white px-3 py-3">
      <p className="text-xl font-semibold text-marine-text">{value}</p>
      <p className="mt-1 text-xs text-marine-muted">{label}</p>
    </div>
  );
}

function FeatureTile({
  icon: Icon,
  title,
  description
}: {
  icon: typeof ShieldCheck;
  title: string;
  description: string;
}) {
  return (
    <div className="rounded-lg border border-[rgba(77,100,124,0.14)] bg-white p-5 shadow-soft">
      <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-marine-bg text-marine-deep">
        <Icon size={20} />
      </span>
      <h2 className="mt-4 text-base font-semibold text-marine-text">{title}</h2>
      <p className="mt-2 text-sm leading-6 text-marine-muted">{description}</p>
    </div>
  );
}
