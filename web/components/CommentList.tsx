"use client";

import { useCallback, useState, useTransition } from "react";
import { ChevronDown, ChevronUp, Heart, MessageCircle, MoreHorizontal, Trash2, X } from "lucide-react";
import { deleteComment, getReplyList, voteComment } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { formatCount, formatDate } from "@/lib/format";
import type { CommentInfo, CommentItem } from "@/lib/types";
import { CommentComposer } from "@/components/CommentComposer";
import { StateBlock } from "@/components/StateBlock";

export function CommentList({
  comments: initialComments,
  postId
}: {
  comments: CommentItem[];
  postId: number;
}) {
  const { userId, username } = useAuth();
  const [comments, setComments] = useState(initialComments);
  const [isOpen, setIsOpen] = useState(false);
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();
  const [replyTarget, setReplyTarget] = useState<{
    parentId: number;
    replyToUserId: number;
    username: string;
  } | null>(null);
  const [loadingReplies, setLoadingReplies] = useState<Record<number, boolean>>({});
  const visibleCommentCount = comments.reduce(
    (sum, item) => sum + 1 + item.replies.length,
    0
  );

  function withLocalUser(comment: CommentInfo): CommentInfo {
    return {
      ...comment,
      username: comment.username || username || (userId ? `用户 ${userId}` : "我"),
      avatar: comment.avatar || "",
      replyToUsername: comment.replyToUsername || replyTarget?.username || "",
      isLiked: comment.isLiked ?? false,
      isDisliked: comment.isDisliked ?? false
    };
  }

  function handleRootCreated(comment: CommentInfo) {
    const nextComment = withLocalUser(comment);
    setComments((prev) => [
      { root: nextComment, replies: [], hasMoreReplies: false },
      ...prev
    ]);
  }

  function handleReplyCreated(comment: CommentInfo) {
    const nextReply = withLocalUser(comment);
    const parentId = replyTarget?.parentId ?? comment.parentId;
    setComments((prev) =>
      prev.map((item) => {
        if (item.root.id !== parentId) return item;
        return {
          ...item,
          root: { ...item.root, replyCount: item.root.replyCount + 1 },
          replies: [...item.replies, nextReply]
        };
      })
    );
    setReplyTarget(null);
  }

  function vote(commentId: number, voteType: 1 | 2) {
    setMessage("");
    startTransition(async () => {
      const result = await voteComment(commentId, voteType);
      if (!result.ok) {
        setMessage(result.message);
        return;
      }

      setComments((prev) =>
        prev.map((item) => {
          if (item.root.id === commentId) {
            return { ...item, root: applyVote(item.root, result.data) };
          }
          return {
            ...item,
            replies: item.replies.map((reply) =>
              reply.id === commentId ? applyVote(reply, result.data) : reply
            )
          };
        })
      );
    });
  }

  function handleDelete(commentId: number) {
    if (!confirm("确定要删除这条评论吗？")) return;
    setMessage("");
    startTransition(async () => {
      const result = await deleteComment(commentId);
      if (!result.ok) {
        setMessage(result.message);
        return;
      }

      setComments((prev) =>
        prev
          .filter((item) => item.root.id !== commentId)
          .map((item) => {
            const replies = item.replies.filter((reply) => reply.id !== commentId);
            const removedCount = item.replies.length - replies.length;
            return {
              ...item,
              replies,
              root: { ...item.root, replyCount: Math.max(0, item.root.replyCount - removedCount) }
            };
          })
      );
      setMessage("评论已删除");
    });
  }

  const loadMoreReplies = useCallback(async (rootIdx: number, commentId: number) => {
    setLoadingReplies((prev) => ({ ...prev, [commentId]: true }));
    try {
      const cursor = comments[rootIdx]?.replies.at(-1)?.id ?? 0;
      const result = await getReplyList({ commentId, cursor, pageSize: 20 });
      if (result.ok) {
        setComments((prev) => {
          const next = prev.map((item) => ({ ...item, root: { ...item.root }, replies: [...item.replies] }));
          const existingIds = new Set(next[rootIdx].replies.map((reply) => reply.id));
          for (const reply of result.data.list) {
            if (!existingIds.has(reply.id)) {
              next[rootIdx].replies.push(reply);
            }
          }
          next[rootIdx].hasMoreReplies = result.data.hasMore;
          next[rootIdx].root.replyCount = Math.max(next[rootIdx].root.replyCount, next[rootIdx].replies.length);
          return next;
        });
      } else {
        setMessage(result.message);
      }
    } finally {
      setLoadingReplies((prev) => ({ ...prev, [commentId]: false }));
    }
  }, [comments]);

  return (
    <div className="mt-3 rounded-lg border border-[rgba(77,100,124,0.14)] bg-white">
      <button
        type="button"
        onClick={() => setIsOpen((value) => !value)}
        className="focus-ring flex w-full items-center justify-between gap-3 rounded-lg px-4 py-3 text-left"
      >
        <span className="inline-flex min-w-0 items-center gap-2">
          <MessageCircle size={18} className="shrink-0 text-marine-deep" />
          <span className="truncate text-sm font-semibold text-marine-text">评论区</span>
          <span className="rounded-lg bg-marine-bg px-2 py-0.5 text-xs text-marine-deep">{visibleCommentCount} 条</span>
        </span>
        <span className="inline-flex items-center gap-1 text-xs font-medium text-marine-muted">
          {isOpen ? "收起" : "展开"}
          {isOpen ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
        </span>
      </button>

      {isOpen ? (
        <div className="border-t border-[rgba(77,100,124,0.10)] px-4 pb-4">
          <div className="pt-3">
            <CommentComposer postId={postId} onSuccess={handleRootCreated} />
          </div>

          {comments.length === 0 ? (
            <div className="mt-4">
              <StateBlock tone="empty" title="暂无评论" description="可以发布第一条评论。" />
            </div>
          ) : (
            <div className="mt-4 max-h-[520px] divide-y divide-[rgba(77,100,124,0.10)] overflow-y-auto pr-1">
              {comments.map((item, rootIdx) => (
                <article key={item.root.id} className="py-4">
                  <CommentRow
                    comment={item.root}
                    pending={isPending}
                    onVote={() => vote(item.root.id, 1)}
                    onReply={() =>
                      setReplyTarget({
                        parentId: item.root.id,
                        replyToUserId: item.root.userId,
                        username: item.root.username || `用户 ${item.root.userId}`
                      })
                    }
                    onDelete={() => handleDelete(item.root.id)}
                  />

                  {replyTarget?.parentId === item.root.id ? (
                    <div className="ml-12 mt-3 rounded-lg bg-marine-bg/70 p-3">
                      <button
                        type="button"
                        onClick={() => setReplyTarget(null)}
                        className="mb-2 inline-flex items-center gap-1 text-xs text-marine-muted hover:text-red-600"
                      >
                        <X size={14} /> 取消回复
                      </button>
                      <CommentComposer
                        postId={postId}
                        parentId={item.root.id}
                        replyToUserId={replyTarget.replyToUserId}
                        replyToUsername={replyTarget.username}
                        placeholder={`回复 ${replyTarget.username}`}
                        onCancel={() => setReplyTarget(null)}
                        onSuccess={handleReplyCreated}
                      />
                    </div>
                  ) : null}

                  {item.replies.length > 0 ? (
                    <div className="ml-12 mt-3 space-y-3 rounded-lg bg-marine-bg/60 p-3">
                      {item.replies.map((reply) => (
                        <CommentRow
                          key={reply.id}
                          comment={reply}
                          compact
                          pending={isPending}
                          onVote={() => vote(reply.id, 1)}
                          onReply={() =>
                            setReplyTarget({
                              parentId: item.root.id,
                              replyToUserId: reply.userId,
                              username: reply.username || `用户 ${reply.userId}`
                            })
                          }
                          onDelete={() => handleDelete(reply.id)}
                        />
                      ))}

                      {item.hasMoreReplies ? (
                        <button
                          type="button"
                          disabled={loadingReplies[item.root.id]}
                          onClick={() => loadMoreReplies(rootIdx, item.root.id)}
                          className="inline-flex items-center gap-1 text-xs font-medium text-marine-deep hover:text-marine-text disabled:opacity-60"
                        >
                          <ChevronDown size={14} />
                          {loadingReplies[item.root.id]
                            ? "加载中..."
                            : `展开更多回复 (${Math.max(0, item.root.replyCount - item.replies.length)} 条)`}
                        </button>
                      ) : null}
                    </div>
                  ) : null}
                </article>
              ))}
            </div>
          )}

          {message ? <p className="mt-4 text-sm text-marine-deep">{message}</p> : null}
        </div>
      ) : null}
    </div>
  );
}

function CommentRow({
  comment,
  compact = false,
  pending,
  onVote,
  onReply,
  onDelete
}: {
  comment: CommentInfo;
  compact?: boolean;
  pending: boolean;
  onVote: () => void;
  onReply: () => void;
  onDelete: () => void;
}) {
  const displayName = comment.username || `用户 ${comment.userId}`;
  const avatarText = displayName.slice(0, 1).toUpperCase();

  return (
    <div className="flex gap-3">
      <div className={`${compact ? "h-8 w-8 text-xs" : "h-10 w-10 text-sm"} flex shrink-0 items-center justify-center rounded-full bg-marine-deep font-semibold text-white`}>
        {comment.avatar ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={comment.avatar} alt={displayName} className="h-full w-full rounded-full object-cover" />
        ) : (
          avatarText
        )}
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold text-marine-text">{displayName}</p>
            <p className="text-xs text-marine-muted">{formatDate(comment.createdAt)}</p>
          </div>
          <button type="button" className="text-marine-muted hover:text-marine-text" aria-label="更多">
            <MoreHorizontal size={18} />
          </button>
        </div>
        <p className={`${compact ? "mt-1 text-sm" : "mt-2 text-[15px]"} break-words leading-6 text-marine-text`}>
          {comment.replyToUsername ? <span className="mr-1 text-marine-muted">回复 {comment.replyToUsername}</span> : null}
          {comment.content}
        </p>
        <div className="mt-2 flex items-center gap-4 text-xs text-marine-muted">
          <button
            type="button"
            disabled={pending}
            onClick={onVote}
            className={`inline-flex items-center gap-1 font-medium transition hover:text-marine-pink disabled:opacity-60 ${
              comment.isLiked ? "text-marine-pink" : ""
            }`}
          >
            <Heart size={14} fill={comment.isLiked ? "currentColor" : "none"} />
            {formatCount(comment.likeCount)}
          </button>
          <button type="button" onClick={onReply} className="inline-flex items-center gap-1 hover:text-marine-deep">
            <MessageCircle size={14} /> 回复
          </button>
          <button
            type="button"
            disabled={pending}
            onClick={onDelete}
            className="inline-flex items-center gap-1 hover:text-red-600 disabled:opacity-60"
          >
            <Trash2 size={14} /> 删除
          </button>
        </div>
      </div>
    </div>
  );
}

function applyVote(comment: CommentInfo, vote: {
  likeCount: number;
  dislikeCount: number;
  isLiked: boolean;
  isDisliked: boolean;
}): CommentInfo {
  return {
    ...comment,
    likeCount: vote.likeCount,
    dislikeCount: vote.dislikeCount,
    isLiked: vote.isLiked,
    isDisliked: vote.isDisliked
  };
}
