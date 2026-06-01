import Link from "next/link";
import { SearchForm } from "@/components/SearchForm";
import { Pagination } from "@/components/Pagination";
import { PostCard } from "@/components/PostCard";
import { StateBlock } from "@/components/StateBlock";
import { searchPosts, searchTopics, searchUsers } from "@/lib/api";
import type { SearchPostItem, SearchTopicItem, SearchUserItem } from "@/lib/types";

type SearchPageProps = {
  searchParams: Promise<{ keyword?: string; type?: string; page?: string }>;
};

export default async function SearchPage({ searchParams }: SearchPageProps) {
  const params = await searchParams;
  const keyword = params.keyword?.trim() ?? "";
  const type = params.type === "user" || params.type === "topic" ? params.type : "post";
  const page = Number(params.page ?? "1") || 1;

  const postResult = keyword && type === "post" ? await searchPosts({ keyword, page, pageSize: 10 }) : null;
  const userResult = keyword && type === "user" ? await searchUsers({ keyword, page, pageSize: 10 }) : null;
  const topicResult = keyword && type === "topic" ? await searchTopics({ keyword, page, pageSize: 10 }) : null;

  return (
    <main className="mx-auto w-full max-w-6xl px-4 py-8 md:px-8">
      <div className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-6 shadow-soft">
        <h1 className="text-3xl font-semibold text-marine-text">搜索社区</h1>
        <p className="mt-2 text-sm leading-6 text-marine-muted">
          支持搜索文章、用户和话题，接口来自 `service/search/api/search.api`。
        </p>
        <SearchForm keyword={keyword} type={type} />
      </div>

      <section className="mt-6 space-y-4">
        {!keyword ? (
          <StateBlock tone="info" title="输入关键词开始搜索" description="关键词不能为空，页大小默认 10 条。" />
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
    return <StateBlock tone="empty" title="没有结果" description="换一个关键词或结果类型再试。" />;
  }
  return (
    <>
      {result.data.list.map((item: SearchPostItem) => (
        <PostCard key={item.id} post={item} />
      ))}
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
    return <StateBlock tone="empty" title="没有结果" description="换一个关键词或结果类型再试。" />;
  }
  return (
    <>
      <div className="grid gap-4 md:grid-cols-2">
        {result.data.list.map((item: SearchUserItem) => (
          <SearchEntityCard
            key={item.id}
            href={`/users/${item.id}`}
            title={item.username}
            description={`粉丝 ${item.followerCount} · 文章 ${item.postCount} · 获赞 ${item.likeCount}`}
          />
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
    return <StateBlock tone="empty" title="没有结果" description="换一个关键词或结果类型再试。" />;
  }
  return (
    <>
      <div className="grid gap-4 md:grid-cols-2">
        {result.data.list.map((item: SearchTopicItem) => (
          <SearchEntityCard key={item.id} title={`#${item.name}`} description={`引用 ${item.quoteNum}`} />
        ))}
      </div>
      <Pagination page={page} total={result.data.total} pageSize={10} basePath="/search" query={{ keyword, type: "topic" }} />
    </>
  );
}

function SearchEntityCard({ title, description, href }: { title: string; description: string; href?: string }) {
  const content = (
    <div className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft transition hover:-translate-y-0.5 hover:shadow-float">
      <p className="text-lg font-semibold text-marine-text">{title}</p>
      <p className="mt-2 text-sm text-marine-muted">{description}</p>
    </div>
  );

  if (href) {
    return <Link href={href}>{content}</Link>;
  }
  return content;
}
