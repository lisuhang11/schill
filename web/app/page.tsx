import Link from "next/link";
import {
  ArrowRight,
  BarChart3,
  BookOpen,
  Bookmark,
  CalendarCheck,
  Edit3,
  Flame,
  MessageCircle,
  Route,
  Search,
  Sparkles,
  Tags,
  Trophy,
  Users
} from "lucide-react";
import { getPostList, getTopicList } from "@/lib/api";
import { formatCount } from "@/lib/format";
import { Pagination } from "@/components/Pagination";
import { PostCard } from "@/components/PostCard";
import { StateBlock } from "@/components/StateBlock";
import type { PostInfo, TopicInfo } from "@/lib/types";

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
  const postList = postsResult.ok ? (postsResult.data.list ?? []) : [];
  const postTotal = postsResult.ok ? (postsResult.data.total ?? 0) : 0;
  const dailyPost = postList[0];
  const featuredPosts = postList.slice(0, 3);
  const commentTotal = postList.reduce((sum, post) => sum + (post.commentCount ?? 0), 0);
  const collectionTotal = postList.reduce((sum, post) => sum + (post.collectionCount ?? 0), 0);
  const hotTopic = topics[0];

  return (
    <main className="mx-auto w-full max-w-6xl px-4 py-7 md:px-8 md:py-9">
      <section className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_360px]">
        <div className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft md:p-7">
          <div className="flex flex-wrap items-center gap-2 text-xs font-semibold text-marine-deep">
            <span className="inline-flex items-center gap-1 rounded-lg bg-marine-bg px-2.5 py-1">
              <Sparkles size={14} />
              今日学习台
            </span>
            <span className="rounded-lg bg-marine-warm/45 px-2.5 py-1 text-marine-text">
              海盐蓝白 · 内容练习社区
            </span>
          </div>

          <h1 className="mt-5 max-w-3xl text-3xl font-semibold leading-tight text-marine-text md:text-5xl">
            找到一个主题，开始今天的技术练习
          </h1>
          <p className="mt-4 max-w-2xl text-sm leading-7 text-marine-muted md:text-base">
            像刷题一样浏览内容、追踪话题、沉淀讨论。首页把搜索、今日推荐、学习路径和社区动态放在同一屏，打开就能继续推进。
          </p>

          <form action="/search" className="mt-6 flex flex-col gap-3 sm:flex-row">
            <label className="sr-only" htmlFor="home-search">
              搜索内容
            </label>
            <div className="flex min-h-12 flex-1 items-center gap-3 rounded-lg border border-[rgba(77,100,124,0.18)] bg-marine-bg/60 px-4">
              <Search size={19} className="shrink-0 text-marine-deep" />
              <input
                id="home-search"
                name="keyword"
                placeholder="搜索帖子、话题或学习关键词"
                className="h-12 min-w-0 flex-1 bg-transparent text-sm text-marine-text outline-none placeholder:text-marine-muted/70"
              />
              <input type="hidden" name="type" value="post" />
            </div>
            <button
              type="submit"
              className="focus-ring inline-flex h-12 items-center justify-center gap-2 rounded-lg bg-marine-deep px-5 text-sm font-semibold text-white shadow-soft transition hover:bg-[#244f86]"
            >
              <Search size={18} />
              搜索
            </button>
          </form>

          <div className="mt-6 flex flex-wrap gap-3">
            <Link
              href={dailyPost ? `/posts/${dailyPost.id}` : "/feed"}
              className="focus-ring inline-flex h-11 items-center gap-2 rounded-lg bg-marine-blue px-4 text-sm font-semibold text-marine-text shadow-sm transition hover:bg-[#4bbfee]"
            >
              <CalendarCheck size={18} />
              开始今日推荐
            </Link>
            <Link
              href="/topics"
              className="focus-ring inline-flex h-11 items-center gap-2 rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-4 text-sm font-semibold text-marine-deep transition hover:bg-marine-bg"
            >
              <Route size={18} />
              查看学习路径
            </Link>
            <Link
              href="/posts/new"
              className="focus-ring inline-flex h-11 items-center gap-2 rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-4 text-sm font-semibold text-marine-deep transition hover:bg-marine-bg"
            >
              <Edit3 size={18} />
              发布内容
            </Link>
          </div>
        </div>

        <DailyRecommendation post={dailyPost} hotTopic={hotTopic} />
      </section>

      <section className="mt-5 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <PracticeModule
          icon={Flame}
          title="每日推荐"
          description={dailyPost ? "从最新内容开始热身" : "暂无内容，先逛逛动态"}
          href={dailyPost ? `/posts/${dailyPost.id}` : "/feed"}
          accent="warm"
        />
        <PracticeModule
          icon={BookOpen}
          title="学习路径"
          description={hotTopic ? `围绕 #${hotTopic.name} 继续探索` : "按话题整理知识线索"}
          href={hotTopic ? `/search?keyword=${encodeURIComponent(hotTopic.name)}&type=post` : "/topics"}
          accent="mint"
        />
        <PracticeModule
          icon={MessageCircle}
          title="社区动态"
          description={`${formatCount(commentTotal)} 条当前页讨论`}
          href="/feed"
          accent="blue"
        />
        <PracticeModule
          icon={Trophy}
          title="创作打卡"
          description="把今天的理解写下来"
          href="/posts/new"
          accent="pink"
        />
      </section>

      <section className="mt-8 grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
        <div className="space-y-5">
          <div className="flex flex-col gap-3 border-b border-[rgba(77,100,124,0.12)] pb-4 md:flex-row md:items-end md:justify-between">
            <div>
              <p className="text-sm font-medium text-marine-deep">练习内容流</p>
              <h2 className="mt-1 text-2xl font-semibold text-marine-text">最新发布</h2>
            </div>
            <Link
              href="/feed"
              className="inline-flex items-center gap-1 text-sm font-semibold text-marine-deep transition hover:text-marine-text"
            >
              查看全部动态
              <ArrowRight size={16} />
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
              description="当前列表为空。可以先发布第一篇纯文本内容，或从话题页整理新的学习方向。"
            />
          ) : (
            <>
              <div className="space-y-4">
                {postList.map((post) => (
                  <PostCard key={post.id} post={post} />
                ))}
              </div>
              <Pagination page={page} total={postTotal} pageSize={10} basePath="/" />
            </>
          )}
        </div>

        <aside className="space-y-5">
          <section className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h2 className="text-base font-semibold text-marine-text">今日概览</h2>
                <p className="mt-1 text-xs text-marine-muted">根据当前内容页实时汇总</p>
              </div>
              <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-marine-mint/25 text-marine-deep">
                <BarChart3 size={20} />
              </span>
            </div>
            <div className="mt-5 grid grid-cols-2 gap-3">
              <Metric label="内容" value={formatCount(postTotal)} />
              <Metric label="讨论" value={formatCount(commentTotal)} />
              <Metric label="收藏" value={formatCount(collectionTotal)} />
              <Metric label="话题" value={formatCount(topics.length)} />
            </div>
          </section>

          <TopicPanel topics={topics} />
          <FeaturedPanel posts={featuredPosts} />
        </aside>
      </section>
    </main>
  );
}

