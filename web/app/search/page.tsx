import Link from "next/link";
import { FileText, Hash, UserRound } from "lucide-react";
import { SearchForm } from "@/components/SearchForm";
import { Pagination } from "@/components/Pagination";
import { PostCard } from "@/components/PostCard";
import { StateBlock } from "@/components/StateBlock";
import { searchPosts, searchTopics, searchUsers } from "@/lib/api";
import { formatCount, formatDate } from "@/lib/format";
import type { SearchPostItem, SearchTopicItem, SearchUserItem } from "@/lib/types";

type SearchType = "post" | "user" | "topic";

type SearchPageProps = {
  searchParams: Promise<{ keyword?: string; type?: string; page?: string }>;
};

const searchTabs: Array<{
  type: SearchType;
  label: string;
  icon: typeof FileText;
}> = [
  { type: "post", label: "帖子", icon: FileText },
  { type: "user", label: "用户", icon: UserRound },
  { type: "topic", label: "话题", icon: Hash }
];

export default async function SearchPage({ searchParams }: SearchPageProps) {
  const params = await searchParams;
  const keyword = params.keyword?.trim() ?? "";
  const type: SearchType = params.type === "user" || params.type === "topic" ? params.type : "post";
  const page = Number(params.page ?? "1") || 1;

  const postResult = keyword && type === "post" ? await searchPosts({ keyword, page, pageSize: 10 }) : null;
  const userResult = keyword && type === "user" ? await searchUsers({ keyword, page, pageSize: 10 }) : null;
  const topicResult = keyword && type === "topic" ? await searchTopics({ keyword, page, pageSize: 10 }) : null;

  return (
    <main className="mx-auto w-full max-w-5xl px-4 py-8 md:px-8">
      <section className="mx-auto max-w-3xl">
        <SearchForm keyword={keyword} type={type} />
        <nav className="mt-4 flex gap-2 overflow-x-auto border-b border-[rgba(77,100,124,0.14)] pb-2" aria-label="搜索类型">
          {searchTabs.map((tab) => (
            <SearchTab key={tab.type} tab={tab} active={tab.type === type} keyword={keyword} />
          ))}
        </nav>
      </section>

      <section className="mt-6 space-y-4">
        {!keyword ? (
          <StateBlock tone="info" title="输入关键词开始搜索" />
        ) : type === "post" ? (
          <PostSearchResults result={postResult} page={page} keyword={keyword} />
        ) : type === "user" ? (
          <UserSearchResults result={userResult} page={page} keyword={keyword} />
        ) : (
          <TopicSearchResults result={topicResult} page={page} keyword={keyword} />
        )}
      </section>
    </main>
  );
}

function SearchTab({
  tab,
  active,
  keyword
}: {
  tab: (typeof searchTabs)[number];
  active: boolean;
  keyword: string;
}) {
  const Icon = tab.icon;
  const href = `/search?${new URLSearchParams({ keyword, type: tab.type }).toString()}`;

  return (
    <Link
      href={href}
      className={`focus-ring flex min-w-24 items-center justify-center gap-2 rounded-lg px-3 py-2 text-sm font-semibold transition ${
        active
          ? "bg-marine-deep text-white shadow-sm"
          : "bg-white text-marine-muted hover:bg-marine-bg hover:text-marine-deep"
      }`}
    >
      <Icon size={17} />
      {tab.label}
    </Link>
  );
}

function PostSearchResults({
  result,
  page,
  keyword
}: {
  result: Awaited<ReturnType<typeof searchPosts>> | null;
  page: number;
  keyword: string;
}) {
  if (!result?.ok) {
    return <StateBlock tone="error" title="搜索失败" description={result?.message ?? "未知错误"} />;
  }
  if (result.data.list.length === 0) {
    return <StateBlock tone="empty" title="没有帖子结果" />;
  }
  return (
    <>
      <div className="space-y-4">
        {result.data.list.map((item: SearchPostItem) => (
          <PostCard key={item.id} post={item} />
        ))}
      </div>
      <Pagination page={page} total={result.data.total} pageSize={10} basePath="/search" query={{ keyword, type: "post" }} />
    </>
  );
}

function UserSearchResults({
  result,
  page,
  keyword
}: {
  result: Awaited<ReturnType<typeof searchUsers>> | null;
  page: number;
  keyword: string;
}) {
  if (!result?.ok) {
    return <StateBlock tone="error" title="搜索失败" description={result?.message ?? "未知错误"} />;
  }
  if (result.data.list.length === 0) {
    return <StateBlock tone="empty" title="没有用户结果" />;
  }
  return (
    <>
      <div className="grid gap-4 md:grid-cols-2">
        {result.data.list.map((item: SearchUserItem) => (
          <UserResultCard key={item.id} item={item} />
        ))}
      </div>
      <Pagination page={page} total={result.data.total} pageSize={10} basePath="/search" query={{ keyword, type: "user" }} />
    </>
  );
}

function TopicSearchResults({
  result,
  page,
  keyword
}: {
  result: Awaited<ReturnType<typeof searchTopics>> | null;
  page: number;
  keyword: string;
}) {
  if (!result?.ok) {
    return <StateBlock tone="error" title="搜索失败" description={result?.message ?? "未知错误"} />;
  }
  if (result.data.list.length === 0) {
    return <StateBlock tone="empty" title="没有话题结果" />;
  }
  return (
    <>
      <div className="grid gap-4 md:grid-cols-2">
        {result.data.list.map((item: SearchTopicItem) => (
          <TopicResultCard key={item.id} item={item} />
        ))}
      </div>
      <Pagination page={page} total={result.data.total} pageSize={10} basePath="/search" query={{ keyword, type: "topic" }} />
    </>
  );
}

function UserResultCard({ item }: { item: SearchUserItem }) {
  return (
    <Link
      href={`/users/${item.id}`}
      className="focus-ring block rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft transition hover:-translate-y-0.5 hover:shadow-float"
    >
      <div className="flex items-start gap-3">
        <span className="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-marine-mint/25 text-marine-deep">
          <UserRound size={21} />
        </span>
        <div className="min-w-0">
          <p className="truncate text-lg font-semibold text-marine-text">{item.username}</p>
          <p className="mt-1 text-xs text-marine-muted">最近活跃 {formatDate(item.lastActiveTime)}</p>
        </div>
      </div>
      <div className="mt-4 grid grid-cols-3 gap-2 text-center">
        <SmallMetric label="帖子" value={item.postCount} />
        <SmallMetric label="粉丝" value={item.followerCount} />
        <SmallMetric label="获赞" value={item.likeCount} />
      </div>
    </Link>
  );
}

function TopicResultCard({ item }: { item: SearchTopicItem }) {
  return (
    <Link
      href={`/search?keyword=${encodeURIComponent(item.name)}&type=post`}
      className="focus-ring block rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft transition hover:-translate-y-0.5 hover:shadow-float"
    >
      <p className="truncate text-lg font-semibold text-marine-text">#{item.name}</p>
      <p className="mt-2 text-sm leading-6 text-marine-muted">
        已被引用 {formatCount(item.quoteNum)} 次
      </p>
    </Link>
  );
}

function SmallMetric({ label, value }: { label: string; value: number }) {
  return (
    <span className="rounded-lg bg-marine-bg px-2 py-2">
      <span className="block text-sm font-semibold text-marine-text">{formatCount(value)}</span>
      <span className="mt-1 block text-xs text-marine-muted">{label}</span>
    </span>
  );
}
