"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { Save } from "lucide-react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { createPost } from "@/lib/api";
import { POST_VISIBILITY_OPTIONS, type CreatePostRequest, type PostVisibility } from "@/lib/types";

const postSchema = z.object({
  title: z.string().trim().min(2, "标题至少 2 个字符").max(80, "标题不能超过 80 个字符"),
  cover: z.string().trim().url("封面必须是 URL").or(z.literal("")),
  visibility: z.coerce.number().pipe(z.union([
    z.literal(0),
    z.literal(10),
    z.literal(20),
    z.literal(50),
    z.literal(90)
  ])),
  content: z.string().trim().min(5, "正文至少 5 个字符").max(10000, "正文不能超过 10000 个字符"),
  topicsText: z.string().trim().max(120, "话题总长度不能超过 120 个字符"),
  tagsText: z.string().trim().max(120, "标签总长度不能超过 120 个字符")
});

type PostFormValues = z.infer<typeof postSchema>;

function splitInput(value: string): string[] {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

export function PostEditor() {
  const router = useRouter();
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();
  const {
    register,
    handleSubmit,
    formState: { errors }
  } = useForm<PostFormValues>({
    resolver: zodResolver(postSchema),
    defaultValues: {
      title: "",
      cover: "",
      visibility: 90,
      content: "",
      topicsText: "",
      tagsText: ""
    }
  });

  function onSubmit(values: PostFormValues) {
    setMessage("");
    const payload: CreatePostRequest = {
      title: values.title,
      cover: values.cover,
      visibility: values.visibility as PostVisibility,
      contents: [{ type: 2, content: values.content, sort: 100 }],
      topics: splitInput(values.topicsText),
      tags: splitInput(values.tagsText).join(",")
    };

    startTransition(async () => {
      const result = await createPost(payload);
      if (result.ok) {
        router.push(`/posts/${result.data.postId}`);
      } else {
        setMessage(result.message);
      }
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-5">
      <div>
        <label className="text-sm font-medium text-marine-text" htmlFor="title">标题</label>
        <input
          id="title"
          {...register("title")}
          className="focus-ring mt-2 h-11 w-full rounded-lg border border-[rgba(77,100,124,0.18)] px-3 text-sm"
        />
        {errors.title ? <p className="mt-1 text-xs text-red-600">{errors.title.message}</p> : null}
      </div>

      <div>
        <label className="text-sm font-medium text-marine-text" htmlFor="cover">封面 URL</label>
        <input
          id="cover"
          {...register("cover")}
          placeholder="可留空"
          className="focus-ring mt-2 h-11 w-full rounded-lg border border-[rgba(77,100,124,0.18)] px-3 text-sm"
        />
        {errors.cover ? <p className="mt-1 text-xs text-red-600">{errors.cover.message}</p> : null}
      </div>

      <div>
        <label className="text-sm font-medium text-marine-text" htmlFor="visibility">可见性</label>
        <select
          id="visibility"
          {...register("visibility")}
          className="focus-ring mt-2 h-11 w-full rounded-lg border border-[rgba(77,100,124,0.18)] px-3 text-sm"
        >
          {POST_VISIBILITY_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label} - {option.description}
            </option>
          ))}
        </select>
        {errors.visibility ? <p className="mt-1 text-xs text-red-600">{errors.visibility.message}</p> : null}
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <div>
          <label className="text-sm font-medium text-marine-text" htmlFor="tagsText">标签</label>
          <input
            id="tagsText"
            {...register("tagsText")}
            placeholder="后端,前端,社区"
            className="focus-ring mt-2 h-11 w-full rounded-lg border border-[rgba(77,100,124,0.18)] px-3 text-sm"
          />
          <p className="mt-1 text-xs text-marine-muted">逗号分隔，提交为后端 `tags` 字符串。</p>
        </div>
        <div>
          <label className="text-sm font-medium text-marine-text" htmlFor="topicsText">话题</label>
          <input
            id="topicsText"
            {...register("topicsText")}
            placeholder="go-zero,Next.js"
            className="focus-ring mt-2 h-11 w-full rounded-lg border border-[rgba(77,100,124,0.18)] px-3 text-sm"
          />
          <p className="mt-1 text-xs text-marine-muted">逗号分隔为字符串数组，不做自动补全。</p>
        </div>
      </div>

      <div>
        <label className="text-sm font-medium text-marine-text" htmlFor="content">正文</label>
        <textarea
          id="content"
          rows={12}
          {...register("content")}
          className="focus-ring mt-2 w-full resize-y rounded-lg border border-[rgba(77,100,124,0.18)] p-3 text-sm leading-6"
        />
        <p className="mt-1 text-xs text-marine-muted">只提交一个 `contents[]` 项，`type` 固定为 `2` 纯文本。</p>
        {errors.content ? <p className="mt-1 text-xs text-red-600">{errors.content.message}</p> : null}
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-[rgba(77,100,124,0.12)] pt-5">
        <p className="text-xs text-marine-muted">Markdown、图片、代码块和富媒体编辑暂不开放。</p>
        <button
          type="submit"
          disabled={isPending}
          className="focus-ring inline-flex items-center gap-2 rounded-lg bg-marine-deep px-5 py-2.5 text-sm font-semibold text-white disabled:opacity-60"
        >
          <Save size={18} /> 发布
        </button>
      </div>
      {message ? <p className="text-sm text-red-600">{message}</p> : null}
    </form>
  );
}
