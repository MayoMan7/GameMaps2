import type { InputHTMLAttributes } from "react";
import styles from "./Input.module.css";

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
};

export default function Input({ label, className, ...props }: InputProps) {
  return (
    <label className={`${styles.wrapper}${className ? ` ${className}` : ""}`}>
      {label && <span className={styles.label}>{label}</span>}
      <input className={styles.input} {...props} />
    </label>
  );
}
