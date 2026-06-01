import { notFound } from "next/navigation";
import { getPostList, getUserInfo } from "@/lib/api";
import { ProfileContentTabs } from "@/components/ProfileContentTabs";
import { StateBlock } from "@/components/StateBlock";
import { formatDate } from "@/lib/format";

type UserPageProps = {
  params: Promise<{ userId: string }>;
  searchParams: Promise<{ page?: string }>;
};

const STATUS_LABELS: Record<number, string> = { 1: "正常", 2: "禁言", 3: "冻结" };
const PAGE_SIZE = 10;

export default async function UserPage({ params, searchParams }: UserPageProps) {
  const { userId: rawId } = await params;
  const userId = Number(rawId);
  if (!Number.isFinite(userId) || userId < 1) notFound();

  const sp = await searchParams;
  const page = Number(sp.page ?? "1") || 1;

  const [userResult, postsResult] = await Promise.all([
    getUserInfo(userId),
    getPostList({ userId, page, pageSize: PAGE_SIZE })
  ]);

  if (!userResult.ok) {
    return (
      <main className="mx-auto w-full max-w-4xl px-4 py-8 md:px-8">
        <StateBlock tone="error" title="用户加载失败" description={userResult.message} />
      </main>
    );
  }

  const { userInfo, profile, stat } = userResult.data;
  const hasProfile = Boolean(
    profile.location || profile.company || profile.jobTitle || profile.website
  );
  const profileItems = [
    profile.signature,
    profile.location,
    profile.company,
    profile.jobTitle,
    profile.website
  ];
  const profilePercent = Math.round(
    (profileItems.filter(Boolean).length / profileItems.length) * 100
  );

  const postList = postsResult.ok ? (postsResult.data.list ?? []) : null;
  const postTotal = postsResult.ok ? (postsResult.data.total ?? 0) : 0;

  return (
    <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-8 md:px-8 lg:grid-cols-[minmax(0,1fr)_300px]">
      <section>
        <section className="rounded-lg border border-[rgba(77,100,124,0.18)] bg-gradient-to-br from-white to-[#f5fdff] p-6 shadow-soft">
          <div className="flex flex-col gap-5 sm:flex-row sm:items-start">
            <div className="grid h-[88px] w-[88px] shrink-0 place-items-center rounded-3xl border-4 border-white bg-gradient-to-br from-[#bcefff] to-[#fff2bf] text-3xl font-black text-marine-deep shadow-soft">
              {userInfo.avatar ? "" : userInfo.username.slice(0, 2).toUpperCase()}
            </div>

            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-2.5">
                <h1 className="text-3xl font-semibold leading-tight text-marine-text">
                  {userInfo.username}
                </h1>
                <span className="inline-flex h-6 items-center rounded-full bg-marine-mint/20 px-2.5 text-xs font-extrabold text-marine-deep">
                  {STATUS_LABELS[userInfo.status] ?? `状态 ${userInfo.status}`}
                </span>
                <span className="inline-flex h-6 items-center rounded-full bg-marine-warm/45 px-2.5 text-xs font-extrabold text-[#7a5a00]">
                  Lv.12 创作者
                </span>
              </div>

              {profile.signature ? (
                <p className="mt-3 text-sm leading-7 text-marine-muted">{profile.signature}</p>
              ) : (
                <p className="mt-3 text-sm leading-7 text-marine-muted">
                  这个用户还没有填写个人签名。
                </p>
              )}

              <div className="mt-4 flex flex-wrap gap-x-4 gap-y-2 text-sm text-marine-muted">
                <span>注册于 {formatDate(userInfo.createdAt)}</span>
                {profile.location ? <span>{profile.location}</span> : null}
                {profile.company ? <span>{profile.company}</span> : null}
                {profile.jobTitle ? <span>{profile.jobTitle}</span> : null}
                {profile.website ? (
                  <span className="break-all text-marine-deep">{profile.website}</span>
                ) : null}
              </div>
            </div>
          </div>

          <div className="mt-5 grid grid-cols-2 gap-2.5 text-center sm:grid-cols-3 md:grid-cols-5">
            <StatBox label="文章" value={stat.postCount ?? 0} />
            <StatBox label="评论" value={stat.commentCount ?? 0} />
            <StatBox label="粉丝" value={stat.followerCount ?? 0} />
            <StatBox label="关注" value={stat.followingCount ?? 0} />
            <StatBox label="获赞" value={stat.likeCount ?? 0} />
          </div>
        </section>

        <ProfileContentTabs
          userId={userId}
          postCount={stat.postCount ?? postTotal}
          collectionCount={stat.collectionCount ?? 0}
          postPage={page}
          postTotal={postTotal}
          posts={postList}
          postsError={postsResult.ok ? undefined : postsResult.message}
        />
      </section>

      <aside className="space-y-4">
        <section className="rounded-lg border border-[rgba(77,100,124,0.18)] bg-white p-5 shadow-soft">
          <h2 className="text-base font-semibold text-marine-text">资料完整度</h2>
          <div className="mt-4 h-2.5 overflow-hidden rounded-full bg-[#e8f6fb]">
            <span
              className="block h-full rounded-full bg-gradient-to-r from-marine-blue to-marine-mint"
              style={{ width: `${profilePercent}%` }}
            />
          </div>
          <p className="mt-3 text-sm leading-6 text-marine-muted">
            当前资料完整度 {profilePercent}%。补充签名、地区、职位或网站后，个人主页展示会更完整。
          </p>
        </section>

        <section className="rounded-lg border border-[rgba(77,100,124,0.18)] bg-white p-5 shadow-soft">
          <h2 className="text-base font-semibold text-marine-text">最近活跃</h2>
          <p className="mt-3 text-sm leading-6 text-marine-muted">
            {stat.lastActiveTime
              ? `最后活跃于 ${formatDate(stat.lastActiveTime)}。`
              : "暂无最近活跃时间。"}
            {hasProfile ? " 个人资料已补充基础信息。" : " 该用户还没有补充更多资料。"}
          </p>
        </section>
      </aside>
    </main>
  );
}

function StatBox({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg bg-marine-bg p-3">
      <p className="text-2xl font-bold text-marine-deep">{value}</p>
      <p className="mt-1 text-xs text-marine-muted">{label}</p>
    </div>
  );
}
