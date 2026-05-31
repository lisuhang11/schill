"use client";

import { useState, useTransition } from "react";
import { Bookmark, Heart, Share2 } from "lucide-react";
import { collectPost, sharePost, starPost, uncollectPost, unstarPost } from "@/lib/api";
import { formatCount } from "@/lib/format";

type PostActionsProps = {
  postId: number;
  counts: {
    upvoteCount: number;
    collectionCount: number;
    shareCount: number;
  };
};

export function PostActions({ postId, counts }: PostActionsProps) {
  const [starred, setStarred] = useState(false);
  const [collected, setCollected] = useState(false);
  const [upvotes, setUpvotes] = useState(counts.upvoteCount);
  const [collections, setCollections] = useState(counts.collectionCount);
  const [shares, setShares] = useState(counts.shareCount);
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();

  function run(action: () => Promise<{ ok: boolean; data?: { starCount?: number; shareCount?: number }; message?: string }>) {
    setMessage("");
    startTransition(async () => {
      const result = await action();
      if (!result.ok) {
        setMessage(result.message ?? "操作失败");
      }
    });
  }

  return (
    <div className="mt-8 border-t border-[rgba(77,100,124,0.12)] pt-5">
      <div className="flex flex-wrap gap-2">
        <button
          type="button"
          disabled={isPending}
          onClick={() =>
            run(async () => {
              const result = starred ? await unstarPost(postId) : await starPost(postId);
              if (result.ok) {
                setStarred(!starred);
                setUpvotes(result.data.starCount ?? upvotes + (starred ? -1 : 1));
              }
              return result;
            })
          }
          className="focus-ring inline-flex items-center gap-2 rounded-lg bg-marine-bg px-4 py-2 text-sm font-semibold text-marine-deep disabled:opacity-60"
        >
          <Heart size={18} /> {starred ? "取消点赞" : "点赞"} {formatCount(upvotes)}
        </button>
        <button
          type="button"
          disabled={isPending}
          onClick={() =>
            run(async () => {
              const result = collected ? await uncollectPost(postId) : await collectPost(postId);
              if (result.ok) {
                setCollected(!collected);
                setCollections(collections + (collected ? -1 : 1));
              }
              return result;
            })
          }
          className="focus-ring inline-flex items-center gap-2 rounded-lg bg-marine-bg px-4 py-2 text-sm font-semibold text-marine-deep disabled:opacity-60"
        >
          <Bookmark size={18} /> {collected ? "取消收藏" : "收藏"} {formatCount(collections)}
        </button>
        <button
          type="button"
          disabled={isPending}
          onClick={() =>
            run(async () => {
              const result = await sharePost(postId);
              if (result.ok) {
                setShares(result.data.shareCount ?? shares + 1);
              }
              return result;
            })
          }
          className="focus-ring inline-flex items-center gap-2 rounded-lg bg-marine-deep px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
        >
          <Share2 size={18} /> 分享 {formatCount(shares)}
        </button>
      </div>
      {message ? <p className="mt-3 text-sm text-red-600">{message}</p> : null}
    </div>
  );
}
