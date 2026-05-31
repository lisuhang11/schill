import type { PostVisibility } from "@/lib/types";

export function toBoolean(value: boolean | number | undefined): boolean {
  return value === true || value === 1;
}

export function splitTags(tags: string | string[] | undefined): string[] {
  if (Array.isArray(tags)) {
    return tags.filter(Boolean);
  }
  return (tags ?? "")
    .split(",")
    .map((tag) => tag.trim())
    .filter(Boolean);
}

export function joinTags(tags: string[]): string {
  return tags.map((tag) => tag.trim()).filter(Boolean).join(",");
}

export function formatCount(value: number | undefined): string {
  const safe = value ?? 0;
  if (safe >= 10000) {
    return `${(safe / 10000).toFixed(1)}w`;
  }
  return String(safe);
}

export function formatDate(timestamp: number | undefined): string {
  if (!timestamp) {
    return "未知时间";
  }
  const millis = timestamp > 9999999999 ? timestamp : timestamp * 1000;
  return new Intl.DateTimeFormat("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(millis));
}

export function visibilityLabel(value: PostVisibility | number | undefined): string {
  switch (value) {
    case 90:
      return "公开";
    case 50:
      return "互关";
    case 20:
      return "粉丝";
    case 10:
      return "充电";
    case 0:
      return "私密";
    default:
      return "未知";
  }
}
