"use client";

import { useState, useTransition } from "react";
import { Send, X } from "lucide-react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { createComment } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import type { CommentInfo } from "@/lib/types";

const commentSchema = z.object({
  content: z.string().trim().min(2, "评论至少 2 个字符").max(500, "评论不能超过 500 个字符")
});

type CommentValues = z.infer<typeof commentSchema>;

type CommentComposerProps = {
  postId: number;
  parentId?: number;
  replyToUserId?: number;
  replyToUsername?: string;
  placeholder?: string;
  onCancel?: () => void;
  onSuccess?: (comment: CommentInfo) => void;
};

export function CommentComposer({
  postId,
  parentId = 0,
  replyToUserId = 0,
  replyToUsername,
  placeholder,
  onCancel,
  onSuccess
}: CommentComposerProps) {
  const { userId } = useAuth();
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
    if (!userId) {
      setMessage("请先登录后再评论");
      return;
    }
    startTransition(async () => {
      const result = await createComment({
        postId,
        parentId,
        replyToUserId,
        content: values.content
      });
      if (result.ok) {
        setMessage(isReply ? "回复已发布" : "评论已发布");
        reset();
        onSuccess?.(result.data.comment);
      } else {
        setMessage(result.message);
      }
    });
  }

  const isReply = parentId !== 0;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className={isReply ? "mt-3" : "mt-5"}>
      {isReply && replyToUsername ? (
        <div className="mb-2 flex items-center justify-between text-xs text-marine-muted">
          <span>回复 <span className="font-medium text-marine-deep">{replyToUsername}</span></span>
          {onCancel ? (
            <button type="button" onClick={onCancel} className="inline-flex items-center gap-1 text-marine-muted hover:text-red-600">
              <X size={14} /> 取消
            </button>
          ) : null}
        </div>
      ) : null}
      <textarea
        {...register("content")}
        rows={isReply ? 2 : 3}
        placeholder={placeholder ?? "说点什么..."}
        className="focus-ring w-full resize-none rounded-lg border border-[rgba(77,100,124,0.14)] bg-white p-3 text-sm leading-6 shadow-sm"
      />
      {errors.content ? <p className="mt-1 text-xs text-red-600">{errors.content.message}</p> : null}
      <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
        <p className="text-xs text-marine-muted">{isReply ? "回复会展示在当前评论下方。" : "友好交流，别让键盘替你生气。"}</p>
        <button
          type="submit"
          disabled={isPending}
          className="focus-ring inline-flex items-center gap-2 rounded-lg bg-marine-deep px-4 py-2 text-sm font-semibold text-white shadow-sm disabled:opacity-60"
        >
          <Send size={18} /> {isReply ? "回复" : "发布评论"}
        </button>
      </div>
      {message ? <p className="mt-3 text-sm text-marine-deep">{message}</p> : null}
    </form>
  );
}
