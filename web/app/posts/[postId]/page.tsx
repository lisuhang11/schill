import { notFound } from "next/navigation";
import { getCommentList, getPostDetail } from "@/lib/api";
import { formatDate, splitTags, toBoolean, visibilityLabel } from "@/lib/format";
import { CommentComposer } from "@/components/CommentComposer";
import { CommentList } from "@/components/CommentList";
import { PostActions } from "@/components/PostActions";
import { StateBlock } from "@/components/StateBlock";

type PostDetailPageProps = {
  params: Promise<{ postId: string }>;
};

export default async function PostDetailPage({ params }: PostDetailPageProps) {
  const { postId: rawPostId } = await params;
  const postId = Number(rawPostId);
  if (!Number.isFinite(postId)) {
    notFound();
  }

  const [detailResult, commentsResult] = await Promise.all([
    getPostDetail(postId),
    getCommentList({ postId, pageSize: 20, sortType: "time" })
  ]);

  if (!detailResult.ok) {
    return (
      <main className="mx-auto w-full max-w-4xl px-4 py-8 md:px-8">
        <StateBlock tone="error" title="文章加载失败" description={detailResult.message} />
      </main>
    );
  }

  const { post, contents, topics } = detailResult.data;
  const tags = splitTags(post.tags);

  return (
    <main className="mx-auto w-full max-w-4xl px-4 py-8 md:px-8">
      <article className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-6 shadow-soft">
        <div className="flex flex-wrap items-center gap-2 text-xs text-marine-muted">
          {toBoolean(post.isTop) ? <span className="rounded-full bg-marine-warm/50 px-2 py-1 text-marine-text">置顶</span> : null}
          {toBoolean(post.isEssence) ? <span className="rounded-full bg-marine-mint/30 px-2 py-1 text-marine-deep">精华</span> : null}
          {toBoolean(post.isLock) ? <span className="rounded-full bg-slate-100 px-2 py-1 text-slate-600">锁定</span> : null}
          <span>{visibilityLabel(post.visibility)}</span>
          <span>{formatDate(post.createdAt)}</span>
        </div>

        <h1 className="mt-4 text-3xl font-semibold leading-tight text-marine-text">{post.title}</h1>
        <p className="mt-3 text-sm text-marine-muted">作者 ID {post.userId}</p>

        <div className="mt-5 flex flex-wrap gap-2">
          {topics.map((topic) => (
            <span key={topic.topicId} className="rounded-full bg-marine-bg px-3 py-1 text-sm text-marine-deep">
              #{topic.topicName}
            </span>
          ))}
          {tags.map((tag) => (
            <span key={tag} className="rounded-full bg-marine-warm/35 px-3 py-1 text-sm text-marine-text">
              {tag}
            </span>
          ))}
        </div>

        <div className="mt-8 space-y-5 text-base leading-8 text-marine-text">
          {contents.length ? (
            contents
              .filter((item) => item.type === 2)
              .sort((a, b) => a.sort - b.sort)
              .map((item, index) => <p key={`${item.sort}-${index}`}>{item.content}</p>)
          ) : (
            <p className="text-marine-muted">这篇文章暂无正文内容。</p>
          )}
        </div>

        <PostActions postId={post.id} counts={post} />
      </article>

      <section className="mt-6 rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-6 shadow-soft">
        <h2 className="text-xl font-semibold text-marine-text">评论</h2>
        <CommentComposer postId={post.id} />
        {!commentsResult.ok ? (
          <div className="mt-5">
            <StateBlock tone="error" title="评论加载失败" description={commentsResult.message} />
          </div>
        ) : commentsResult.data.list.length === 0 ? (
          <div className="mt-5">
            <StateBlock tone="empty" title="暂无评论" description="可以发布第一条评论。" />
          </div>
        ) : (
          <CommentList comments={commentsResult.data.list} />
        )}
      </section>
    </main>
  );
}
