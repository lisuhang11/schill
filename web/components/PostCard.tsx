import Link from "next/link";
import { Bookmark, Eye, Heart, MessageCircle, Share2 } from "lucide-react";
import { formatCount, formatDate, splitTags, toBoolean, visibilityLabel } from "@/lib/format";
import type { FeedItem, PostInfo, SearchPostItem } from "@/lib/types";

type PostCardProps = {
  post: PostInfo | SearchPostItem | FeedItem;
  compact?: boolean;
};

function isSearchPost(post: PostCardProps["post"]): post is SearchPostItem {
  return "content" in post;
}

function isFeedItem(post: PostCardProps["post"]): post is FeedItem {
  return "author" in post && typeof (post as FeedItem).author === "object";
}

export function PostCard({ post, compact = false }: PostCardProps) {
  const id = isFeedItem(post) ? post.postId : isSearchPost(post) ? post.id : post.id;
  const title = isSearchPost(post)
    ? post.content.slice(0, 48) || "未命名内容"
    : post.title;
  const summary = isSearchPost(post) ? post.content : post.summary;
  const tags = isFeedItem(post)
    ? post.tags
    : isSearchPost(post)
    ? post.tags
    : splitTags(post.tags as string);
  const author = isFeedItem(post)
    ? post.author.username
    : isSearchPost(post)
    ? post.username
    : `用户 ${post.userId}`;
  const isTop = toBoolean(post.isTop);
  const isEssence = toBoolean(post.isEssence);

  return (
    <article className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft transition hover:-translate-y-0.5 hover:shadow-float">
      <div className="flex flex-wrap items-center gap-2 text-xs text-marine-muted">
        {isTop ? <span className="rounded-full bg-marine-warm/50 px-2 py-1 text-marine-text">置顶</span> : null}
        {isEssence ? <span className="rounded-full bg-marine-mint/30 px-2 py-1 text-marine-deep">精华</span> : null}
        <span>{visibilityLabel(post.visibility)}</span>
        <span>{formatDate(post.createdAt)}</span>
      </div>

      <Link href={`/posts/${id}`} className="mt-3 block">
        <h2 className="line-clamp-2 text-xl font-semibold leading-7 text-marine-text">
          {title}
        </h2>
        {!compact ? (
          <p className="mt-2 line-clamp-2 text-sm leading-6 text-marine-muted">
            {summary || "这篇内容暂时没有摘要，打开详情查看完整正文。"}
          </p>
        ) : null}
      </Link>

      <div className="mt-4 flex flex-wrap items-center gap-2">
        {tags.length ? (
          tags.slice(0, 4).map((tag) => (
            <span key={tag} className="rounded-full bg-marine-bg px-2.5 py-1 text-xs text-marine-deep">
              #{tag}
            </span>
          ))
        ) : (
          <span className="text-xs text-marine-muted/80">暂无标签</span>
        )}
      </div>

      <footer className="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-[rgba(77,100,124,0.12)] pt-4 text-sm text-marine-muted">
        <span>{author}</span>
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-1" title="评论">
            <MessageCircle size={16} />{formatCount(post.commentCount)}
          </span>
          <span className="inline-flex items-center gap-1" title="点赞">
            <Heart size={16} />{formatCount(post.upvoteCount)}
          </span>
          <span className="inline-flex items-center gap-1" title="收藏">
            <Bookmark size={16} />{formatCount(post.collectionCount)}
          </span>
          <span className="inline-flex items-center gap-1" title="分享">
            <Share2 size={16} />{formatCount(post.shareCount)}
          </span>
          <span className="sr-only"><Eye size={16} /></span>
        </div>
      </footer>
    </article>
  );
}
