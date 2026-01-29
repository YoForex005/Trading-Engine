import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "RTX Admin - Broker Terminal",
  description: "Broker administration dashboard for RTX Trading Engine",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className="antialiased">
        {children}
      </body>
    </html>
  );
}
