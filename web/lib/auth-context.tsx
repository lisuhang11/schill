"use client";

import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react";

type AuthState = {
  userId: number | null;
  username: string | null;
};

type AuthContextValue = AuthState & {
  login: (userId: number, username: string) => void;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

function readAuthState(): AuthState {
  if (typeof window === "undefined") return { userId: null, username: null };
  const rawUserId = window.localStorage.getItem("schill:userId");
  const username = window.localStorage.getItem("schill:username");
  const userId = rawUserId ? Number(rawUserId) : null;
  return userId && !Number.isNaN(userId) ? { userId, username } : { userId: null, username: null };
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [auth, setAuth] = useState<AuthState>({ userId: null, username: null });

  useEffect(() => {
    setAuth(readAuthState());
  }, []);

  const login = useCallback((userId: number, username: string) => {
    window.localStorage.setItem("schill:username", username);
    setAuth({ userId, username });
  }, []);

  const logout = useCallback(() => {
    window.localStorage.removeItem("schill:userId");
    window.localStorage.removeItem("schill:accessToken");
    window.localStorage.removeItem("schill:refreshToken");
    window.localStorage.removeItem("schill:username");
    setAuth({ userId: null, username: null });
  }, []);

  return (
    <AuthContext.Provider value={{ ...auth, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within <AuthProvider>");
  return ctx;
}
