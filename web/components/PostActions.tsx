"use client";

import { useEffect, useState } from "react";
import { Bookmark, Heart, Share2 } from "lucide-react";
import { checkPostCollection, checkPostStar, collectPost, sharePost, starPost, uncollectPost, unstarPost } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { formatCount } from "@/lib/format";
import type { InteractionToggleResponse } from "@/lib/types";

type PostActionsProps = {
  postId: number;
  counts: {
    upvoteCount: number;
    collectionCount: number;
    shareCount: number;
  };
  initialState?: {
    isStarred?: boolean;
    isCollected?: boolean;
  };
};

export function PostActions({ postId, counts, initialState }: PostActionsProps) {
  const { userId } = useAuth();
  const [starred, setStarred] = useState(Boolean(initialState?.isStarred));
  const [collected, setCollected] = useState(Boolean(initialState?.isCollected));
  const [upvotes, setUpvotes] = useState(counts.upvoteCount);
  const [collections, setCollections] = useState(counts.collectionCount);
  const [shares, setShares] = useState(counts.shareCount);
  const [message, setMessage] = useState("");
  const [pendingAction, setPendingAction] = useState<"star" | "collect" | "share" | null>(null);
  const isPending = pendingAction !== null;

  useEffect(() => {
    if (!userId) {
      setStarred(false);
      setCollected(false);
      return;
    }

    let alive = true;
    async function loadState() {
      const [starResult, collectionResult] = await Promise.all([
        checkPostStar(postId),
        checkPostCollection(postId)
      ]);
      if (!alive) return;
      if (starResult.ok) {
        setStarred(starResult.data.isStarred);
        setUpvotes(starResult.data.starCount);
      }
      if (collectionResult.ok) {
        setCollected(collectionResult.data.isCollected);
      }
    }
    loadState();
    return () => {
      alive = false;
    };
  }, [postId, userId]);

  async function run(
    actionName: "star" | "collect" | "share",
    action: () => Promise<{ ok: boolean; data?: InteractionToggleResponse; message?: string }>
  ) {
    setMessage("");
    if (!userId) {
      setMessage("请先登录后再操作");
      return;
    }
    setPendingAction(actionName);
    const result = await action();
    if (!result.ok) {
      setMessage(result.message ?? "操作失败");
    }
    setPendingAction(null);
  }

  return (
    <div className="mt-8 border-t border-[rgba(77,100,124,0.12)] pt-5">
      <div className="flex flex-wrap gap-2">
        <button
          type="button"
          disabled={isPending}
          onClick={() =>
            run("star", async () => {
              const wasStarred = starred;
              const result = wasStarred ? await unstarPost(postId) : await starPost(postId);
              if (result.ok) {
                const nextStarred = result.data.isStarred ?? !wasStarred;
                setStarred(nextStarred);
                setUpvotes((current) => result.data.starCount ?? Math.max(0, current + (nextStarred ? 1 : -1)));
                setMessage(nextStarred ? "已点赞" : "已取消点赞");
              }
              return result;
            })
          }
          className={`focus-ring inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-semibold transition disabled:opacity-60 ${
            starred ? "bg-marine-pink text-white" : "bg-marine-bg text-marine-deep"
          }`}
        >
          <Heart size={18} fill={starred ? "currentColor" : "none"} /> {starred ? "已点赞" : "点赞"} {formatCount(upvotes)}
        </button>
        <button
          type="button"
          disabled={isPending}
          onClick={() =>
            run("collect", async () => {
              const wasCollected = collected;
              const result = wasCollected ? await uncollectPost(postId) : await collectPost(postId);
              if (result.ok) {
                const nextCollected = result.data.isCollected ?? !wasCollected;
                setCollected(nextCollected);
                setCollections((current) => Math.max(0, current + (nextCollected ? 1 : -1)));
                setMessage(nextCollected ? "已收藏" : "已取消收藏");
              }
              return result;
            })
          }
          className={`focus-ring inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-semibold transition disabled:opacity-60 ${
            collected ? "bg-marine-warm text-marine-text" : "bg-marine-bg text-marine-deep"
          }`}
        >
          <Bookmark size={18} fill={collected ? "currentColor" : "none"} /> {collected ? "已收藏" : "收藏"} {formatCount(collections)}
        </button>
        <button
          type="button"
          disabled={isPending}
          onClick={() =>
            run("share", async () => {
              const result = await sharePost(postId);
              if (result.ok) {
                setShares((current) => result.data.shareCount ?? current + 1);
                setMessage("已记录分享");
              }
              return result;
            })
          }
          className="focus-ring inline-flex items-center gap-2 rounded-lg bg-marine-deep px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
        >
          <Share2 size={18} /> 分享 {formatCount(shares)}
        </button>
      </div>
      {message ? <p className="mt-3 text-sm text-marine-deep">{message}</p> : null}
    </div>
  );
}
