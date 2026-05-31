import { PostEditor } from "@/components/PostEditor";

export default function NewPostPage() {
  return (
    <main className="mx-auto w-full max-w-4xl px-4 py-8 md:px-8">
      <div className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-6 shadow-soft">
        <h1 className="text-3xl font-semibold text-marine-text">发布文章</h1>
        <p className="mt-2 text-sm leading-6 text-marine-muted">
          当前编辑器只支持纯文本内容，标签以逗号分隔，话题为简单文本列表。
        </p>
        <PostEditor />
      </div>
    </main>
  );
}