function DailyRecommendation({
  post,
  hotTopic
}: {
  post?: PostInfo;
  hotTopic?: TopicInfo;
}) {
  return (
    <section className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft md:p-6">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-sm font-semibold text-marine-deep">今日推荐</p>
          <h2 className="mt-1 text-xl font-semibold text-marine-text">先完成一个小目标</h2>
        </div>
        <span className="inline-flex h-11 w-11 items-center justify-center rounded-lg bg-marine-warm/55 text-marine-text">
          <CalendarCheck size={22} />
        </span>
      </div>

      {post ? (
        <Link
          href={`/posts/${post.id}`}
          className="mt-5 block rounded-lg border border-[rgba(77,100,124,0.14)] bg-marine-bg/55 p-4 transition hover:bg-marine-bg"
        >
          <span className="inline-flex items-center gap-1 rounded-lg bg-white px-2.5 py-1 text-xs font-semibold text-marine-deep">
            <Flame size={13} />
            推荐阅读
          </span>
          <h3 className="mt-3 line-clamp-2 text-lg font-semibold leading-7 text-marine-text">
            {post.title || "未命名内容"}
          </h3>
          <p className="mt-2 line-clamp-3 text-sm leading-6 text-marine-muted">
            {post.summary || "这篇内容还没有摘要，打开详情继续阅读完整正文。"}
          </p>
          <div className="mt-4 flex flex-wrap items-center gap-3 text-xs text-marine-muted">
            <span className="inline-flex items-center gap-1">
              <MessageCircle size={14} />
              {formatCount(post.commentCount)} 讨论
            </span>
            <span className="inline-flex items-center gap-1">
              <Bookmark size={14} />
              {formatCount(post.collectionCount)} 收藏
            </span>
          </div>
        </Link>
      ) : (
        <div className="mt-5 rounded-lg border border-dashed border-[rgba(77,100,124,0.22)] bg-marine-bg/45 p-4">
          <h3 className="text-base font-semibold text-marine-text">暂无推荐内容</h3>
          <p className="mt-2 text-sm leading-6 text-marine-muted">
            内容列表为空时，首页仍保留搜索、话题和发布入口。可以先创建一篇内容作为今日练习。
          </p>
        </div>
      )}

      <div className="mt-4 flex flex-wrap gap-2">
        <Link
          href="/feed"
          className="focus-ring inline-flex h-10 flex-1 items-center justify-center gap-2 rounded-lg bg-marine-deep px-3 text-sm font-semibold text-white"
        >
          <Users size={17} />
          进入动态
        </Link>
        <Link
          href={hotTopic ? `/search?keyword=${encodeURIComponent(hotTopic.name)}&type=post` : "/topics"}
          className="focus-ring inline-flex h-10 flex-1 items-center justify-center gap-2 rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-3 text-sm font-semibold text-marine-deep"
        >
          <Tags size={17} />
          热门话题
        </Link>
      </div>
    </section>
  );
}

