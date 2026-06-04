"use client";

import { useAuth } from "@/lib/auth-context";
import { useRouter } from "next/navigation";
import { formatDate } from "@/lib/format";

const STATUS_LABELS: Record<number, string> = { 1: "正常", 2: "禁言", 3: "冻结" };

export type ProfileHeaderProps = {
  userId: number;
  userInfo: {
    userId: number;
    username: string;
    avatar: string;
    status: number;
    createdAt: number;
  };
  profile: {
    userId: number;
    gender: number;
    signature: string;
    location: string;
    website: string;
    company: string;
    jobTitle: string;
    education: string;
  };
};

export function ProfileHeader({ userId, userInfo, profile }: ProfileHeaderProps) {
  const { userId: currentUserId } = useAuth();
  const router = useRouter();
  const isOwner = currentUserId === userId;

  return (
    <section className="rounded-lg border border-[rgba(77,100,124,0.18)] bg-gradient-to-br from-white to-[#f5fdff] p-6 shadow-soft">
      <div className="flex flex-col gap-5 sm:flex-row sm:items-start">
        <div className="grid h-[88px] w-[88px] shrink-0 place-items-center rounded-3xl border-4 border-white bg-gradient-to-br from-[#bcefff] to-[#fff2bf] text-3xl font-black text-marine-deep shadow-soft">
          {userInfo.avatar ? "" : userInfo.username.slice(0, 2).toUpperCase()}
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center justify-between gap-2.5">
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

            {isOwner && (
              <button
                onClick={() => router.push(`/users/${userId}/edit`)}
                className="inline-flex h-9 items-center gap-1.5 rounded-lg border border-marine-blue/30 bg-white px-4 text-sm font-medium text-marine-blue shadow-sm transition hover:bg-marine-blue/5 hover:border-marine-blue/50 active:scale-[0.97]"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                  className="h-4 w-4"
                >
                  <path d="M5.433 13.917l1.262-3.155A4 4 0 017.58 9.42l6.92-6.918a2.121 2.121 0 013 3l-6.92 6.918c-.383.383-.84.685-1.343.886l-3.154 1.262a.5.5 0 01-.65-.65z" />
                  <path d="M3.5 5.75c0-.69.56-1.25 1.25-1.25H10A.75.75 0 0010 3H4.75A2.75 2.75 0 002 5.75v9.5A2.75 2.75 0 004.75 18h9.5A2.75 2.75 0 0017 15.25V10a.75.75 0 00-1.5 0v5.25c0 .69-.56 1.25-1.25 1.25h-9.5c-.69 0-1.25-.56-1.25-1.25v-9.5z" />
                </svg>
                编辑资料
              </button>
            )}
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
    </section>
  );
}
