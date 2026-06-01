"use client";

import { useCallback, useEffect, useState, useTransition } from "react";
import { Layers, Zap } from "lucide-react";
import { getFeedList } from "@/lib/api";
import type { FeedItem } from "@/lib/types";
import { PostCard } from "@/components/PostCard";
import { Pagination } from "@/components/Pagination";
import { StateBlock } from "@/components/StateBlock";

type FeedTab = "latest" | "following";

const TAB_OPTIONS: { key: FeedTab; label: string; icon: typeof Layers }[] = [
  { key: "latest", label: "最新", icon: Zap },
  { key: "following", label: "关注", icon: Layers }
];

export default function FeedPage() {
  const [tab, setTab] = useState<FeedTab>("latest");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [items, setItems] = useState<FeedItem[]>([]);
  const [status, setStatus] = useState<"loading" | "empty" | "error" | "ok">("loading");
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();

  const fetchFeed = useCallback((feedType: FeedTab, pageNum: number) => {
    startTransition(async () => {
      setStatus("loading");
      const result = await getFeedList({ feedType, page: pageNum, pageSize: 10 });
      if (!result.ok) {
        setStatus("error");
        setMessage(result.message);
        return;
      }
      setItems(result.data.list);
      setTotal(result.data.total);
      setStatus(result.data.list.length === 0 ? "empty" : "ok");
    });
  }, []);

  useEffect(() => {
    fetchFeed(tab, page);
  }, [tab, page, fetchFeed]);

  function switchTab(newTab: FeedTab) {
    if (newTab === tab) return;
    setTab(newTab);
    setPage(1);
  }

  return (
    <main className="mx-auto w-full max-w-4xl px-4 py-8 md:px-8">
      <h1 className="text-3xl font-semibold text-marine-text">动态</h1>

      <div className="mt-6 flex gap-2">
        {TAB_OPTIONS.map(({ key, label, icon: Icon }) => (
          <button
            key={key}
            onClick={() => switchTab(key)}
            className={`focus-ring inline-flex items-center gap-1.5 rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
              tab === key
                ? "bg-marine-deep text-white"
                : "bg-gray-100 text-marine-muted hover:bg-gray-200"
            }`}
          >
            <Icon size={16} /> {label}
          </button>
        ))}
      </div>

      <div className="mt-6">
        {status === "loading" && <StateBlock tone="loading" title="加载中…" />}
        {status === "error" && <StateBlock tone="error" title={message || "加载失败"} />}
        {status === "empty" && <StateBlock tone="empty" title="暂无内容" description="还没有动态，去发布第一篇吧。" />}
        {status === "ok" && (
          <>
            <div className="space-y-4">
              {items.map((item) => (
                <PostCard key={item.postId} post={item} />
              ))}
            </div>
            <div className="mt-6">
              <Pagination
                page={page}
                total={total}
                pageSize={10}
                onPageChange={setPage}
              />
            </div>
          </>
        )}
      </div>
    </main>
  );
}
