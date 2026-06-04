"use client";

import { useCallback, useEffect, useState, useTransition } from "react";
import { Bookmark } from "lucide-react";
import { getMyCollections } from "@/lib/api";
import type { PostInfo } from "@/lib/types";
import { AuthGuard } from "@/components/AuthGuard";
import { PostCard } from "@/components/PostCard";
import { Pagination } from "@/components/Pagination";
import { StateBlock } from "@/components/StateBlock";

export default function CollectionsPage() {
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [items, setItems] = useState<PostInfo[]>([]);
  const [status, setStatus] = useState<"loading" | "empty" | "error" | "ok">("loading");
  const [message, setMessage] = useState("");
  const [, startTransition] = useTransition();

  const fetch = useCallback((pageNum: number) => {
    startTransition(async () => {
      setStatus("loading");
      const result = await getMyCollections({ page: pageNum, pageSize: 10 });
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
    fetch(page);
  }, [page, fetch]);

  return (
    <AuthGuard>
      <main className="mx-auto w-full max-w-4xl px-4 py-8 md:px-8">
        <h1 className="flex items-center gap-2 text-3xl font-semibold text-marine-text">
          <Bookmark size={28} /> 我的收藏
        </h1>

        <div className="mt-6">
          {status === "loading" && <StateBlock tone="loading" title="加载中…" />}
          {status === "error" && <StateBlock tone="error" title={message || "加载失败"} />}
          {status === "empty" && <StateBlock tone="empty" title="暂无收藏" description="浏览内容时点击收藏按钮即可添加到此处。" />}
          {status === "ok" && (
            <>
              <div className="space-y-4">
                {items.map((post) => (
                  <PostCard key={post.id} post={post} />
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
    </AuthGuard>
  );
}
