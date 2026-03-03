"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { apiGet, apiSend, type ApiPayload } from "./api";

export type AuthUser = {
  id: number;
  name: string;
  email: string;
  games_liked: number[];
};

type AuthContextValue = {
  user: AuthUser | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  login: (email: string, password: string) => Promise<ApiPayload<AuthUser>>;
  register: (name: string, email: string, password: string) => Promise<ApiPayload<AuthUser>>;
  logout: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiGet<AuthUser>("/auth/me");
      if (res.status === "success" && res.data) {
        setUser(res.data);
        setError(null);
      } else {
        setUser(null);
        setError(null);
      }
    } catch {
      setUser(null);
      setError("Could not load your session.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const login = useCallback(async (email: string, password: string) => {
    const res = await apiSend<AuthUser>("/auth/login", { email, password });
    if (res.status === "success" && res.data) {
      setUser(res.data);
      setError(null);
      return res;
    }
    setError(res.error ?? "Login failed");
    return res;
  }, []);

  const register = useCallback(async (name: string, email: string, password: string) => {
    const res = await apiSend<AuthUser>("/auth/register", { name, email, password });
    if (res.status === "success" && res.data) {
      setUser(res.data);
      setError(null);
      return res;
    }
    setError(res.error ?? "Registration failed");
    return res;
  }, []);

  const logout = useCallback(async () => {
    await apiSend("/auth/logout");
    setUser(null);
    setError(null);
  }, []);

  const value = useMemo(
    () => ({ user, loading, error, refresh, login, register, logout }),
    [user, loading, error, refresh, login, register, logout]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return value;
}