function PracticeModule({
  icon: Icon,
  title,
  description,
  href,
  accent
}: {
  icon: typeof Flame;
  title: string;
  description: string;
  href: string;
  accent: "warm" | "mint" | "blue" | "pink";
}) {
  const accentClass = {
    warm: "bg-marine-warm/50 text-marine-text",
    mint: "bg-marine-mint/25 text-marine-deep",
    blue: "bg-marine-bg text-marine-deep",
    pink: "bg-marine-pink/18 text-marine-deep"
  }[accent];

  return (
    <Link
      href={href}
      className="focus-ring group min-h-36 rounded-lg border border-[rgba(77,100,124,0.14)] bg-white p-5 shadow-soft transition hover:-translate-y-0.5 hover:shadow-float"
    >
      <span className={`inline-flex h-10 w-10 items-center justify-center rounded-lg ${accentClass}`}>
        <Icon size={20} />
      </span>
      <h2 className="mt-4 text-base font-semibold text-marine-text">{title}</h2>
      <p className="mt-2 min-h-10 text-sm leading-5 text-marine-muted">{description}</p>
      <span className="mt-3 inline-flex items-center gap-1 text-xs font-semibold text-marine-deep">
        继续
        <ArrowRight size={14} className="transition group-hover:translate-x-0.5" />
      </span>
    </Link>
  );
}

function TopicPanel({ topics }: { topics: TopicInfo[] }) {
  return (
    <section className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-marine-text">热门话题</h2>
          <p className="mt-1 text-xs text-marine-muted">按关键词快速进入练习</p>
        </div>
        <Tags size={18} className="text-marine-deep" />
      </div>
      <div className="mt-4 flex flex-wrap gap-2">
        {topics.length ? (
          topics.slice(0, 10).map((topic) => (
            <Link
              href={`/search?keyword=${encodeURIComponent(topic.name)}&type=post`}
              key={topic.id}
              className="rounded-lg bg-marine-bg px-3 py-2 text-xs font-medium text-marine-deep transition hover:bg-marine-blue/20"
            >
              #{topic.name}
            </Link>
          ))
        ) : (
          <p className="text-sm leading-6 text-marine-muted">暂无话题数据。</p>
        )}
      </div>
    </section>
  );
}

function FeaturedPanel({ posts }: { posts: PostInfo[] }) {
  return (
    <section className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-marine-text">精选内容</h2>
          <p className="mt-1 text-xs text-marine-muted">从当前列表挑选前三条</p>
        </div>
        <Trophy size={18} className="text-marine-warm" />
      </div>
      <div className="mt-4 space-y-3">
        {posts.length ? (
          posts.map((post) => (
            <Link
              key={post.id}
              href={`/posts/${post.id}`}
              className="block rounded-lg border border-[rgba(77,100,124,0.12)] p-3 transition hover:bg-marine-bg/70"
            >
              <p className="line-clamp-1 text-sm font-semibold text-marine-text">
                {post.title || "未命名内容"}
              </p>
              <p className="mt-2 inline-flex items-center gap-1 text-xs text-marine-muted">
                <MessageCircle size={13} />
                {formatCount(post.commentCount)} 条讨论
              </p>
            </Link>
          ))
        ) : (
          <p className="text-sm leading-6 text-marine-muted">暂无精选内容。</p>
        )}
      </div>
    </section>
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
