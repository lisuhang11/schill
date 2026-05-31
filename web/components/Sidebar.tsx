import Link from "next/link";
import { API_GAPS } from "@/lib/gaps";
import type { TopicInfo } from "@/lib/types";

export function Sidebar({ topics }: { topics: TopicInfo[] }) {
  return (
    <aside className="space-y-5">
      <section className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft">
        <h2 className="text-base font-semibold text-marine-text">热门话题</h2>
        <div className="mt-4 space-y-3">
          {topics.length ? (
            topics.slice(0, 8).map((topic) => (
              <Link
                href={`/search?keyword=${encodeURIComponent(topic.name)}&type=post`}
                key={topic.id}
                className="flex items-center justify-between rounded-lg px-2 py-2 text-sm hover:bg-marine-bg"
              >
                <span className="text-marine-text">#{topic.name}</span>
                <span className="text-xs text-marine-muted">{topic.quoteNum}</span>
              </Link>
            ))
          ) : (
            <p className="text-sm leading-6 text-marine-muted">暂无话题数据。</p>
          )}
        </div>
      </section>

      <section className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-5 shadow-soft">
        <h2 className="text-base font-semibold text-marine-text">接口缺口</h2>
        <div className="mt-3 space-y-3">
          {API_GAPS.map((gap) => (
            <div key={gap.capability} className="rounded-lg bg-marine-bg p-3">
              <p className="text-sm font-medium text-marine-text">{gap.capability}</p>
              <p className="mt-1 text-xs leading-5 text-marine-muted">{gap.impact}</p>
            </div>
          ))}
        </div>
      </section>
    </aside>
  );
}
