import type { ReactNode } from "react";
import styles from "./Section.module.css";

type SectionProps = {
  eyebrow?: string;
  title: string;
  subtitle?: string;
  children: ReactNode;
  className?: string;
};

export default function Section({ eyebrow, title, subtitle, children, className }: SectionProps) {
  return (
    <section className={`${styles.section}${className ? ` ${className}` : ""}`}>
      <header className={styles.header}>
        {eyebrow && <span className={styles.eyebrow}>{eyebrow}</span>}
        <h2 className={styles.title}>{title}</h2>
        {subtitle && <p className={styles.subtitle}>{subtitle}</p>}
      </header>
      {children}
    </section>
  );
}
