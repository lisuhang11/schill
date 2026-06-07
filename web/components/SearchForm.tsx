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
    const params = new URLSearchParams({
      keyword: values.keyword.trim(),
      type: values.type
    });
    router.push(`/search?${params.toString()}`);
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="w-full">
      <input type="hidden" value={type} {...register("type")} />
      <div className="flex flex-col gap-3 sm:flex-row">
        <div className="flex min-h-12 flex-1 items-center gap-3 rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-4 shadow-sm">
          <Search size={20} className="shrink-0 text-marine-deep" />
          <input
            {...register("keyword")}
            placeholder="搜索帖子、用户或话题"
            className="h-12 min-w-0 flex-1 bg-transparent text-sm text-marine-text outline-none placeholder:text-marine-muted/70"
          />
        </div>
        <button
          type="submit"
          disabled={isSubmitting}
          className="focus-ring inline-flex h-12 items-center justify-center gap-2 rounded-lg bg-marine-deep px-5 text-sm font-semibold text-white shadow-soft transition hover:bg-[#244f86] disabled:opacity-60"
        >
          <Search size={18} />
          搜索
        </button>
      </div>
      {errors.keyword ? <p className="mt-2 text-xs text-red-600">{errors.keyword.message}</p> : null}
    </form>
  );
}
