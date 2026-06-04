"use client";

import { useState } from "react";
import type { PostInfo } from "@/lib/types";
import { ProfileContentTabs } from "@/components/ProfileContentTabs";
import { ProfileStats } from "@/components/ProfileStats";

type ProfileMainContentProps = {
  userId: number;
  stats: {
    postCount: number;
    commentCount: number;
    followerCount: number;
    followingCount: number;
    likeCount: number;
    collectionCount: number;
  };
  postPage: number;
  postTotal: number;
  posts: PostInfo[];
  postsError?: string;
};

export function ProfileMainContent({
  userId,
  stats,
  postPage,
  postTotal,
  posts,
  postsError
}: ProfileMainContentProps) {
  const [collectionCount, setCollectionCount] = useState(stats.collectionCount);

  return (
    <>
      <ProfileStats
        userId={userId}
        postCount={stats.postCount}
        commentCount={stats.commentCount}
        followerCount={stats.followerCount}
        followingCount={stats.followingCount}
        likeCount={stats.likeCount}
        collectionCount={collectionCount}
        onCollectionCountChange={setCollectionCount}
      />

      <ProfileContentTabs
        userId={userId}
        postCount={stats.postCount}
        collectionCount={collectionCount}
        postPage={postPage}
        postTotal={postTotal}
        posts={posts}
        postsError={postsError}
        onCollectionTotalChange={setCollectionCount}
      />
    </>
  );
}
