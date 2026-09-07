"use client";

import Link from "next/link";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";

import { DealCalculator } from "@/components/DealCalculator";
import {
  ApiError,
  getDeal,
  getListings,
  getMarket,
  getReferences,
  getTrim,
} from "@/lib/api";
import { formatARS, formatARSCompact, formatKm } from "@/lib/format";
import type {
  DealAssessment,
  Listing,
  ReferencePrice,
  TrimDetail,
  TrimMarketSummary,
} from "@/lib/types";
import { defaultYear, yearOptions } from "@/lib/year";

function parsePriceInput(value: string): number | null {
  const digits = value.replace(/\D/g, "");
  if (!digits) {
    return null;
  }
  const price = Number(digits);
  return Number.isFinite(price) && price > 0 ? price : null;
}

function referenceKindLabel(kind: string): string {
  if (kind === "fiscal") {
    return "Fiscal";
  }
  if (kind === "guide") {
    return "Guía";
  }
  return kind;
}

const COMPARE_IDS_KEY = "compare_trim_ids";

function readCompareIds(): number[] {
  return (sessionStorage.getItem(COMPARE_IDS_KEY) ?? "")
    .split(",")
    .map(Number)
    .filter((id) => Number.isFinite(id) && id > 0)
    .slice(0, 3);
}

function addCompareId(trimId: number): number[] {
  const ids = readCompareIds();
  if (!ids.includes(trimId) && ids.length < 3) {
    ids.push(trimId);
    sessionStorage.setItem(COMPARE_IDS_KEY, ids.join(","));
  }
  return ids;
}

function groupFeaturesByCategory(
  features: TrimDetail["features"],
): Map<string, TrimDetail["features"]> {
  const grouped = new Map<string, TrimDetail["features"]>();
  for (const feature of features) {
    const list = grouped.get(feature.category) ?? [];
    list.push(feature);
    grouped.set(feature.category, list);
  }
  return grouped;
}

