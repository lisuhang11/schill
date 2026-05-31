"use client";

import { useState, useTransition } from "react";
import { Send } from "lucide-react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { createComment } from "@/lib/api";

const commentSchema = z.object({
  content: z.string().trim().min(2, "评论至少 2 个字符").max(500, "评论不能超过 500 个字符")
});

type CommentValues = z.infer<typeof commentSchema>;

export function CommentComposer({ postId }: { postId: number }) {
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();
  const {
    register,
    reset,
    handleSubmit,
    formState: { errors }
  } = useForm<CommentValues>({
    resolver: zodResolver(commentSchema),
    defaultValues: { content: "" }
  });

  function onSubmit(values: CommentValues) {
    setMessage("");
    startTransition(async () => {
      const result = await createComment({
        postId,
        parentId: 0,
        replyToUserId: 0,
        content: values.content
      });
      if (result.ok) {
        setMessage("评论已提交，刷新后可查看最新列表。");
        reset();
      } else {
        setMessage(result.message);
      }
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="mt-5">
      <textarea
        {...register("content")}
        rows={4}
        placeholder="写下你的评论"
        className="focus-ring w-full resize-y rounded-lg border border-[rgba(77,100,124,0.18)] bg-white p-3 text-sm leading-6"
      />
      {errors.content ? <p className="mt-1 text-xs text-red-600">{errors.content.message}</p> : null}
      <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
        <p className="text-xs text-marine-muted">评论接口来自 `/comment`，当前表单提交纯文本。</p>
        <button
          type="submit"
          disabled={isPending}
          className="focus-ring inline-flex items-center gap-2 rounded-lg bg-marine-deep px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
        >
          <Send size={18} /> 发布评论
        </button>
      </div>
      {message ? <p className="mt-3 text-sm text-marine-deep">{message}</p> : null}
    </form>
  );
}
