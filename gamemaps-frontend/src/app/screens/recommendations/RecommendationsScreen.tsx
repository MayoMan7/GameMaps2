"use client";

import { useState } from "react";
import styles from "./RecommendationsScreen.module.css";
import Section from "../../components/layout/Section";
import Card from "../../components/ui/Card";
import Button from "../../components/ui/Button";
import { apiGet, apiSend } from "../../lib/api";
import { useAuth } from "../../lib/useAuth";

type RecommendedGame = {
  app_id: number;
  name: string;
  score: number;
};

export default function RecommendationsScreen() {
  const { user } = useAuth();
  const [recs, setRecs] = useState<RecommendedGame[]>([]);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  function toast(text: string) {
    setMsg(text);
    window.setTimeout(() => setMsg(null), 2400);
  }

  async function loadRecommendations() {
    if (!user) return toast("Login to get recommendations.");
    setBusy(true);
    try {
      const res = await apiGet<RecommendedGame[]>(`/users/${user.id}/recommendations`);
      if (res.status !== "success" || !res.data) {
        setRecs([]);
        return toast(res.error ?? "Failed to load recommendations.");
      }
      setRecs(res.data);
      toast("Recommendations updated.");
    } finally {
      setBusy(false);
    }
  }

  async function recomputeTaste() {
    if (!user) return toast("Login to update your taste profile.");
    setBusy(true);
    try {
      const res = await apiSend<{ games_used: number }>(`/users/${user.id}/recompute-taste`);
      if (res.status !== "success" || !res.data) {
        return toast(res.error ?? "Failed to recompute taste.");
      }
      if (res.data.games_used === 0) {
        return toast("Like games first, then recompute taste.");
      }
      toast(`Taste updated from ${res.data.games_used} liked game(s).`);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className={styles.page}>
      <Section
        eyebrow="Drops"
        title="Premium Recommendations"
        subtitle="Curated pulls based on your taste profile."
      >
        <Card>
          <div className={styles.header}>
            <div>
              <div className={styles.title}>Drop Queue</div>
              <div className={styles.subtitle}>Recompute taste before pulling for best results.</div>
            </div>
            <Button onClick={loadRecommendations} disabled={!user || busy}>
              {busy ? "Loading..." : "Load Recs"}
            </Button>
            <Button variant="outline" onClick={recomputeTaste} disabled={!user || busy}>
              Recompute Taste
            </Button>
          </div>
          {msg && <div className={styles.toast}>{msg}</div>}
          {busy && recs.length === 0 ? (
            <div className={styles.loadingState}>
              <div className={styles.loadingPulse} />
              <span>Scanning the catalog for your best matches...</span>
            </div>
          ) : recs.length === 0 ? (
            <div className={styles.emptyState}>No recommendations yet.</div>
          ) : (
            <div className={styles.recs} aria-busy={busy}>
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
  );
}
