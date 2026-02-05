"use client";

import { useEffect, useMemo, useState } from "react";
import styles from "./SearchScreen.module.css";
import Section from "../../components/layout/Section";
import Card from "../../components/ui/Card";
import Input from "../../components/ui/Input";
import Button from "../../components/ui/Button";
import Badge from "../../components/ui/Badge";
import { apiGet, apiSend } from "../../lib/api";
import { useAuth } from "../../lib/useAuth";

type GameSearchResult = {
  app_id: number;
  name: string;
};

export default function SearchScreen() {
  const { user, refresh } = useAuth();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<GameSearchResult[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  const selectedGame = useMemo(
    () => results.find((g) => g.app_id === selectedId) ?? null,
    [results, selectedId]
  );

  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      setSelectedId(null);
      return;
    }
    const handle = window.setTimeout(async () => {
      const res = await apiGet<GameSearchResult[]>(
        `/search?q=${encodeURIComponent(query.trim())}&limit=8`
      );
      if (res.status === "success" && res.data) {
        setResults(res.data);
        setSelectedId(res.data[0]?.app_id ?? null);
      }
    }, 200);
    return () => window.clearTimeout(handle);
  }, [query]);

  function toast(text: string) {
    setMsg(text);
    window.setTimeout(() => setMsg(null), 2400);
  }

  async function likeSelectedGame() {
    if (!user) return toast("Login first to like games.");
    if (!selectedId) return toast("Select a game to like.");
    setBusy(true);
    try {
      const res = await apiSend(`/users/${user.id}/like/${selectedId}`);
      if (res.status !== "success") {
        return toast(res.error ?? "Failed to like game.");
      }
      await refresh();
      toast(`Liked: ${selectedGame?.name ?? selectedId}`);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className={styles.page}>
      <Section
        eyebrow="Search"
        title="Live Game Scout"
        subtitle="Type a title and watch the results lock in as you type."
      >
        <Card>
          <div className={styles.searchHeader}>
            <Input
              label="Game name"
              placeholder="Search by title..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
            <Button onClick={likeSelectedGame} disabled={!selectedId || busy}>
              Like Selected
            </Button>
          </div>
          {msg && <div className={styles.toast}>{msg}</div>}
          <div className={styles.results}>
            {results.length === 0 ? (
              <div className={styles.emptyState}>Start typing to see results.</div>
            ) : (
              results.map((game) => {
                const active = game.app_id === selectedId;
                return (
                  <button
                    key={game.app_id}
                    className={`${styles.resultRow}${active ? ` ${styles.active}` : ""}`}
                    type="button"
                    onClick={() => setSelectedId(game.app_id)}
                  >
                    <div>
                      <div className={styles.resultName}>{game.name}</div>
                      <div className={styles.resultMeta}>#{game.app_id}</div>
                    </div>
                    <Badge tone={active ? "hot" : "neutral"}>{active ? "Locked" : "Pick"}</Badge>
                  </button>
                );
              })
            )}
          </div>
        </Card>
      </Section>

      <Section
        eyebrow="Profile"
        title="Active Loadout"
        subtitle="Your latest liked games and session status."
      >
        <Card>
          {user ? (
            <div className={styles.profileRow}>
              <div>
                <div className={styles.profileName}>{user.name}</div>
                <div className={styles.profileMeta}>{user.email}</div>
              </div>
              <Badge tone="accent">Online</Badge>
            </div>
          ) : (
            <div className={styles.emptyState}>Login to start building your taste profile.</div>
          )}
          {user && (
            <div className={styles.pillRow}>
              {user.games_liked.length === 0 ? (
                <span className={styles.emptyState}>No liked games yet.</span>
              ) : (
                user.games_liked.map((id) => (
                  <span key={id} className={styles.pill}>
                    {id}
                  </span>
                ))
              )}
            </div>
          )}
        </Card>
      </Section>
    </div>
  );
}
