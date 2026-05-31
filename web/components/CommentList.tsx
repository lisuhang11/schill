"use client";

import { useState, useTransition } from "react";
import { ThumbsDown, ThumbsUp } from "lucide-react";
import { voteComment } from "@/lib/api";
import { formatDate } from "@/lib/format";
import type { CommentItem } from "@/lib/types";

export function CommentList({ comments }: { comments: CommentItem[] }) {
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();

  function vote(commentId: number, voteType: 1 | 2) {
    setMessage("");
    startTransition(async () => {
      const result = await voteComment(commentId, voteType);
      setMessage(result.ok ? "投票已提交，刷新后查看最新计数。" : result.message);
    });
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
            </div>
          </div>
          <p className="mt-3 text-sm leading-6 text-marine-text">{item.root.content}</p>
          {item.replies.length ? (
            <div className="mt-3 space-y-2 border-l-2 border-marine-blue/40 pl-3">
              {item.replies.slice(0, 3).map((reply) => (
                <p key={reply.id} className="text-sm leading-6 text-marine-muted">
                  <span className="font-medium text-marine-text">{reply.username || `用户 ${reply.userId}`}：</span>
                  {reply.content}
                </p>
              ))}
            </div>
          ) : null}
        </article>
      ))}
      {message ? <p className="text-sm text-marine-deep">{message}</p> : null}
    </div>
  );
}
