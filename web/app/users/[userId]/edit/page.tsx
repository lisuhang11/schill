"use client";

import { useAuth } from "@/lib/auth-context";
import { getUserInfo, updateProfile } from "@/lib/api";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

type ProfileData = {
  signature: string;
  location: string;
  website: string;
  company: string;
  jobTitle: string;
  education: string;
  gender: number;
};

const EMPTY_PROFILE: ProfileData = {
  signature: "",
  location: "",
  website: "",
  company: "",
  jobTitle: "",
  education: "",
  gender: 0
};

const FIELD_CONFIG: { key: keyof ProfileData; label: string; placeholder: string; multiline?: boolean }[] = [
  { key: "signature", label: "个性签名", placeholder: "用一句话介绍自己...", multiline: true },
  { key: "location", label: "所在地", placeholder: "如：广东广州" },
  { key: "company", label: "公司", placeholder: "如：腾讯" },
  { key: "jobTitle", label: "职位", placeholder: "如：前端工程师" },
  { key: "website", label: "个人网站", placeholder: "如：https://example.com" },
  { key: "education", label: "教育经历", placeholder: "如：清华大学" }
];

export default function EditProfilePage() {
  const { userId, username } = useAuth();
  const router = useRouter();
  const [form, setForm] = useState<ProfileData>(EMPTY_PROFILE);
  const [original, setOriginal] = useState<ProfileData>(EMPTY_PROFILE);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ tone: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    if (!userId) {
      router.push("/login");
      return;
    }
    (async () => {
      const result = await getUserInfo(userId);
      if (result.ok) {
        const p = result.data.profile;
        const data: ProfileData = {
          signature: p.signature ?? "",
          location: p.location ?? "",
          website: p.website ?? "",
          company: p.company ?? "",
          jobTitle: p.jobTitle ?? "",
          education: p.education ?? "",
          gender: p.gender ?? 0
        };
        setForm(data);
        setOriginal(data);
      }
      setLoading(false);
    })();
  }, [userId, router]);

  const changed = JSON.stringify(form) !== JSON.stringify(original);

  const handleChange = (key: keyof ProfileData, value: string) => {
    setForm(prev => ({ ...prev, [key]: value }));
  };

  const handleSave = async () => {
    setSaving(true);
    setMessage(null);

    const payload: Record<string, string | number | undefined> = {};
    for (const key of Object.keys(form) as (keyof ProfileData)[]) {
      if (form[key] !== original[key]) {
        payload[key] = form[key];
      }
    }

    const result = await updateProfile(payload);
    if (result.ok) {
      setOriginal(form);
      setMessage({ tone: "success", text: "个人资料已更新" });
    } else {
      setMessage({ tone: "error", text: result.message || "更新失败" });
    }
    setSaving(false);
  };

  if (loading) {
    return (
      <main className="mx-auto w-full max-w-2xl px-4 py-8 md:px-8">
        <div className="animate-pulse space-y-6">
          <div className="h-8 w-48 rounded bg-marine-bg" />
          {[...Array(6)].map((_, i) => (
            <div key={i} className="h-12 rounded bg-marine-bg" />
          ))}
        </div>
      </main>
    );
  }

  if (!userId) return null;

  return (
    <main className="mx-auto w-full max-w-2xl px-4 py-8 md:px-8">
      <div className="mb-8 flex items-center gap-4">
        <button
          onClick={() => router.push(`/users/${userId}`)}
          className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-[rgba(77,100,124,0.18)] bg-white text-marine-muted shadow-soft transition hover:text-marine-blue hover:border-marine-blue/30"
          aria-label="返回个人主页"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="h-5 w-5">
            <path fillRule="evenodd" d="M17 10a.75.75 0 01-.75.75H5.612l4.158 3.96a.75.75 0 11-1.04 1.08l-5.5-5.25a.75.75 0 010-1.08l5.5-5.25a.75.75 0 111.04 1.08L5.612 9.25H16.25A.75.75 0 0117 10z" clipRule="evenodd" />
          </svg>
        </button>
        <div>
          <h1 className="text-2xl font-semibold text-marine-text">编辑资料</h1>
          <p className="mt-0.5 text-sm text-marine-muted">{username}</p>
        </div>
      </div>

      {message && (
        <div
          className={`mb-6 rounded-lg border px-4 py-3 text-sm ${
            message.tone === "success"
              ? "border-marine-mint/40 bg-marine-mint/10 text-marine-deep"
              : "border-red-200 bg-red-50 text-red-700"
          }`}
        >
          {message.text}
        </div>
      )}

      <div className="space-y-6">
        {FIELD_CONFIG.map(({ key, label, placeholder, multiline }) => (
          <div key={key}>
            <label className="mb-2 block text-sm font-medium text-marine-text">{label}</label>
            {multiline ? (
              <textarea
                value={form[key] as string}
                onChange={e => handleChange(key, e.target.value)}
                placeholder={placeholder}
                rows={3}
                className="w-full rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-4 py-2.5 text-sm text-marine-text shadow-soft placeholder:text-marine-muted/50 focus:border-marine-blue/40 focus:outline-none focus:ring-2 focus:ring-marine-blue/10 transition resize-none"
              />
            ) : (
              <input
                type="text"
                value={form[key] as string}
                onChange={e => handleChange(key, e.target.value)}
                placeholder={placeholder}
                className="w-full rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-4 py-2.5 text-sm text-marine-text shadow-soft placeholder:text-marine-muted/50 focus:border-marine-blue/40 focus:outline-none focus:ring-2 focus:ring-marine-blue/10 transition"
              />
            )}
          </div>
        ))}
      </div>

      <div className="mt-8 flex items-center gap-3">
        <button
          onClick={handleSave}
          disabled={!changed || saving}
          className="inline-flex h-10 items-center gap-2 rounded-lg bg-marine-blue px-6 text-sm font-semibold text-white shadow-soft transition hover:bg-marine-deep active:scale-[0.97] disabled:opacity-40 disabled:cursor-not-allowed disabled:active:scale-100"
        >
          {saving ? (
            <>
              <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              保存中...
            </>
          ) : (
            "保存修改"
          )}
        </button>
        <button
          onClick={() => router.push(`/users/${userId}`)}
          className="inline-flex h-10 items-center rounded-lg border border-[rgba(77,100,124,0.18)] bg-white px-5 text-sm font-medium text-marine-muted shadow-soft transition hover:text-marine-text hover:border-marine-muted/40 active:scale-[0.97]"
        >
          取消
        </button>
      </div>
    </main>
  );
}
