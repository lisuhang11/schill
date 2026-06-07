"use client";

import Link from "next/link";
import {
  ChevronRight,
  LogOut,
  Moon,
  Palette,
  Settings,
  Sun,
  User,
  UserCircle,
  Monitor
} from "lucide-react";
import {
  useCallback,
  useEffect,
  useId,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
  type ReactNode
} from "react";

type UserProfileMenuProps = {
  userId: number;
  username: string | null;
  onLogout: () => void;
};

type MenuItemProps = {
  children: ReactNode;
  icon: ReactNode;
  danger?: boolean;
  onClick?: () => void;
  href?: string;
  hasSubmenu?: boolean;
  expanded?: boolean;
};

const appearanceOptions = [
  { label: "跟随系统", icon: Monitor },
  { label: "浅色", icon: Sun },
  { label: "深色", icon: Moon }
];

function initials(name: string) {
  return name.trim().slice(0, 2).toUpperCase();
}

function MenuItem({
  children,
  icon,
  danger = false,
  onClick,
  href,
  hasSubmenu = false,
  expanded = false
}: MenuItemProps) {
  const className = [
    "user-menu-item focus-ring group flex h-12 w-full items-center gap-3 px-4 text-left text-sm font-medium transition duration-100",
    danger
      ? "text-[#F23D4C] hover:bg-red-50 active:bg-red-100"
      : "text-marine-text hover:bg-[#F4F5F7] active:bg-[#EAF5FA]"
  ].join(" ");
  const content = (
    <>
      <span
        className={[
          "user-menu-item-icon grid h-[18px] w-[18px] place-items-center transition duration-100 group-hover:translate-x-0.5",
          danger ? "text-[#F23D4C]" : "text-marine-muted"
        ].join(" ")}
      >
        {icon}
      </span>
      <span className="min-w-0 flex-1 truncate">{children}</span>
      {hasSubmenu ? (
        <ChevronRight
          aria-hidden
          size={18}
          className={[
            "shrink-0 text-marine-muted transition-transform duration-200",
            expanded ? "rotate-90" : ""
          ].join(" ")}
        />
      ) : null}
    </>
  );

  if (href) {
    return (
      <Link href={href} role="menuitem" className={className} onClick={onClick}>
        {content}
      </Link>
    );
  }

  return (
    <button type="button" role="menuitem" className={className} onClick={onClick}>
      {content}
    </button>
  );
}

