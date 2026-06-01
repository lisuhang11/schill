"use client";

import { useState, useTransition } from "react";
import { MessageCircle, ThumbsDown, ThumbsUp } from "lucide-react";
import { voteComment } from "@/lib/api";
import { formatDate } from "@/lib/format";
import type { CommentItem } from "@/lib/types";
import { CommentComposer } from "@/components/CommentComposer";

export function CommentList({ comments, postId }: { comments: CommentItem[]; postId: number }) {
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();
  const [replyTarget, setReplyTarget] = useState<{
    commentId: number;
    userId: number;
    username: string;
  } | null>(null);

  function vote(commentId: number, voteType: 1 | 2) {
    setMessage("");
    startTransition(async () => {
      const result = await voteComment(commentId, voteType);
      setMessage(result.ok ? "投票已提交，刷新后查看最新计数。" : result.message);
    });
  }

  function handleReply(commentId: number, userId: number, username: string) {
    setReplyTarget({ commentId, userId, username });
  }

  function cancelReply() {
    setReplyTarget(null);
  }

  return (
    <div className="mt-6 space-y-4">
      {comments.map((item) => (
        <article key={item.root.id} className="rounded-lg bg-marine-bg p-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p className="font-semibold text-marine-text">{item.root.username || `用户 ${item.root.userId}`}</p>
              <p className="text-xs text-marine-muted">{formatDate(item.root.createdAt)}</p>
            </div>
            <div className="flex gap-2">
              <button
                type="button"
                disabled={isPending}
                onClick={() => vote(item.root.id, 1)}
                className="focus-ring inline-flex items-center gap-1 rounded-lg bg-white px-3 py-1.5 text-xs text-marine-deep disabled:opacity-60"
              >
                <ThumbsUp size={14} /> {item.root.likeCount}
              </button>
              <button
                type="button"
                disabled={isPending}
                onClick={() => vote(item.root.id, 2)}
                className="focus-ring inline-flex items-center gap-1 rounded-lg bg-white px-3 py-1.5 text-xs text-marine-muted disabled:opacity-60"
              >
                <ThumbsDown size={14} /> {item.root.dislikeCount}
              </button>
              <button
                type="button"
                onClick={() => handleReply(item.root.id, item.root.userId, item.root.username || `用户 ${item.root.userId}`)}
                className="focus-ring inline-flex items-center gap-1 rounded-lg bg-white px-3 py-1.5 text-xs text-marine-muted hover:text-marine-deep disabled:opacity-60"
              >
                <MessageCircle size={14} /> 回复
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

          {item.replies.length ? (
            <div className="mt-3 space-y-2 border-l-2 border-marine-blue/40 pl-3">
              {item.replies.slice(0, 3).map((reply) => (
                <div key={reply.id}>
                  <p className="text-sm leading-6 text-marine-muted">
                    <span className="font-medium text-marine-text">{reply.username || `用户 ${reply.userId}`}：</span>
                    {reply.content}
                  </p>
                  <button
                    type="button"
                    onClick={() => handleReply(reply.id, reply.userId, reply.username || `用户 ${reply.userId}`)}
                    className="mt-1 inline-flex items-center gap-1 text-xs text-marine-muted hover:text-marine-deep"
                  >
                    <MessageCircle size={12} /> 回复
                  </button>
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
            </div>
          ) : null}
        </article>
      ))}
      {message ? <p className="text-sm text-marine-deep">{message}</p> : null}
    </div>
  );
}
