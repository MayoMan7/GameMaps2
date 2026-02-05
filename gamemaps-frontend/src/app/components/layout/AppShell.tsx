import type { ReactNode } from "react";
import Link from "next/link";
import styles from "./AppShell.module.css";

type AppShellProps = {
  children: ReactNode;
};

export default function AppShell({ children }: AppShellProps) {
  return (
    <div className={styles.shell}>
      <header className={styles.nav}>
        <Link href="/" className={styles.brand}>
          GameMaps
        </Link>
        <nav className={styles.links}>
          <Link href="/search">Search</Link>
          <Link href="/map">Taste Map</Link>
          <Link href="/recommendations">Recommendations</Link>
          <Link href="/profile">Profile</Link>
        </nav>
        <div className={styles.navGlow} />
      </header>
      <main className={styles.main}>{children}</main>
    </div>
  );
}
