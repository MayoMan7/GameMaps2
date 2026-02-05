"use client";

import styles from "./LandingScreen.module.css";
import Container from "../../components/layout/Container";
import Section from "../../components/layout/Section";
import Card from "../../components/ui/Card";
import Button from "../../components/ui/Button";
import Stat from "../../components/ui/Stat";
import Badge from "../../components/ui/Badge";
import Link from "next/link";
import { useAuth } from "../../lib/useAuth";

export default function LandingScreen() {
  const { user } = useAuth();

  return (
    <div className={styles.page}>
      <Container>
        <header className={styles.hero}>
          <div className={styles.heroCopy}>
            <div className={styles.heroBadge}>
              <Badge tone="hot">Luxury Beta</Badge>
              <span>Precision taste mapping for elite players.</span>
            </div>
            <h1 className={styles.title}>Taste meets telemetry.</h1>
            <p className={styles.lead}>
              GameMaps builds a living profile of your taste, clusters your favorite titles, and streams
              recommendations like a premium launch reel.
            </p>
            <div className={styles.heroActions}>
              <Link href="/search">
                <Button>Start a Search</Button>
              </Link>
              <Link href="/map">
                <Button variant="outline">View Taste Map</Button>
              </Link>
            </div>
            <div className={styles.heroFootnote}>
              {user ? `Signed in as ${user.name}` : "Sign in to sync your profile across devices."}
            </div>
          </div>
          <div className={styles.heroPanel}>
            <div className={styles.heroStats}>
              <Stat label="Mode" value="Taste Radar" helper="Live clustering engine" />
              <Stat label="Signal" value="High Fidelity" helper="TF-IDF + cosine match" />
              <Stat label="Map Style" value="Skill Tree" helper="Anchored recommendations" />
              <Stat label="Session" value={user ? "Active" : "Offline"} helper="Secure profile sync" />
            </div>
          </div>
        </header>

        <div className={styles.grid}>
          <Section
            eyebrow="Pipeline"
            title="A premium data path"
            subtitle="Every tap sharpens your map and focuses the next drop."
          >
            <div className={styles.cardRow}>
              <Card title="Discover" subtitle="Typeahead search with instant scouting.">
                <p className={styles.cardText}>
                  Autocomplete results appear as you type. Lock in a game and feed the taste engine.
                </p>
              </Card>
              <Card title="Cluster" subtitle="See the map with anchor precision.">
                <p className={styles.cardText}>
                  Each recommendation attaches to its closest liked game, forming real clusters.
                </p>
              </Card>
              <Card title="Deploy" subtitle="Get recs tuned to your profile.">
                <p className={styles.cardText}>
                  Pull a curated list that scales with how much you like and explore.
                </p>
              </Card>
            </div>
          </Section>

          <Section
            eyebrow="Status"
            title="Your profile control"
            subtitle="Manage identity, likes, and taste history in one place."
          >
            <Card>
              <div className={styles.profilePanel}>
                <div>
                  <h3 className={styles.profileTitle}>Profile Access</h3>
                  <p className={styles.profileText}>
                    Use the Profile page to log in, update your handle, and review liked games.
                  </p>
                </div>
                <Link href="/profile">
                  <Button variant="ghost">Open Profile</Button>
                </Link>
              </div>
            </Card>
          </Section>
        </div>
      </Container>
    </div>
  );
}
