"use client";

import { useCallback, useEffect, useState, useTransition } from "react";
import { Flame, Layers, Zap } from "lucide-react";
import { getFeedList } from "@/lib/api";
import type { FeedItem } from "@/lib/types";
import { PostCard } from "@/components/PostCard";
import { Pagination } from "@/components/Pagination";
import { StateBlock } from "@/components/StateBlock";

type FeedTab = "following" | "hot" | "latest";
type HotRange = "week" | "month" | "year";

const TAB_OPTIONS: { key: FeedTab; label: string; icon: typeof Layers }[] = [
  { key: "following", label: "关注", icon: Layers },
  { key: "hot", label: "高热度", icon: Flame },
  { key: "latest", label: "最新", icon: Zap }
];

const HOT_RANGE_OPTIONS: { key: HotRange; label: string }[] = [
  { key: "week", label: "本周" },
  { key: "month", label: "本月" },
  { key: "year", label: "本年" }
];

export default function FeedPage() {
  const [tab, setTab] = useState<FeedTab>("following");
  const [hotRange, setHotRange] = useState<HotRange>("week");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [items, setItems] = useState<FeedItem[]>([]);
  const [status, setStatus] = useState<"loading" | "empty" | "error" | "ok">("loading");
  const [message, setMessage] = useState("");
  const [, startTransition] = useTransition();

  const fetchFeed = useCallback((feedType: FeedTab, pageNum: number) => {
    startTransition(async () => {
      setStatus("loading");
      const result = await getFeedList({ feedType, page: pageNum, pageSize: 10 });
      if (!result.ok) {
        setStatus("error");
        setMessage(result.message);
        setItems([]);
        setTotal(0);
        return;
      }
      setItems(result.data.list);
      setTotal(result.data.total);
      setStatus(result.data.list.length === 0 ? "empty" : "ok");
    });
  }, []);

  useEffect(() => {
    fetchFeed(tab, page);
  }, [tab, hotRange, page, fetchFeed]);

  function switchTab(newTab: FeedTab) {
    if (newTab === tab) return;
    setTab(newTab);
    setPage(1);
  }

  function switchHotRange(newRange: HotRange) {
    if (newRange === hotRange) return;
    setHotRange(newRange);
    setPage(1);
  }

  return (
    <main className="mx-auto w-full max-w-4xl px-4 py-8 md:px-8">
      <div className="flex flex-col gap-3 border-b border-[rgba(77,100,124,0.14)] pb-3 sm:flex-row sm:items-center sm:justify-between">
        <nav className="flex gap-2 overflow-x-auto" aria-label="动态类型">
          {TAB_OPTIONS.map(({ key, label, icon: Icon }) => (
            <button
              key={key}
              type="button"
              onClick={() => switchTab(key)}
              className={`focus-ring inline-flex min-w-20 items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-semibold transition ${
                tab === key
                  ? "bg-marine-deep text-white shadow-sm"
                  : "bg-white text-marine-muted hover:bg-marine-bg hover:text-marine-deep"
              }`}
            >
              <Icon size={17} />
              {label}
            </button>
          ))}
        </nav>

        {tab === "hot" && (
          <div className="inline-flex w-fit rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-1 shadow-[0_8px_24px_rgba(45,68,92,0.08)]">
            {HOT_RANGE_OPTIONS.map(({ key, label }) => (
              <button
                key={key}
                type="button"
                onClick={() => switchHotRange(key)}
                className={`focus-ring min-w-14 rounded-md px-3 py-1.5 text-sm font-semibold transition ${
                  hotRange === key
                    ? "bg-marine-deep text-white"
                    : "text-marine-muted hover:bg-marine-bg hover:text-marine-deep"
                }`}
              >
                {label}
              </button>
            ))}
          </div>
        )}
      </div>

      <section className="mt-6">
        {status === "loading" && <StateBlock tone="loading" title="加载中" />}
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
              <Pagination page={page} total={total} pageSize={10} onPageChange={setPage} />
            </div>
          </>
        )}
      </section>
    </main>
  );
}
