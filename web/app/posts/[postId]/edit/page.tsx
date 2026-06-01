"use client";

import { useCallback, useEffect, useState, useTransition } from "react";
import { useParams, useRouter } from "next/navigation";
import { Edit3 } from "lucide-react";
import { getPostDetail, updatePost } from "@/lib/api";
import type { PostDetailResponse, PostVisibility } from "@/lib/types";
import { AuthGuard } from "@/components/AuthGuard";
import { PostEditor } from "@/components/PostEditor";
import { StateBlock } from "@/components/StateBlock";

export default function EditPostPage() {
  const params = useParams<{ postId: string }>();
  const router = useRouter();
  const postId = Number(params.postId);

  const [post, setPost] = useState<PostDetailResponse | null>(null);
  const [status, setStatus] = useState<"loading" | "error" | "ok">("loading");
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    if (Number.isNaN(postId)) {
      setStatus("error");
      setMessage("无效的文章 ID");
      return;
    }
    startTransition(async () => {
      const result = await getPostDetail(postId);
      if (!result.ok) {
        setStatus("error");
        setMessage(result.message);
        return;
      }
      setPost(result.data);
      setStatus("ok");
    });
  }, [postId]);

  const handleSubmit = useCallback(async (data: { title: string; cover: string; visibility: number; contents: { type: number; content: string; sort: number }[]; topics: string[]; tags: string }) => {
    startTransition(async () => {
      const result = await updatePost({
        postId,
        title: data.title,
        cover: data.cover,
        visibility: data.visibility as PostVisibility,
        contents: data.contents.map((c) => ({ ...c, type: 2 as const })),
        topics: data.topics,
        tags: data.tags
      });
      if (!result.ok) {
        setMessage(result.message);
        return;
      }
      router.push(`/posts/${postId}`);
    });
  }, [postId, router]);

  return (
    <AuthGuard>
      <main className="mx-auto w-full max-w-4xl px-4 py-8 md:px-8">
        <div className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-6 shadow-soft">
          <h1 className="flex items-center gap-2 text-3xl font-semibold text-marine-text">
            <Edit3 size={28} /> 编辑文章
          </h1>

          <div className="mt-6">
            {status === "loading" && <StateBlock tone="loading" title="加载中…" />}
            {status === "error" && <StateBlock tone="error" title={message || "加载失败"} />}
            {status === "ok" && post && (
              <PostEditor
                initialTitle={post.post.title}
                initialCover={post.post.cover}
                initialVisibility={post.post.visibility as number}
                initialContents={post.contents}
                initialTopics={post.topics?.map((t) => t.topicName) ?? []}
                initialTags={typeof post.post.tags === "string" ? post.post.tags : ""}
                onSubmit={handleSubmit}
                submitLabel="保存修改"
              />
            )}
            {message ? <p className="mt-4 text-sm text-red-600">{message}</p> : null}
          </div>
        </div>
      </main>
    </AuthGuard>
  );
}
