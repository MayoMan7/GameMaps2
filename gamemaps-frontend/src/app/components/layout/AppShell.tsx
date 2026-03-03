"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import styles from "./AppShell.module.css";

type AppShellProps = {
  children: ReactNode;
};

export default function AppShell({ children }: AppShellProps) {
  const pathname = usePathname();
  const navItems = [
    { href: "/search", label: "Search" },
    { href: "/map", label: "Taste Map" },
    { href: "/recommendations", label: "Recommendations" },
    { href: "/profile", label: "Profile" },
  ];

  return (
    <div className={styles.shell}>
      <header className={styles.nav}>
        <Link href="/" className={styles.brand}>
          GameMaps
        </Link>
        <nav className={styles.links}>
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={`${styles.navLink}${pathname === item.href ? ` ${styles.active}` : ""}`}
            >
              {item.label}
            </Link>
          ))}
        </nav>
        <div className={styles.navGlow} />
      </header>
      <main className={styles.main}>{children}</main>
    </div>
  );
}
