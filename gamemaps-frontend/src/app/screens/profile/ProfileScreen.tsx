"use client";

import { useState } from "react";
import styles from "./ProfileScreen.module.css";
import Section from "../../components/layout/Section";
import Card from "../../components/ui/Card";
import Input from "../../components/ui/Input";
import Button from "../../components/ui/Button";
import Badge from "../../components/ui/Badge";
import { apiSend } from "../../lib/api";
import { useAuth } from "../../lib/useAuth";

export default function ProfileScreen() {
  const { user, login, register, logout, refresh, error } = useAuth();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [newName, setNewName] = useState("");
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  function toast(text: string) {
    setMsg(text);
    window.setTimeout(() => setMsg(null), 2400);
  }

  async function handleLogin() {
    if (!email.trim()) return toast("Enter your email.");
    if (!password) return toast("Enter your password.");

    setBusy(true);
    try {
      const res = await login(email.trim(), password);
      if (res.status !== "success") {
        return toast(res.error ?? "Login failed.");
      }
      toast("Signed in.");
    } finally {
      setBusy(false);
    }
  }

  async function handleRegister() {
    if (!name.trim()) return toast("Enter a display name.");
    if (!email.trim()) return toast("Enter your email.");
    if (password.length < 6) return toast("Use a password with at least 6 characters.");

    setBusy(true);
    try {
      const res = await register(name.trim(), email.trim(), password);
      if (res.status !== "success") {
        return toast(res.error ?? "Registration failed.");
      }
      toast("Account created.");
    } finally {
      setBusy(false);
    }
  }

  async function handleUpdateName() {
    if (!user) return;
    if (!newName.trim()) return toast("Enter a new display name.");

    setBusy(true);
    try {
      const res = await apiSend(`/users/${user.id}`, { name: newName.trim() }, "PATCH");
      if (res.status !== "success") {
        return toast(res.error ?? "Failed to update name.");
      }
      await refresh();
      setNewName("");
      toast("Name updated.");
    } finally {
      setBusy(false);
    }
  }

  async function handleLogout() {
    setBusy(true);
    try {
      await logout();
      toast("Signed out.");
    } finally {
      setBusy(false);
    }
  }

  function submitAuth() {
    if (mode === "login") {
      void handleLogin();
    } else {
      void handleRegister();
    }
  }

  return (
    <div className={styles.page}>
      <Section
        eyebrow="Profile"
        title="User Management"
        subtitle="Secure access with premium presentation."
      >
        <div className={styles.grid}>
          <Card title={mode === "login" ? "Login" : "Register"} subtitle="Secure access to your taste profile.">
            <div className={styles.switchRow}>
              <Button
                size="sm"
                variant={mode === "login" ? "primary" : "ghost"}
                onClick={() => setMode("login")}
              >
                Login
              </Button>
              <Button
                size="sm"
                variant={mode === "register" ? "primary" : "ghost"}
                onClick={() => setMode("register")}
              >
                Register
              </Button>
            </div>
            {mode === "register" && (
              <Input
                label="Display name"
                placeholder="Your gamer handle"
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") submitAuth();
                }}
              />
            )}
            <Input
              label="Email"
              placeholder="you@domain.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") submitAuth();
              }}
            />
            <Input
              label="Password"
              type="password"
              placeholder="********"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") submitAuth();
              }}
            />
            <Button onClick={submitAuth} disabled={busy}>
              {mode === "login" ? "Login" : "Create Account"}
            </Button>
            {error && <div className={styles.error}>{error}</div>}
            {msg && <div className={styles.toast}>{msg}</div>}
          </Card>

          <Card title="Session Status" subtitle="Live player identity and controls.">
            {user ? (
              <>
                <div className={styles.profileRow}>
                  <div>
                    <div className={styles.profileName}>{user.name}</div>
                    <div className={styles.profileMeta}>{user.email}</div>
                  </div>
                  <Badge tone="accent">Online</Badge>
                </div>
                <div className={styles.formRow}>
                  <Input
                    label="Update display name"
                    placeholder="New handle"
                    value={newName}
                    onChange={(e) => setNewName(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") {
                        void handleUpdateName();
                      }
                    }}
                  />
                  <Button variant="outline" onClick={handleUpdateName} disabled={busy}>
                    Save
                  </Button>
                </div>
                <div className={styles.actionsRow}>
                  <Button variant="ghost" onClick={handleLogout} disabled={busy}>
                    Logout
                  </Button>
                </div>
              </>
            ) : (
              <div className={styles.emptyState}>No active session. Login or register.</div>
            )}
          </Card>
        </div>
      </Section>
    </div>
  );
}
