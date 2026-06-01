"use client";

import { useRouter } from "next/navigation";
import { type ReactNode, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";

export function AuthGuard({ children }: { children: ReactNode }) {
  const { userId } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (userId === null) {
      router.replace("/login");
    }
  }, [userId, router]);

  if (userId === null) {
    return null;
  }

  return <>{children}</>;
}
