"use client";

import { useCallback, useEffect, useState, useTransition } from "react";
import { getMyCollections } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import type { PostInfo } from "@/lib/types";
import { Pagination } from "@/components/Pagination";
import { PostCard } from "@/components/PostCard";
import { StateBlock } from "@/components/StateBlock";

type ProfileContentTabsProps = {
  userId: number;
  postCount: number;
  collectionCount: number;
  postPage: number;
  postTotal: number;
  posts: PostInfo[] | null;
  postsError?: string;
  onCollectionTotalChange?: (count: number) => void;
};

type TabKey = "posts" | "collections";
type CollectionStatus = "idle" | "loading" | "empty" | "error" | "ok";

const PAGE_SIZE = 10;

export function ProfileContentTabs({
  userId,
  postCount,
  collectionCount,
  postPage,
  postTotal,
  posts,
  postsError,
  onCollectionTotalChange
}: ProfileContentTabsProps) {
  const { userId: currentUserId } = useAuth();
  const isSelf = currentUserId === userId;
  const [activeTab, setActiveTab] = useState<TabKey>("posts");
  const [collectionPage, setCollectionPage] = useState(1);
  const [collectionTotal, setCollectionTotal] = useState(0);
  const [collections, setCollections] = useState<PostInfo[]>([]);
  const [collectionStatus, setCollectionStatus] = useState<CollectionStatus>("idle");
  const [collectionMessage, setCollectionMessage] = useState("");
  const [isPending, startTransition] = useTransition();

  const loadCollections = useCallback((page: number) => {
    startTransition(async () => {
      setCollectionStatus("loading");
      const result = await getMyCollections({ page, pageSize: PAGE_SIZE });
      if (!result.ok) {
        setCollectionStatus("error");
        setCollectionMessage(result.message);
        return;
      }

      const list = result.data.list ?? [];
      const total = result.data.total ?? 0;
      setCollections(list);
      setCollectionTotal(total);
      onCollectionTotalChange?.(total);
      setCollectionStatus(list.length === 0 ? "empty" : "ok");
    });
  }, [onCollectionTotalChange]);

  useEffect(() => {
    if (activeTab === "collections" && isSelf && collectionStatus === "idle") {
      loadCollections(collectionPage);
    }
  }, [activeTab, collectionPage, collectionStatus, isSelf, loadCollections]);

  function selectCollections() {
    setActiveTab("collections");
    if (!isSelf) {
      setCollectionStatus("error");
      setCollectionMessage("收藏仅本人可见，请登录对应账号后查看。");
    }
  }

  function changeCollectionPage(page: number) {
    setCollectionPage(page);
    loadCollections(page);
  }

  return (
    <section>
      <div
        className="mt-5 flex gap-2 overflow-x-auto border-b border-[rgba(77,100,124,0.18)]"
        aria-label="个人主页内容切换"
      >
        <button
          type="button"
          onClick={() => setActiveTab("posts")}
          className={`border-b-[3px] px-3.5 py-3 text-sm font-extrabold transition ${
            activeTab === "posts"
              ? "border-marine-blue text-marine-deep"
              : "border-transparent text-marine-muted hover:text-marine-deep"
          }`}
        >
          个人文章
          <span className="ml-1.5 text-[13px] opacity-70">{postCount}</span>
        </button>
        <button
          type="button"
          onClick={selectCollections}
          className={`border-b-[3px] px-3.5 py-3 text-sm font-extrabold transition ${
            activeTab === "collections"
              ? "border-marine-blue text-marine-deep"
              : "border-transparent text-marine-muted hover:text-marine-deep"
          }`}
        >
          收藏
          <span className="ml-1.5 text-[13px] opacity-70">{collectionCount}</span>
        </button>
      </div>

      {activeTab === "posts" ? (
        <div className="mt-4 space-y-4">
          {postsError ? (
            <StateBlock tone="error" title="文章加载失败" description={postsError} />
          ) : !posts || posts.length === 0 ? (
            <StateBlock tone="empty" title="暂无文章" description="该用户还没有发布任何文章。" />
          ) : (
            <>
              {posts.map((post) => (
                <PostCard key={post.id} post={post} />
              ))}
              <Pagination
                page={postPage}
                total={postTotal}
                pageSize={PAGE_SIZE}
                basePath={`/users/${userId}`}
              />
            </>
          )}
        </div>
      ) : (
        <div className="mt-4 space-y-4">
          {(collectionStatus === "loading" || isPending) && (
            <StateBlock tone="loading" title="收藏加载中" />
          )}
          {collectionStatus === "error" && (
            <StateBlock tone="error" title="收藏加载失败" description={collectionMessage} />
          )}
          {collectionStatus === "empty" && (
            <StateBlock
              tone="empty"
              title="暂无收藏"
              description="浏览内容时点击收藏后，会在这里集中展示。"
            />
          )}
          {collectionStatus === "ok" && (
            <>
              {collections.map((post) => (
                <PostCard key={post.id} post={post} />
              ))}
              <Pagination
                page={collectionPage}
                total={collectionTotal}
                pageSize={PAGE_SIZE}
                onPageChange={changeCollectionPage}
              />
            </>
          )}
        </div>
      )}
    </section>
  );
}