export function VehicleSheet() {
  const params = useParams<{ id: string }>();
  const searchParams = useSearchParams();
  const router = useRouter();
  const trimId = Number(params.id);
  const invalidId = !Number.isFinite(trimId) || trimId <= 0;

  const [trim, setTrim] = useState<TrimDetail | null>(null);
  const [year, setYear] = useState<number | null>(null);
  const [market, setMarket] = useState<TrimMarketSummary | null>(null);
  const [references, setReferences] = useState<ReferencePrice[]>([]);
  const [listings, setListings] = useState<Listing[]>([]);
  const [deal, setDeal] = useState<DealAssessment | null>(null);
  const [priceInput, setPriceInput] = useState("");
  const [dealLoading, setDealLoading] = useState(false);

  const [trimLoading, setTrimLoading] = useState(!invalidId);
  const [dataLoading, setDataLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notFound, setNotFound] = useState(invalidId);

  const dealTimeoutRef = useRef<number | null>(null);

  const refreshYearData = useCallback(
    async (selectedYear: number) => {
      setDataLoading(true);
      setError(null);
      try {
        const [marketData, referencesData, listingsData] = await Promise.all([
          getMarket(trimId, selectedYear),
          getReferences(trimId, selectedYear),
          getListings(trimId, selectedYear),
        ]);
        setMarket(marketData);
        setReferences(referencesData);
        setListings(listingsData);
      } catch {
        setError("No pudimos cargar los datos de mercado para este año.");
      } finally {
        setDataLoading(false);
      }
    },
    [trimId],
  );

  useEffect(() => {
    if (invalidId) {
      return;
    }

    let cancelled = false;
    const queryYear = Number(searchParams.get("year"));
    const initialYear = Number.isFinite(queryYear) && queryYear > 0 ? queryYear : undefined;

    getTrim(trimId, initialYear)
      .then(async (detail) => {
        if (cancelled) {
          return;
        }
        const selectedYear = defaultYear(detail, initialYear);
        setTrim(detail);
        setYear(selectedYear);
        await refreshYearData(selectedYear);
      })
      .catch((err) => {
        if (cancelled) {
          return;
        }
        if (err instanceof ApiError && err.status === 404) {
          setNotFound(true);
          return;
        }
        setError("No pudimos cargar esta versión.");
      })
      .finally(() => {
        if (!cancelled) {
          setTrimLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
    // Initial year comes from the URL only on first load; year changes refresh data via onYearChange.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [invalidId, trimId, refreshYearData]);

  function onYearChange(nextYear: number) {
    setYear(nextYear);
    setDeal(null);
    const nextParams = new URLSearchParams(searchParams.toString());
    nextParams.set("year", String(nextYear));
    router.replace(`/trims/${trimId}?${nextParams.toString()}`, { scroll: false });
    void refreshYearData(nextYear);
  }

  function onPriceChange(value: string) {
    setPriceInput(value);

    if (dealTimeoutRef.current != null) {
      window.clearTimeout(dealTimeoutRef.current);
    }

    const price = parsePriceInput(value);
    if (price == null || year == null) {
      setDeal(null);
      setDealLoading(false);
      return;
    }

    setDealLoading(true);
    dealTimeoutRef.current = window.setTimeout(() => {
      getDeal(trimId, year, price)
        .then(setDeal)
        .catch(() => setDeal(null))
        .finally(() => setDealLoading(false));
    }, 400);
  }

  if (invalidId || notFound) {
    return (
      <main className="mx-auto w-full max-w-3xl px-6 py-16">
        <Link href="/" className="text-sm font-medium text-emerald-700 hover:underline">
          ← Volver a buscar
        </Link>
        <p className="mt-8 text-zinc-600">Versión no encontrada.</p>
      </main>
    );
  }

  if (trimLoading) {
    return (
      <main className="mx-auto w-full max-w-3xl px-6 py-16">
        <p className="text-zinc-600">Cargando versión…</p>
      </main>
    );
  }

  if (!trim || year == null) {
    return (
      <main className="mx-auto w-full max-w-3xl px-6 py-16">
        <Link href="/" className="text-sm font-medium text-emerald-700 hover:underline">
          ← Volver a buscar
        </Link>
        <p className="mt-8 text-zinc-600">
          {error ?? "No pudimos cargar esta versión."}
        </p>
      </main>
    );
  }

  const years = yearOptions(trim.year_from, trim.year_to);
  const featuresByCategory = groupFeaturesByCategory(trim.features);

  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-10 sm:py-16">
      <Link href="/" className="text-sm font-medium text-emerald-700 hover:underline">
        ← Volver a buscar
      </Link>

      <header className="mt-6">
        <p className="text-sm font-semibold uppercase tracking-wide text-emerald-700">
          {trim.brand_name}
        </p>
        <h1 className="mt-1 text-3xl font-semibold tracking-tight text-zinc-900 sm:text-4xl">
          {trim.model_name} {trim.trim_name}
        </h1>
        <p className="mt-2 text-zinc-600">
          {trim.generation_name}
          {trim.year_from != null && trim.year_to != null
            ? ` · ${trim.year_from}–${trim.year_to}`
            : ""}
        </p>

        <div className="mt-4 flex flex-wrap items-end gap-3">
          {years.length > 0 ? (
            <div>
              <label htmlFor="year-select" className="text-sm text-zinc-600">
                Año
              </label>
              <select
                id="year-select"
                value={year}
                onChange={(event) => onYearChange(Number(event.target.value))}
                className="mt-1 block rounded-xl border border-zinc-300 bg-white px-4 py-2 text-base outline-none ring-emerald-600/30 focus:border-emerald-600 focus:ring-2"
              >
                {years.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
            </div>
          ) : null}
          <button
            type="button"
            onClick={() => {
              const ids = addCompareId(trimId);
              router.push(`/compare?trim_ids=${ids.join(",")}&year=${year}`);
            }}
            className="rounded-xl bg-emerald-700 px-5 py-2 text-base font-medium text-white transition hover:bg-emerald-800"
          >
            Comparar
          </button>
        </div>
      </header>

      {error ? (
        <p className="mt-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          {error}
        </p>
      ) : null}

      <div className="mt-8 grid gap-6">
        <section className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-zinc-500">
            Mercado
          </h2>
          {dataLoading ? (
            <p className="mt-4 text-sm text-zinc-500">Cargando mercado…</p>
          ) : market && market.count > 0 ? (
            <dl className="mt-4 grid gap-3 sm:grid-cols-2">
              <div>
                <dt className="text-sm text-zinc-500">Mediana</dt>
                <dd className="text-xl font-semibold text-zinc-900">
                  {market.median != null ? formatARSCompact(market.median) : "—"}
                </dd>
              </div>
              <div>
                <dt className="text-sm text-zinc-500">Publicaciones</dt>
                <dd className="text-xl font-semibold text-zinc-900">{market.count}</dd>
              </div>
              <div className="sm:col-span-2">
                <dt className="text-sm text-zinc-500">Rango</dt>
                <dd className="text-base text-zinc-900">
                  {market.minimum != null && market.maximum != null
                    ? `${formatARSCompact(market.minimum)} – ${formatARSCompact(market.maximum)}`
                    : "—"}
                </dd>
              </div>
            </dl>
          ) : (
            <p className="mt-4 text-sm text-zinc-600">
              No hay publicaciones suficientes para estimar el mercado de esta versión/año.
            </p>
          )}
        </section>

        <DealCalculator
          priceInput={priceInput}
          onPriceChange={onPriceChange}
          deal={deal}
          loading={dealLoading}
        />

        <p className="text-sm text-zinc-600">
          {deal?.disclaimer ??
            "Estimación a partir de publicaciones comparables; el precio publicado no es precio de venta. Catálogo de demostración."}
        </p>

        <section className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-zinc-500">
            Referencias
          </h2>
          {dataLoading ? (
            <p className="mt-4 text-sm text-zinc-500">Cargando referencias…</p>
          ) : references.length === 0 ? (
            <p className="mt-4 text-sm text-zinc-600">Sin referencias para este año.</p>
          ) : (
            <div className="mt-4 overflow-x-auto">
              <table className="w-full min-w-[20rem] text-left text-sm">
                <thead>
                  <tr className="border-b border-zinc-200 text-zinc-500">
                    <th className="py-2 pr-4 font-medium">Fuente</th>
                    <th className="py-2 pr-4 font-medium">Tipo</th>
                    <th className="py-2 font-medium">Precio</th>
                  </tr>
                </thead>
                <tbody>
                  {references.map((reference) => (
                    <tr key={reference.id} className="border-b border-zinc-100">
                      <td className="py-3 pr-4 text-zinc-900">{reference.source}</td>
                      <td className="py-3 pr-4 text-zinc-600">
                        {referenceKindLabel(reference.kind)}
                      </td>
                      <td className="py-3 font-medium text-zinc-900">
                        {formatARS(reference.price)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>

        <section className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-zinc-500">
            Ficha técnica
          </h2>
          {trim.specs ? (
            <dl className="mt-4 grid gap-3 sm:grid-cols-2">
              <div>
                <dt className="text-sm text-zinc-500">Motor</dt>
                <dd className="text-zinc-900">{trim.specs.engine}</dd>
              </div>
              <div>
                <dt className="text-sm text-zinc-500">Transmisión</dt>
                <dd className="text-zinc-900">{trim.specs.transmission}</dd>
              </div>
              <div>
                <dt className="text-sm text-zinc-500">Combustible</dt>
                <dd className="text-zinc-900">{trim.specs.fuel}</dd>
              </div>
              {trim.specs.horsepower != null ? (
                <div>
                  <dt className="text-sm text-zinc-500">Potencia</dt>
                  <dd className="text-zinc-900">{trim.specs.horsepower} CV</dd>
                </div>
              ) : null}
            </dl>
          ) : (
            <p className="mt-4 text-sm text-zinc-600">
              Ficha técnica no cargada para esta versión.
            </p>
          )}
        </section>

        <section className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-zinc-500">
            Equipamiento
          </h2>
          {featuresByCategory.size > 0 ? (
            <div className="mt-4 space-y-4">
              {[...featuresByCategory.entries()].map(([category, features]) => (
                <div key={category}>
                  <h3 className="text-sm font-medium text-zinc-800">{category}</h3>
                  <ul className="mt-2 space-y-1 text-sm text-zinc-600">
                    {features.map((feature) => (
                      <li key={feature.code}>
                        {feature.name}
                        {feature.value ? ` — ${feature.value}` : ""}
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          ) : (
            <p className="mt-4 text-sm text-zinc-600">
              Sin equipamiento cargado.
            </p>
          )}
        </section>

        <section className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-zinc-500">
            Publicaciones actuales
          </h2>
          {dataLoading ? (
            <p className="mt-4 text-sm text-zinc-500">Cargando publicaciones…</p>
          ) : listings.length === 0 ? (
            <p className="mt-4 text-sm text-zinc-600">Sin publicaciones para este año.</p>
          ) : (
            <ul className="mt-4 divide-y divide-zinc-100">
              {listings.map((listing) => (
                <li
                  key={listing.id}
                  className="flex flex-col gap-1 py-3 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div>
                    <p className="font-medium text-zinc-900">{formatARS(listing.price)}</p>
                    <p className="text-sm text-zinc-600">
                      {listing.km != null ? formatKm(listing.km) : "Km no informado"}
                      {listing.location ? ` · ${listing.location}` : ""}
                    </p>
                  </div>
                  {listing.url ? (
                    <a
                      href={listing.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm font-medium text-emerald-700 hover:underline"
                    >
                      Ver publicación
                    </a>
                  ) : null}
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </main>
  );
}
