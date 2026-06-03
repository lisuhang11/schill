"use client";

import { useCallback, useState, useTransition } from "react";
import { MessageCircle, ThumbsDown, ThumbsUp, Trash2, ChevronDown } from "lucide-react";
import { voteComment, deleteComment, getReplyList } from "@/lib/api";
import { formatDate } from "@/lib/format";
import type { CommentInfo, CommentItem } from "@/lib/types";
import { CommentComposer } from "@/components/CommentComposer";

export function CommentList({
  comments: initialComments,
  postId
}: {
  comments: CommentItem[];
  postId: number;
}) {
  const [comments, setComments] = useState(initialComments);
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();
  const [replyTarget, setReplyTarget] = useState<{
    commentId: number;
    userId: number;
    username: string;
  } | null>(null);
  // Track loading state per-comment for "load more replies"
  const [loadingReplies, setLoadingReplies] = useState<Record<number, boolean>>({});

  function vote(commentId: number, voteType: 1 | 2, isRoot: boolean, rootIdx?: number) {
    setMessage("");
    startTransition(async () => {
      const result = await voteComment(commentId, voteType);
      if (result.ok) {
        setComments((prev) => {
          const next = prev.map((item) => ({ ...item, root: { ...item.root }, replies: [...item.replies] }));
          if (isRoot && rootIdx !== undefined) {
            const root = next[rootIdx].root;
            if (root.id === commentId) {
              root.likeCount = result.data.likeCount;
              root.dislikeCount = result.data.dislikeCount;
              root.isLiked = result.data.isLiked;
              root.isDisliked = result.data.isDisliked;
            }
          } else {
            // Update in replies list
            for (const item of next) {
              for (const reply of item.replies) {
                if (reply.id === commentId) {
                  reply.likeCount = result.data.likeCount;
                  reply.dislikeCount = result.data.dislikeCount;
                  reply.isLiked = result.data.isLiked;
                  reply.isDisliked = result.data.isDisliked;
                }
              }
            }
          }
          return next;
        });
      } else {
        setMessage(result.message);
      }
    });
  }

  function handleDelete(commentId: number) {
    if (!confirm("确定要删除这条评论吗？")) return;
    setMessage("");
    startTransition(async () => {
      const result = await deleteComment(commentId);
      if (result.ok) {
        setComments((prev) =>
          prev
            .filter((item) => item.root.id !== commentId)
            .map((item) => ({
              ...item,
              replies: item.replies.filter((r) => r.id !== commentId)
            }))
        );
        setMessage("评论已删除");
      } else {
        setMessage(result.message);
      }
    });
  }

  const loadMoreReplies = useCallback(
    async (rootIdx: number, commentId: number, cursor: number) => {
      setLoadingReplies((prev) => ({ ...prev, [commentId]: true }));
      try {
        const result = await getReplyList({ commentId, cursor, pageSize: 20 });
        if (result.ok) {
          setComments((prev) => {
            const next = prev.map((item) => ({ ...item, root: { ...item.root }, replies: [...item.replies] }));
            const existingIds = new Set(next[rootIdx].replies.map((r) => r.id));
            for (const r of result.data.list) {
              if (!existingIds.has(r.id)) {
                next[rootIdx].replies.push(r);
              }
            }
            next[rootIdx].hasMoreReplies = result.data.hasMore;
            next[rootIdx].root.replyCount = Math.max(
              next[rootIdx].root.replyCount,
              next[rootIdx].replies.length
            );
            return next;
          });
        }
      } finally {
        setLoadingReplies((prev) => ({ ...prev, [commentId]: false }));
      }
    },
    []
  );

  function handleReply(commentId: number, userId: number, username: string) {
    setReplyTarget({ commentId, userId, username });
  }

  function cancelReply() {
    setReplyTarget(null);
  }

  function voteButtonClass(isActive: boolean) {
    return `focus-ring inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs disabled:opacity-60 ${
      isActive ? "bg-marine-deep text-white" : "bg-white text-marine-deep"
    }`;
  }

  return (
    <div className="mt-6 space-y-4">
      {comments.map((item, rootIdx) => (
        <article key={item.root.id} className="rounded-lg bg-marine-bg p-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p className="font-semibold text-marine-text">
                {item.root.username || `用户 ${item.root.userId}`}
              </p>
              <p className="text-xs text-marine-muted">{formatDate(item.root.createdAt)}</p>
            </div>
            <div className="flex gap-2">
              <button
                type="button"
                disabled={isPending}
                onClick={() => vote(item.root.id, 1, true, rootIdx)}
                className={voteButtonClass(item.root.isLiked)}
              >
                <ThumbsUp size={14} /> {item.root.likeCount}
              </button>
              <button
                type="button"
                disabled={isPending}
                onClick={() => vote(item.root.id, 2, true, rootIdx)}
                className={voteButtonClass(item.root.isDisliked)}
              >
                <ThumbsDown size={14} /> {item.root.dislikeCount}
              </button>
              <button
                type="button"
                onClick={() =>
                  handleReply(item.root.id, item.root.userId, item.root.username || `用户 ${item.root.userId}`)
                }
                className="focus-ring inline-flex items-center gap-1 rounded-lg bg-white px-3 py-1.5 text-xs text-marine-muted hover:text-marine-deep"
              >
                <MessageCircle size={14} /> 回复
              </button>
              <button
                type="button"
                disabled={isPending}
                onClick={() => handleDelete(item.root.id)}
                className="focus-ring inline-flex items-center gap-1 rounded-lg bg-white px-3 py-1.5 text-xs text-red-500 hover:text-red-700 disabled:opacity-60"
              >
                <Trash2 size={14} />
              </button>
            </div>
          </div>
          <p className="mt-3 text-sm leading-6 text-marine-text">{item.root.content}</p>

          {replyTarget?.commentId === item.root.id ? (
            <CommentComposer
              postId={postId}
              parentId={item.root.id}
              replyToUserId={replyTarget.userId}
              replyToUsername={replyTarget.username}
              placeholder={`回复 ${replyTarget.username}`}
              onCancel={cancelReply}
              onSuccess={cancelReply}
            />
          ) : null}

          {item.replies.length > 0 ? (
            <div className="mt-3 space-y-2 border-l-2 border-marine-blue/40 pl-3">
              {item.replies.map((reply) => (
                <div key={reply.id}>
                  <p className="text-sm leading-6 text-marine-muted">
                    <span className="font-medium text-marine-text">
                      {reply.username || `用户 ${reply.userId}`}
                      {reply.replyToUsername ? (
                        <span className="text-marine-muted font-normal">
                          {" "}
                          回复 {reply.replyToUsername}
                        </span>
                      ) : null}
                      ：
                    </span>
                    {reply.content}
                  </p>
                  <div className="mt-1 flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() =>
                        handleReply(reply.id, reply.userId, reply.username || `用户 ${reply.userId}`)
                      }
                      className="inline-flex items-center gap-1 text-xs text-marine-muted hover:text-marine-deep"
                    >
                      <MessageCircle size={12} /> 回复
                    </button>
                    <button
                      type="button"
                      disabled={isPending}
                      onClick={() => vote(reply.id, 1, false)}
                      className={`inline-flex items-center gap-1 text-xs ${
                        reply.isLiked ? "text-marine-deep" : "text-marine-muted"
                      } hover:text-marine-deep disabled:opacity-60`}
                    >
                      <ThumbsUp size={12} /> {reply.likeCount}
                    </button>
                    <button
                      type="button"
                      disabled={isPending}
                      onClick={() => handleDelete(reply.id)}
                      className="inline-flex items-center gap-1 text-xs text-red-400 hover:text-red-600 disabled:opacity-60"
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>
                  {replyTarget?.commentId === reply.id ? (
                    <CommentComposer
                      postId={postId}
                      parentId={reply.id}
                      replyToUserId={replyTarget.userId}
                      replyToUsername={replyTarget.username}
                      placeholder={`回复 ${replyTarget.username}`}
                      onCancel={cancelReply}
                      onSuccess={cancelReply}
                    />
                  ) : null}
                </div>
              ))}

              {item.hasMoreReplies ? (
                <button
                  type="button"
                  disabled={loadingReplies[item.root.id]}
                  onClick={() =>
                    loadMoreReplies(rootIdx, item.root.id, item.root.id)
                  }
                  className="inline-flex items-center gap-1 text-xs text-marine-deep hover:text-marine-text disabled:opacity-60"
                >
                  <ChevronDown size={14} />
                  {loadingReplies[item.root.id] ? "加载中..." : `加载更多回复 (${item.root.replyCount - item.replies.length} 条)`}
                </button>
              ) : null}
            </div>
          ) : null}
        </article>
      ))}
      {message ? <p className="text-sm text-marine-deep">{message}</p> : null}
    </div>
  );
}
