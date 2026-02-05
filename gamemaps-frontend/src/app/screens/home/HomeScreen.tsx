"use client";

import { useEffect, useMemo, useState } from "react";
import styles from "./HomeScreen.module.css";
import Container from "../../components/layout/Container";
import Section from "../../components/layout/Section";
import Card from "../../components/ui/Card";
import Button from "../../components/ui/Button";
import Input from "../../components/ui/Input";
import Badge from "../../components/ui/Badge";
import Stat from "../../components/ui/Stat";

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

type MapNode = {
  id: string;
  label: string;
  kind: "user" | "liked" | "recommended";
  app_id?: number;
  x: number;
  y: number;
  score?: number;
  anchor?: string;
};

type MapEdge = {
  from: string;
  to: string;
  kind: "liked" | "recommended";
};

type MapPayload = {
  user_id: number;
  nodes: MapNode[];
  edges: MapEdge[];
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

export default function HomeScreen() {
  const [userName, setUserName] = useState("");
  const [userIdInput, setUserIdInput] = useState("");
  const [activeUser, setActiveUser] = useState<UserResponse | null>(null);

  const [query, setQuery] = useState("");
  const [searchResults, setSearchResults] = useState<GameSearchResult[]>([]);
  const [selectedAppId, setSelectedAppId] = useState<number | null>(null);

  const [taste, setTaste] = useState<TasteResponse | null>(null);
  const [recs, setRecs] = useState<RecommendedGame[]>([]);
  const [mapPayload, setMapPayload] = useState<MapPayload | null>(null);
  const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null);

  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  const selectedGame = useMemo(
    () => searchResults.find((g) => g.app_id === selectedAppId) ?? null,
    [searchResults, selectedAppId]
  );

  function toast(text: string) {
    setMsg(text);
    window.setTimeout(() => setMsg(null), 2500);
  }

  const canLike = !!activeUser && !!selectedAppId;

  async function createUser() {
    const name = userName.trim();
    if (!name) return toast("Enter a name to create a user.");

    setBusy(true);
    try {
      const created = await apiPost<CreateUserResponse>("/users", { name });
      if (created.status !== "success" || !created.data) {
        return toast(created.error ?? "Failed to create user.");
      }

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

  async function getMap() {
    if (!activeUser) return toast("Load or create a user first.");

    setBusy(true);
    try {
      const r = await apiGet<MapPayload>(`/users/${activeUser.id}/map`);
      if (r.status !== "success" || !r.data) {
        setMapPayload(null);
        return toast(r.error ?? "Failed to get map.");
      }
      setMapPayload(r.data);
      toast("Taste map loaded.");
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    if (searchResults.length === 0) setSelectedAppId(null);
  }, [searchResults]);

  const mapLayout = useMemo(() => {
    if (!mapPayload || mapPayload.nodes.length === 0) {
      return { viewBox: "-400 -260 800 520", nodeById: new Map<string, MapNode>() };
    }
    let minX = mapPayload.nodes[0].x;
    let maxX = mapPayload.nodes[0].x;
    let minY = mapPayload.nodes[0].y;
    let maxY = mapPayload.nodes[0].y;
    const nodeById = new Map<string, MapNode>();

    for (const node of mapPayload.nodes) {
      nodeById.set(node.id, node);
      minX = Math.min(minX, node.x);
      maxX = Math.max(maxX, node.x);
      minY = Math.min(minY, node.y);
      maxY = Math.max(maxY, node.y);
    }

    const pad = 80;
    const viewBox = `${minX - pad} ${minY - pad} ${maxX - minX + pad * 2} ${maxY - minY + pad * 2}`;
    return { viewBox, nodeById };
  }, [mapPayload]);

  const highlighted = useMemo(() => {
    if (!hoveredNodeId || !mapPayload) {
      return { nodes: new Set<string>(), edges: new Set<string>() };
    }
    const nodes = new Set<string>([hoveredNodeId]);
    const edges = new Set<string>();
    for (const edge of mapPayload.edges) {
      if (edge.from === hoveredNodeId || edge.to === hoveredNodeId) {
        edges.add(`${edge.from}-${edge.to}`);
        nodes.add(edge.from);
        nodes.add(edge.to);
      }
    }
    return { nodes, edges };
  }, [hoveredNodeId, mapPayload]);

  return (
    <div className={styles.page}>
      <Container>
        <header className={styles.hero}>
          <div>
            <div className={styles.heroBadge}>
              <Badge tone="hot">Live Beta</Badge>
              <span>Play smarter with taste-aware discovery.</span>
            </div>
            <h1 className={styles.title}>GameMaps</h1>
            <p className={styles.lead}>
              Build your player profile, map your taste, and pull recommendations shaped by the games you
              actually love.
            </p>
            <div className={styles.heroActions}>
              <Button onClick={getRecommendations} disabled={!activeUser || busy}>
                Get Recommendations
              </Button>
              <Button variant="outline" onClick={getMap} disabled={!activeUser || busy}>
                Load Taste Map
              </Button>
              <Button variant="ghost" onClick={recomputeTaste} disabled={!activeUser || busy}>
                Recompute Taste
              </Button>
            </div>
            <div className={styles.heroFootnote}>
              Premium clustering • skill-tree layout • live taste recalibration
            </div>
          </div>
          <div className={styles.heroPanel}>
            <div className={styles.statGrid}>
              <Stat label="Active User" value={activeUser ? `#${activeUser.id}` : "None"} />
              <Stat label="Liked Games" value={activeUser ? activeUser.games_liked.length : 0} />
              <Stat label="Search Hits" value={searchResults.length} />
              <Stat label="Taste Keys" value={taste ? Object.keys(taste.embedding).length : "-"} />
            </div>
            <div className={styles.heroFooter}>
              {activeUser ? (
                <>
                  <span className={styles.heroName}>{activeUser.name}</span>
                  <span className={styles.heroMeta}>Your taste map is ready to evolve.</span>
                </>
              ) : (
                <span className={styles.heroMeta}>Create or load a player to unlock your taste map.</span>
              )}
            </div>
          </div>
        </header>

        {msg && <div className={styles.toast}>{msg}</div>}

        <div className={styles.mainGrid}>
          <Section
            eyebrow="Profile"
            title="Pilot Setup"
            subtitle="Create a player or load an existing profile to sync your picks."
            className={styles.leftColumn}
          >
            <Card title="Create Player" subtitle="Start with a fresh handle.">
              <div className={styles.inlineRow}>
                <Input
                  placeholder="Player handle"
                  value={userName}
                  onChange={(e) => setUserName(e.target.value)}
                />
                <Button onClick={createUser} disabled={busy}>
                  Create
                </Button>
              </div>
            </Card>

            <Card title="Load Player" subtitle="Pick up where you left off.">
              <div className={styles.inlineRow}>
                <Input
                  placeholder="Player ID"
                  value={userIdInput}
                  onChange={(e) => setUserIdInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") loadUserById();
                  }}
                />
                <Button variant="outline" onClick={loadUserById} disabled={busy}>
                  Load
                </Button>
              </div>
            </Card>

            <Card title="Loadout" subtitle="Your current player status.">
              <div className={styles.loadoutRow}>
                <div>
                  <span className={styles.loadoutLabel}>Active</span>
                  <div className={styles.loadoutValue}>
                    {activeUser ? `#${activeUser.id} - ${activeUser.name}` : "None"}
                  </div>
                </div>
                <Badge tone={activeUser ? "accent" : "neutral"}>
                  {activeUser ? "Online" : "Offline"}
                </Badge>
              </div>
              {activeUser && activeUser.games_liked.length > 0 && (
                <div className={styles.pillRow}>
                  {activeUser.games_liked.map((id) => (
                    <span key={id} className={styles.pill}>
                      {id}
                    </span>
                  ))}
                </div>
              )}
            </Card>
          </Section>

          <Section
            eyebrow="Discovery"
            title="Search & Lock-In"
            subtitle="Pick a game, like it, and evolve your taste profile."
          >
            <Card title="Search Games" subtitle="Type a title fragment to scout results.">
              <div className={styles.inlineRow}>
                <Input
                  placeholder="Search games"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") searchGames();
                  }}
                />
                <Button onClick={searchGames} disabled={busy}>
                  Search
                </Button>
              </div>
              {searchResults.length > 0 && (
                <div className={styles.results}>
                  {searchResults.map((g) => {
                    const active = g.app_id === selectedAppId;
                    return (
                      <button
                        key={g.app_id}
                        className={`${styles.resultItem}${active ? ` ${styles.active}` : ""}`}
                        onClick={() => setSelectedAppId(g.app_id)}
                        type="button"
                      >
                        <div>
                          <div className={styles.resultTitle}>{g.name}</div>
                          <div className={styles.resultMeta}>#{g.app_id}</div>
                        </div>
                        <Badge tone={active ? "hot" : "neutral"}>{active ? "Selected" : "Pick"}</Badge>
                      </button>
                    );
                  })}
                </div>
              )}
            </Card>

            <Card title="Taste Actions" subtitle="Lock in a favorite and update the map.">
              <div className={styles.actionRow}>
                <Button
                  onClick={likeSelectedGame}
                  disabled={!canLike || busy}
                  title={!activeUser ? "Load/create a user first" : !selectedAppId ? "Select a game" : ""}
                >
                  Like Selected
                </Button>
                <Button variant="ghost" onClick={recomputeTaste} disabled={!activeUser || busy}>
                  Recompute Taste
                </Button>
                <Button variant="outline" onClick={getRecommendations} disabled={!activeUser || busy}>
                  Get Recs
                </Button>
              </div>
              <div className={styles.selectedLine}>
                Selected: {selectedGame ? `${selectedGame.name} (#${selectedGame.app_id})` : "None"}
              </div>
            </Card>
          </Section>
        </div>

        <div className={styles.bottomGrid}>
          <Section
            eyebrow="Taste"
            title="Embedding Snapshot"
            subtitle="Peek into the first 25 keys that shape your recommendations."
          >
            <Card>
              {!taste ? (
                <div className={styles.emptyState}>Recompute taste after liking a few games.</div>
              ) : (
                <>
                  <div className={styles.tasteMeta}>
                    User #{taste.user_id} | Games used: {taste.games_used} | Keys: {Object.keys(taste.embedding).length}
                  </div>
                  <pre className={styles.pre}>
                    {JSON.stringify(Object.entries(taste.embedding).slice(0, 25), null, 2)}
                  </pre>
                  <div className={styles.helper}>Showing first 25 key-value pairs.</div>
                </>
              )}
            </Card>
          </Section>

          <Section
            eyebrow="Map"
            title="Taste Profile Map"
            subtitle="Clusters show which liked games anchor each recommendation."
          >
            <Card>
              <div className={styles.mapHeader}>
                <div className={styles.mapLabel}>Taste Map</div>
                <div className={styles.mapControls}>
                  <Button size="sm" variant="ghost" onClick={getMap} disabled={!activeUser || busy}>
                    Refresh Map
                  </Button>
                </div>
              </div>
              <div className={styles.mapLegend}>
                <span className={styles.legendItem}>
                  <span className={`${styles.legendDot} ${styles.legendUser}`} />
                  You
                </span>
                <span className={styles.legendItem}>
                  <span className={`${styles.legendDot} ${styles.legendLiked}`} />
                  Liked
                </span>
                <span className={styles.legendItem}>
                  <span className={`${styles.legendDot} ${styles.legendRec}`} />
                  Recommended
                </span>
              </div>
              {!mapPayload ? (
                <div className={styles.emptyState}>Load a taste map to see your clusters.</div>
              ) : (
                <div className={styles.mapWrap}>
                  <svg className={styles.mapSvg} viewBox={mapLayout.viewBox} role="img">
                    <g className={styles.mapGuides}>
                      <circle cx="0" cy="0" r="160" />
                      <circle cx="0" cy="0" r="280" />
                      <circle cx="0" cy="0" r="360" />
                    </g>
                    <g className={styles.mapEdges}>
                      {mapPayload.edges.map((edge, idx) => {
                        const from = mapLayout.nodeById.get(edge.from);
                        const to = mapLayout.nodeById.get(edge.to);
                        if (!from || !to) return null;
                        const key = `${edge.from}-${edge.to}`;
                        const isHot = highlighted.edges.size > 0 && highlighted.edges.has(key);
                        return (
                          <line
                            key={`${edge.from}-${edge.to}-${idx}`}
                            x1={from.x}
                            y1={from.y}
                            x2={to.x}
                            y2={to.y}
                            className={
                              edge.kind === "liked"
                                ? `${styles.edgeLiked}${isHot ? ` ${styles.edgeHot}` : ""}`
                                : `${styles.edgeRec}${isHot ? ` ${styles.edgeHot}` : ""}`
                            }
                          />
                        );
                      })}
                    </g>
                    <g className={styles.mapNodes}>
                      {mapPayload.nodes.map((node) => {
                        const r = node.kind === "user" ? 18 : node.kind === "liked" ? 12 : 9;
                        const isHot = highlighted.nodes.size > 0 && highlighted.nodes.has(node.id);
                        return (
                          <g
                            key={node.id}
                            className={`${styles.node}${isHot ? ` ${styles.nodeHot}` : ""}`}
                            onMouseEnter={() => setHoveredNodeId(node.id)}
                            onMouseLeave={() => setHoveredNodeId(null)}
                          >
                            <circle
                              cx={node.x}
                              cy={node.y}
                              r={r}
                              className={
                                node.kind === "user"
                                  ? styles.nodeUser
                                  : node.kind === "liked"
                                  ? styles.nodeLiked
                                  : styles.nodeRec
                              }
                            />
                            <text x={node.x} y={node.y - r - 6} className={styles.nodeLabel}>
                              {node.label}
                            </text>
                          </g>
                        );
                      })}
                    </g>
                  </svg>
                </div>
              )}
            </Card>
          </Section>

          <Section
            eyebrow="Drops"
            title="Recommendations"
            subtitle="Your latest pulls based on your taste graph."
          >
            <Card>
              {recs.length === 0 ? (
                <div className={styles.emptyState}>
                  Recompute taste, then click "Get Recommendations" to populate your drop list.
                </div>
              ) : (
                <div className={styles.recs}>
                  {recs.map((r) => (
                    <div key={r.app_id} className={styles.recRow}>
                      <div>
                        <div className={styles.recName}>{r.name}</div>
                        <div className={styles.recMeta}>#{r.app_id}</div>
                      </div>
                      <span className={styles.recScore}>{r.score.toFixed(4)}</span>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </Section>
        </div>
      </Container>
    </div>
  );
}
