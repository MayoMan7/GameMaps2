"use client";

import { useEffect, useMemo, useState } from "react";
import styles from "./page.module.css";

type ApiPayload<T> = {
  status: "success" | "error";
  data?: T;
  error?: string;
};

type GameSearchResult = {
  app_id: number;
  name: string;
  header_image?: string;
};

type RecommendedGame = {
  app_id: number;
  name: string;
  score: number;
};

type CreateUserResponse = {
  id: number;
  name: string;
};

type UserResponse = {
  id: number;
  name: string;
  games_liked: number[];
};

type TasteResponse = {
  user_id: number;
  games_used: number;
  embedding: Record<string, number>;
};

async function apiGet<T>(path: string): Promise<ApiPayload<T>> {
  const res = await fetch(`/api${path}`, {
    method: "GET",
    headers: { "Content-Type": "application/json" },
    cache: "no-store",
  });
  const json = (await res.json()) as ApiPayload<T>;
  return json;
}

async function apiPost<T>(path: string, body?: any): Promise<ApiPayload<T>> {
  const res = await fetch(`/api${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
    cache: "no-store",
  });
  const json = (await res.json()) as ApiPayload<T>;
  return json;
}

export default function Home() {
  // user
  const [userName, setUserName] = useState("");
  const [userIdInput, setUserIdInput] = useState("");
  const [activeUser, setActiveUser] = useState<UserResponse | null>(null);

  // search
  const [query, setQuery] = useState("");
  const [searchResults, setSearchResults] = useState<GameSearchResult[]>([]);
  const [selectedAppId, setSelectedAppId] = useState<number | null>(null);

  // taste + recs
  const [taste, setTaste] = useState<TasteResponse | null>(null);
  const [recs, setRecs] = useState<RecommendedGame[]>([]);

  // ui state
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  const selectedGame = useMemo(
    () => searchResults.find((g) => g.app_id === selectedAppId) ?? null,
    [searchResults, selectedAppId]
  );

  function toast(text: string) {
    setMsg(text);
    // auto-clear after a bit
    window.setTimeout(() => setMsg(null), 2500);
  }

  const canLike = !!activeUser && !!selectedAppId;

  // --- actions ---
  async function createUser() {
    const name = userName.trim();
    if (!name) return toast("Enter a name to create a user.");

    setBusy(true);
    try {
      const created = await apiPost<CreateUserResponse>("/users", { name });
      if (created.status !== "success" || !created.data) {
        return toast(created.error ?? "Failed to create user.");
      }

      // fetch user data after creating
      const u = await apiGet<UserResponse>(`/users/${created.data.id}`);
      if (u.status === "success" && u.data) {
        setActiveUser(u.data);
        setUserIdInput(String(u.data.id));
        toast(`Created user #${u.data.id}`);
      } else {
        toast("Created user, but failed to load user profile.");
      }
    } finally {
      setBusy(false);
    }
  }

  async function loadUserById() {
    const id = Number(userIdInput);
    if (!Number.isFinite(id) || id <= 0) return toast("Enter a valid user ID.");

    setBusy(true);
    try {
      const u = await apiGet<UserResponse>(`/users/${id}`);
      if (u.status !== "success" || !u.data) {
        setActiveUser(null);
        return toast(u.error ?? "User not found.");
      }
      setActiveUser(u.data);
      toast(`Loaded user #${u.data.id}`);
    } finally {
      setBusy(false);
    }
  }

  async function searchGames() {
    const q = query.trim();
    if (!q) return toast("Type a game name to search.");

    setBusy(true);
    try {
      const r = await apiGet<GameSearchResult[]>(`/search?q=${encodeURIComponent(q)}`);
      if (r.status !== "success" || !r.data) {
        setSearchResults([]);
        setSelectedAppId(null);
        return toast(r.error ?? "Search failed.");
      }
      setSearchResults(r.data);
      setSelectedAppId(r.data[0]?.app_id ?? null);
      toast(`Found ${r.data.length} result(s).`);
    } finally {
      setBusy(false);
    }
  }

  async function likeSelectedGame() {
    if (!activeUser) return toast("Load or create a user first.");
    if (!selectedAppId) return toast("Select a game first.");

    setBusy(true);
    try {
      const r = await apiPost<any>(`/users/${activeUser.id}/like/${selectedAppId}`);
      if (r.status !== "success") {
        return toast(r.error ?? "Failed to like game.");
      }

      // refresh user so liked list updates
      const u = await apiGet<UserResponse>(`/users/${activeUser.id}`);
      if (u.status === "success" && u.data) {
        setActiveUser(u.data);
      }

      toast(`Liked: ${selectedGame?.name ?? selectedAppId}`);
    } finally {
      setBusy(false);
    }
  }

  async function recomputeTaste() {
    if (!activeUser) return toast("Load or create a user first.");

    setBusy(true);
    try {
      const r = await apiPost<TasteResponse>(`/users/${activeUser.id}/recompute-taste`);
      if (r.status !== "success" || !r.data) {
        return toast(r.error ?? "Failed to recompute taste.");
      }
      setTaste(r.data);
      toast(`Taste updated using ${r.data.games_used} game(s).`);
    } finally {
      setBusy(false);
    }
  }

  async function getRecommendations() {
    if (!activeUser) return toast("Load or create a user first.");

    setBusy(true);
    try {
      const r = await apiGet<RecommendedGame[]>(`/users/${activeUser.id}/recommendations`);
      if (r.status !== "success" || !r.data) {
        setRecs([]);
        return toast(r.error ?? "Failed to get recommendations.");
      }
      setRecs(r.data);
      toast("Recommendations updated.");
    } finally {
      setBusy(false);
    }
  }

  // optional: keep selectedAppId valid if results change
  useEffect(() => {
    if (searchResults.length === 0) setSelectedAppId(null);
  }, [searchResults]);

  return (
    <div className={styles.page}>
      <div className={styles.main}>
        <h1>GameMaps</h1>

        {msg && <div className={styles.toast}>{msg}</div>}

        {/* --- User --- */}
        <section className={styles.card}>
          <h2>User</h2>

          <div className={styles.row}>
            <input
              className={styles.input}
              placeholder="Create user (name)"
              value={userName}
              onChange={(e) => setUserName(e.target.value)}
            />
            <button className={styles.button} onClick={createUser} disabled={busy}>
              Create
            </button>
          </div>

          <div className={styles.row}>
            <input
              className={styles.input}
              placeholder="Load user by ID"
              value={userIdInput}
              onChange={(e) => setUserIdInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") loadUserById();
              }}
            />
            <button className={styles.button} onClick={loadUserById} disabled={busy}>
              Load
            </button>
          </div>

          <div className={styles.subtle}>
            Active user:{" "}
            {activeUser ? (
              <>
                <b>
                  #{activeUser.id}</b> — {activeUser.name} — liked{" "}
                <b>{activeUser.games_liked.length}</b>
              </>
            ) : (
              <i>none</i>
            )}
          </div>

          {activeUser && activeUser.games_liked.length > 0 && (
            <div className={styles.smallList}>
              <div className={styles.label}>Liked game IDs:</div>
              <div className={styles.pills}>
                {activeUser.games_liked.map((id) => (
                  <span key={id} className={styles.pill}>
                    {id}
                  </span>
                ))}
              </div>
            </div>
          )}
        </section>

        {/* --- Search + Like --- */}
        <section className={styles.card}>
          <h2>Search & Like</h2>

          <div className={styles.row}>
            <input
              className={styles.input}
              placeholder="Search games (e.g., Grand)"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") searchGames();
              }}
            />
            <button className={styles.button} onClick={searchGames} disabled={busy}>
              Search
            </button>
          </div>

          {searchResults.length > 0 && (
            <div className={styles.results}>
              {searchResults.map((g) => {
                const active = g.app_id === selectedAppId;
                return (
                  <button
                    key={g.app_id}
                    className={`${styles.resultItem} ${active ? styles.active : ""}`}
                    onClick={() => setSelectedAppId(g.app_id)}
                    type="button"
                  >
                    <div className={styles.resultTitle}>{g.name}</div>
                    <div className={styles.resultMeta}>#{g.app_id}</div>
                  </button>
                );
              })}
            </div>
          )}

          <div className={styles.row}>
            <button
              className={styles.button}
              onClick={likeSelectedGame}
              disabled={!canLike || busy}
              title={!activeUser ? "Load/create a user first" : !selectedAppId ? "Select a game" : ""}
            >
              Like selected game
            </button>

            <button className={styles.button} onClick={recomputeTaste} disabled={!activeUser || busy}>
              Recompute taste
            </button>

            <button className={styles.button} onClick={getRecommendations} disabled={!activeUser || busy}>
              Get recommendations
            </button>
          </div>

          <div className={styles.subtle}>
            Selected:{" "}
            {selectedGame ? (
              <>
                <b>{selectedGame.name}</b> (#{selectedGame.app_id})
              </>
            ) : (
              <i>none</i>
            )}
          </div>
        </section>

        {/* --- Taste snapshot --- */}
        <section className={styles.card}>
          <h2>Taste Profile</h2>
          {!taste ? (
            <div className={styles.subtle}>
              Recompute taste after liking a few games to see your embedding snapshot.
            </div>
          ) : (
            <>
              <div className={styles.subtle}>
                User #{taste.user_id} • games used: <b>{taste.games_used}</b> • keys:{" "}
                <b>{Object.keys(taste.embedding).length}</b>
              </div>
              <pre className={styles.pre}>
                {JSON.stringify(Object.entries(taste.embedding).slice(0, 25), null, 2)}
              </pre>
              <div className={styles.subtle}>
                Showing first 25 (key, value) pairs.
              </div>
            </>
          )}
        </section>

        {/* --- Recommendations --- */}
        <section className={styles.card}>
          <h2>Recommendations</h2>
          {recs.length === 0 ? (
            <div className={styles.subtle}>
              Click “Get recommendations” after recomputing taste.
            </div>
          ) : (
            <div className={styles.recs}>
              {recs.map((r) => (
                <div key={r.app_id} className={styles.recRow}>
                  <div className={styles.recName}>{r.name}</div>
                  <div className={styles.recMeta}>
                    #{r.app_id} • score {r.score.toFixed(4)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
