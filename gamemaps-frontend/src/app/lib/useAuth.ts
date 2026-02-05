"use client";

import { useCallback, useEffect, useState } from "react";
import { apiGet, apiSend, ApiPayload } from "./api";

export type AuthUser = {
  id: number;
  name: string;
  email: string;
  games_liked: number[];
};

type AuthResponse = ApiPayload<AuthUser>;

export function useAuth() {
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
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
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
  }, []);

  return { user, loading, error, refresh, login, register, logout };
}
