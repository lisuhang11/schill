"use client";

import { useRouter } from "next/navigation";
import { Search } from "lucide-react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

const searchSchema = z.object({
  keyword: z.string().trim().min(1, "请输入搜索关键词").max(60, "关键词不能超过 60 个字符"),
  type: z.enum(["post", "user", "topic"])
});

type SearchFormValues = z.infer<typeof searchSchema>;

export function SearchForm({ keyword, type }: SearchFormValues) {
  const router = useRouter();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm<SearchFormValues>({
    resolver: zodResolver(searchSchema),
    defaultValues: { keyword, type }
  });

  function onSubmit(values: SearchFormValues) {
    const params = new URLSearchParams({ keyword: values.keyword, type: values.type });
    router.push(`/search?${params.toString()}`);
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="mt-5 grid gap-3 md:grid-cols-[1fr_160px_auto]">
      <div>
        <input
          {...register("keyword")}
          placeholder="搜索文章、用户或话题"
          className="focus-ring h-11 w-full rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-3 text-sm"
        />
        {errors.keyword ? <p className="mt-1 text-xs text-red-600">{errors.keyword.message}</p> : null}
      </div>
      <select
        {...register("type")}
        className="focus-ring h-11 rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-3 text-sm"
      >
        <option value="post">文章</option>
        <option value="user">用户</option>
        <option value="topic">话题</option>
      </select>
      <button
        type="submit"
        disabled={isSubmitting}
        className="focus-ring inline-flex h-11 items-center justify-center gap-2 rounded-lg bg-marine-deep px-4 text-sm font-semibold text-white disabled:opacity-60"
      >
        <Search size={18} /> 搜索
      </button>
    </form>
  );
}
