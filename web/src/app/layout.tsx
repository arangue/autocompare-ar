import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";

import { SiteNav } from "@/components/SiteNav";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "AutoCompare",
  description:
    "Herramienta para saber si un usado en Argentina está barato. No es un clasificado.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="es"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
      suppressHydrationWarning
    >
      <body
        className="flex min-h-full flex-col bg-zinc-50 text-zinc-900"
        suppressHydrationWarning
      >
        <SiteNav />
        {children}
        <footer className="mt-auto border-t border-zinc-200 px-6 py-4 text-xs text-zinc-500">
          Estimación a partir de publicaciones comparables; el precio publicado
          no es precio de venta.
        </footer>
      </body>
    </html>
  );
}
