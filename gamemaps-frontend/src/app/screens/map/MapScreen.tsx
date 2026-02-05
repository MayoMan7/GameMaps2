"use client";

import { useEffect, useState } from "react";
import styles from "./MapScreen.module.css";
import Button from "../../components/ui/Button";
import TasteMap, { MapPayload } from "../../components/map/TasteMap";
import { apiGet } from "../../lib/api";
import { useAuth } from "../../lib/useAuth";

export default function MapScreen() {
  const { user } = useAuth();
  const [payload, setPayload] = useState<MapPayload | null>(null);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);
  const [zoom, setZoom] = useState(1);

  function toast(text: string) {
    setMsg(text);
    window.setTimeout(() => setMsg(null), 2400);
  }

  async function loadMap() {
    if (!user) return toast("Login to load your taste map.");
    setBusy(true);
    try {
      const res = await apiGet<MapPayload>(`/users/${user.id}/map`);
      if (res.status !== "success" || !res.data) {
        setPayload(null);
        return toast(res.error ?? "Failed to load map.");
      }
      setPayload(res.data);
      toast("Taste map loaded.");
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    if (user) {
      loadMap();
    } else {
      setPayload(null);
    }
  }, [user]);

  return (
    <div className={styles.page}>
      <div className={styles.mapShell}>
        <div className={styles.mapHud}>
          <div>
            <div className={styles.mapTitle}>Taste Profile Map</div>
            <div className={styles.mapSubtitle}>Hover nodes to focus clusters.</div>
          </div>
          <div className={styles.mapControls}>
            <Button variant="ghost" onClick={() => setZoom((z) => Math.max(0.6, Number((z - 0.2).toFixed(2))))}>
              -
            </Button>
            <Button variant="ghost" onClick={() => setZoom((z) => Math.min(2.2, Number((z + 0.2).toFixed(2))))}>
              +
            </Button>
            <Button onClick={loadMap} disabled={!user || busy}>
              {busy ? "Loading..." : "Reload"}
            </Button>
          </div>
        </div>
        {msg && <div className={styles.toast}>{msg}</div>}
        {busy && !payload ? (
          <div className={styles.loadingState}>
            <div className={styles.loadingGlow} />
            <span>Rendering your taste clusters...</span>
          </div>
        ) : payload ? (
          <TasteMap payload={payload} zoom={zoom} onZoomChange={setZoom} />
        ) : (
          <div className={styles.emptyState}>Login to load your taste map.</div>
        )}
        <div className={styles.legend}>
          <div className={styles.legendItem}>
            <span className={`${styles.legendDot} ${styles.legendUser}`} />
            User core
          </div>
          <div className={styles.legendItem}>
            <span className={`${styles.legendDot} ${styles.legendLiked}`} />
            Liked anchor
          </div>
          <div className={styles.legendItem}>
            <span className={`${styles.legendDot} ${styles.legendRec}`} />
            Recommendation
          </div>
        </div>
      </div>
    </div>
  );
}
