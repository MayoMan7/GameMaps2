import type { ReactNode } from "react";
import styles from "./Stat.module.css";

type StatProps = {
  label: string;
  value: ReactNode;
  helper?: string;
};

export default function Stat({ label, value, helper }: StatProps) {
  return (
    <div className={styles.stat}>
      <span className={styles.label}>{label}</span>
      <span className={styles.value}>{value}</span>
      {helper && <span className={styles.helper}>{helper}</span>}
    </div>
  );
}
