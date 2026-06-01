"use client";

import Link from "next/link";
import { Edit3, Home, LogOut, Search, Tags, User, Zap } from "lucide-react";
import { useAuth } from "@/lib/auth-context";

const navItems = [
  { href: "/", label: "首页", icon: Home },
  { href: "/feed", label: "动态", icon: Zap },
  { href: "/search", label: "搜索", icon: Search },
  { href: "/topics", label: "话题", icon: Tags },
  { href: "/posts/new", label: "发布", icon: Edit3 }
];

export function AppHeader() {
  const { userId, username, logout } = useAuth();

  return (
    <header className="sticky top-0 z-20 border-b border-[rgba(77,100,124,0.16)] bg-white/88 backdrop-blur">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between px-4 md:px-8">
        <Link href="/" className="flex items-center gap-3">
          <span className="grid h-10 w-10 place-items-center rounded-lg bg-marine-deep text-sm font-bold text-white shadow-soft">
            SC
          </span>
          <span className="hidden text-base font-semibold text-marine-text sm:block">
            Schill Community
          </span>
        </Link>

        <div className="flex items-center gap-2">
          <nav className="flex items-center gap-1">
            {navItems.map((item) => {
              const Icon = item.icon;
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className="focus-ring inline-flex h-10 items-center gap-2 rounded-lg px-3 text-sm font-medium text-marine-muted transition hover:bg-marine-bg hover:text-marine-deep"
                >
                  <Icon aria-hidden size={18} />
                  <span className="hidden sm:inline">{item.label}</span>
                </Link>
              );
            })}
          </nav>

          <div className="hidden items-center gap-2 border-l border-[rgba(77,100,124,0.16)] pl-3 md:flex">
            {userId ? (
              <>
                <Link
                  href={`/users/${userId}`}
                  className="focus-ring inline-flex h-10 items-center gap-2 rounded-lg px-3 text-sm font-semibold text-marine-deep transition hover:bg-marine-bg"
                >
                  <User size={18} />
                  <span>{username ?? `用户 ${userId}`}</span>
                </Link>
                <button
                  type="button"
                  onClick={logout}
                  className="focus-ring inline-flex h-10 items-center gap-2 rounded-lg px-3 text-sm font-semibold text-marine-muted transition hover:bg-red-50 hover:text-red-600"
                >
                  <LogOut size={18} />
                  <span>退出</span>
                </button>
              </>
            ) : (
              <>
                <Link
                  href="/login"
                  className="focus-ring inline-flex h-10 items-center rounded-lg px-3 text-sm font-semibold text-marine-deep transition hover:bg-marine-bg"
                >
                  登录
                </Link>
                <Link
                  href="/register"
                  className="focus-ring inline-flex h-10 items-center rounded-lg bg-marine-deep px-4 text-sm font-semibold text-white shadow-soft transition hover:bg-[#244f86]"
                >
                  注册
                </Link>
              </>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
