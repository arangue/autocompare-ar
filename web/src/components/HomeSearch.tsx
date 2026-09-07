"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

import { ApiError, searchTrims } from "@/lib/api";
import type { TrimSearchResult } from "@/lib/types";

function trimHref(result: TrimSearchResult): string {
  if (result.year_from != null) {
    return `/trims/${result.trim_id}?year=${result.year_from}`;
  }
  return `/trims/${result.trim_id}`;
}

function yearLabel(result: TrimSearchResult): string {
  if (result.year_from != null && result.year_to != null) {
    return `${result.year_from}–${result.year_to}`;
  }
  if (result.year_from != null) {
    return `${result.year_from}+`;
  }
  return "";
}

export function HomeSearch() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<TrimSearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const q = query.trim();
    if (!q) {
      setResults([]);
      setSearched(false);
      setError(null);
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const data = await searchTrims(q);
      setResults(data);
      setSearched(true);
    } catch (err) {
      setResults([]);
      setSearched(true);
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("No pudimos buscar versiones.");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="w-full max-w-xl">
      <form onSubmit={onSubmit} className="flex flex-col gap-3 sm:flex-row">
        <label htmlFor="trim-search" className="sr-only">
          Buscar versión
        </label>
        <input
          id="trim-search"
          type="search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Buscar versión (ej. Corolla XEi)"
          className="min-w-0 flex-1 rounded-xl border border-zinc-300 bg-white px-4 py-3 text-base text-zinc-900 shadow-sm outline-none ring-emerald-600/30 placeholder:text-zinc-400 focus:border-emerald-600 focus:ring-2"
          autoComplete="off"
        />
        <button
          type="submit"
          disabled={loading}
          className="rounded-xl bg-emerald-700 px-5 py-3 text-base font-medium text-white transition hover:bg-emerald-800 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {loading ? "Buscando…" : "Buscar"}
        </button>
      </form>

      {error ? (
        <p className="mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {error}
        </p>
      ) : null}

      {searched && !error ? (
        <div className="mt-8">
          {results.length === 0 ? (
            <p className="text-sm text-zinc-600">
              No encontramos versiones para &ldquo;{query.trim()}&rdquo;.
            </p>
          ) : (
            <ul className="divide-y divide-zinc-200 overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-sm">
              {results.map((result) => (
                <li key={result.trim_id}>
                  <Link
                    href={trimHref(result)}
                    className="block px-4 py-4 transition hover:bg-zinc-50"
                  >
                    <p className="font-medium text-zinc-900">
                      {result.brand_name} {result.model_name}
                    </p>
                    <p className="text-sm text-zinc-600">
                      {result.generation_name} · {result.trim_name}
                      {yearLabel(result) ? ` · ${yearLabel(result)}` : ""}
                    </p>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
      ) : null}
    </div>
  );
}
