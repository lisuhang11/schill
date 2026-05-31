import { StateBlock } from "@/components/StateBlock";

export default function Loading() {
  return (
    <main className="mx-auto w-full max-w-6xl px-4 py-8 md:px-8">
      <StateBlock tone="loading" title="正在加载内容" description="请稍候，社区内容马上就绪。" />
    </main>
  );
}
