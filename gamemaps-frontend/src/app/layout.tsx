import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "GoGameMaps Barebones",
  description: "Barebones frontend scaffold.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <header className="header">
          <h1>GoGameMaps</h1>
          <nav>
            <Link href="/">Home</Link>
            <Link href="/search">Search</Link>
            <Link href="/map">Map</Link>
            <Link href="/recommendations">Recommendations</Link>
            <Link href="/profile">Profile</Link>
          </nav>
        </header>
        <main className="content">{children}</main>
      </body>
    </html>
  );
}