export function UserProfileMenu({ userId, username, onLogout }: UserProfileMenuProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [expandedSubmenu, setExpandedSubmenu] = useState<string | null>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const menuId = useId();
  const displayName = username ?? `用户 ${userId}`;

  const closeMenu = useCallback(() => {
    setIsOpen(false);
    setExpandedSubmenu(null);
    buttonRef.current?.focus();
  }, []);

  useEffect(() => {
    if (!isOpen) return;

    const handlePointerDown = (event: MouseEvent | TouchEvent) => {
      const target = event.target as Node;
      if (menuRef.current?.contains(target) || buttonRef.current?.contains(target)) {
        return;
      }
      closeMenu();
    };

    const handleKeyDown = (event: globalThis.KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        closeMenu();
      }
    };

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("touchstart", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("touchstart", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [closeMenu, isOpen]);

  const handleMenuKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    if (event.key !== "Tab" || !menuRef.current) return;

    const focusable = Array.from(
      menuRef.current.querySelectorAll<HTMLElement>(
        'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])'
      )
    ).filter((element) => !element.hasAttribute("disabled"));

    if (focusable.length === 0) return;

    const first = focusable[0];
    const last = focusable[focusable.length - 1];

    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  };

  const handleLogout = () => {
    if (!window.confirm("确定要退出登录吗？")) return;
    closeMenu();
    onLogout();
  };

  const toggleOpen = () => {
    setIsOpen((current) => !current);
  };

  return (
    <div className="relative">
      <button
        ref={buttonRef}
        type="button"
        aria-expanded={isOpen}
        aria-haspopup="menu"
        aria-controls={isOpen ? menuId : undefined}
        onClick={toggleOpen}
        className="focus-ring group relative grid h-8 w-8 place-items-center overflow-hidden rounded-full bg-white text-marine-deep shadow-[0_0_0_1px_rgba(77,100,124,0.16)] transition duration-150 hover:scale-[1.02] active:scale-[0.96]"
      >
        <span className="absolute inset-0 rounded-full bg-black opacity-0 transition group-hover:opacity-10" />
        <span className="grid h-full w-full place-items-center rounded-full bg-gradient-to-br from-marine-blue via-white to-marine-warm text-xs font-bold">
          {initials(displayName) || <UserCircle aria-hidden size={18} />}
        </span>
        <span className="sr-only">打开用户菜单</span>
      </button>

      {isOpen ? (
        <>
          <button
            type="button"
            aria-label="关闭用户菜单"
            className="fixed inset-0 z-30 bg-black/40 md:hidden"
            onClick={closeMenu}
          />
          <div
            ref={menuRef}
            id={menuId}
            role="menu"
            aria-label="用户菜单"
            onKeyDown={handleMenuKeyDown}
            className="user-profile-menu-panel fixed inset-x-4 bottom-4 z-40 max-h-[min(560px,calc(100vh-2rem))] overflow-hidden rounded-2xl border border-[rgba(77,100,124,0.14)] bg-white/95 shadow-float backdrop-blur-xl md:absolute md:inset-auto md:right-0 md:top-10 md:w-[280px] md:max-h-[min(400px,80vh)]"
          >
            <div className="flex items-center gap-3 px-4 py-4">
              <span className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-gradient-to-br from-marine-blue via-white to-marine-warm text-sm font-bold text-marine-deep shadow-[0_0_0_1px_rgba(77,100,124,0.12)]">
                {initials(displayName) || <UserCircle aria-hidden size={20} />}
              </span>
              <div className="min-w-0 flex-1">
                <p className="truncate text-base font-bold text-marine-text" title={displayName}>
                  {displayName}
                </p>
                <p className="truncate text-xs text-marine-muted">ID: {userId}</p>
              </div>
              <span
                title="会员权益"
                className="rounded-full bg-marine-warm/45 px-2 py-0.5 text-[11px] font-bold text-marine-deep"
              >
                会员
              </span>
            </div>

            <div className="mx-4 h-px bg-[rgba(0,0,0,0.06)]" />

            <div className="scroll-area max-h-[320px] overflow-y-auto py-2 pr-0.5 scroll-smooth md:max-h-[250px]">
              <MenuItem href={`/users/${userId}`} icon={<User aria-hidden size={18} />} onClick={closeMenu}>
                个人资料
              </MenuItem>
              <MenuItem href={`/users/${userId}/edit`} icon={<Settings aria-hidden size={18} />} onClick={closeMenu}>
                设置
              </MenuItem>
              <MenuItem
                icon={<Palette aria-hidden size={18} />}
                hasSubmenu
                expanded={expandedSubmenu === "appearance"}
                onClick={() => setExpandedSubmenu((current) => (current === "appearance" ? null : "appearance"))}
              >
                外观
              </MenuItem>
              <div
                className={[
                  "grid transition-[grid-template-rows] duration-200 ease-out",
                  expandedSubmenu === "appearance" ? "grid-rows-[1fr]" : "grid-rows-[0fr]"
                ].join(" ")}
              >
                <div className="overflow-hidden">
                  <div className="pb-1 pl-10 pr-2">
                    {appearanceOptions.map((option) => {
                      const Icon = option.icon;
                      return (
                        <button
                          key={option.label}
                          type="button"
                          role="menuitem"
                          className="focus-ring flex h-9 w-full items-center gap-2 rounded-lg px-3 text-left text-sm text-marine-muted transition hover:bg-[#F4F5F7] hover:text-marine-deep active:bg-[#EAF5FA]"
                        >
                          <Icon aria-hidden size={16} />
                          <span>{option.label}</span>
                        </button>
                      );
                    })}
                  </div>
                </div>
              </div>
            </div>

            <div className="mx-4 h-px bg-[rgba(0,0,0,0.06)]" />

            <div className="py-2">
              <MenuItem danger icon={<LogOut aria-hidden size={18} />} onClick={handleLogout}>
                退出登录
              </MenuItem>
            </div>
          </div>
        </>
      ) : null}
    </div>
  );
}
