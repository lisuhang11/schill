"use client";

import { useEffect, useState } from "react";
import { Bookmark, Heart, MessageCircle, PenLine, UserCheck, Users } from "lucide-react";
import { getMyCollections } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { formatCount } from "@/lib/format";

type ProfileStatsProps = {
  userId: number;
  postCount: number;
  commentCount: number;
  followerCount: number;
  followingCount: number;
  likeCount: number;
  collectionCount: number;
  onCollectionCountChange?: (count: number) => void;
};

export function ProfileStats({
  userId,
  postCount,
  commentCount,
  followerCount,
  followingCount,
  likeCount,
  collectionCount,
  onCollectionCountChange
}: ProfileStatsProps) {
  const { userId: currentUserId } = useAuth();
  const [liveCollectionCount, setLiveCollectionCount] = useState(collectionCount);

  useEffect(() => {
    setLiveCollectionCount(collectionCount);
  }, [collectionCount]);

  useEffect(() => {
    if (currentUserId !== userId) return;

    let alive = true;
    async function loadCollectionCount() {
      const result = await getMyCollections({ page: 1, pageSize: 1 });
      if (!alive || !result.ok) return;
      const nextCount = result.data.total ?? 0;
      setLiveCollectionCount(nextCount);
      onCollectionCountChange?.(nextCount);
    }
    loadCollectionCount();
    return () => {
      alive = false;
    };
  }, [currentUserId, onCollectionCountChange, userId]);

  return (
    <div className="mt-5 grid grid-cols-2 gap-2.5 text-center sm:grid-cols-3 md:grid-cols-6">
      <StatBox icon={PenLine} label="文章" value={postCount} />
      <StatBox icon={MessageCircle} label="评论" value={commentCount} />
      <StatBox icon={Users} label="粉丝" value={followerCount} />
      <StatBox icon={UserCheck} label="关注" value={followingCount} />
      <StatBox icon={Heart} label="获赞" value={likeCount} />
      <StatBox icon={Bookmark} label="收藏" value={liveCollectionCount} />
    </div>
  );
}

function StatBox({
  icon: Icon,
  label,
  value
}: {
  icon: typeof PenLine;
  label: string;
  value: number;
}) {
  return (
    <div className="rounded-lg bg-marine-bg p-3">
      <Icon className="mx-auto text-marine-deep" size={18} />
      <p className="mt-2 text-2xl font-bold text-marine-deep">{formatCount(value)}</p>
      <p className="mt-1 text-xs text-marine-muted">{label}</p>
    </div>
  );
}
