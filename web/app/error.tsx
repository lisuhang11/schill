"use client";

import { StateBlock } from "@/components/StateBlock";

export default function Error({
  reset
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <main className="mx-auto w-full max-w-6xl px-4 py-8 md:px-8">
      <StateBlock
        tone="error"
        title="页面加载失败"
        description="当前页面遇到错误，可以重试加载。"
        actionLabel="重试"
        onAction={reset}
      />
    </main>
  );
}
