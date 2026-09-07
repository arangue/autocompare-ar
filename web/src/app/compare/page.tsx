import Link from "next/link";

import { ApiError, getCompare } from "@/lib/api";
import { formatARSCompact } from "@/lib/format";
import type { CompareTrim } from "@/lib/types";

export const metadata = { title: "Comparar | AutoCompare" };

function queryValue(value: string | string[] | undefined): string {
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }
  return value ?? "";
}

function parseTrimIds(raw: string): number[] {
  return raw
    .split(",")
    .map((part) => Number(part.trim()))
    .filter((id) => Number.isFinite(id) && id > 0);
}

function specCell(trim: CompareTrim, field: "engine" | "transmission" | "horsepower"): string {
  if (!trim.specs) {
    return "—";
  }
  if (field === "horsepower") {
    return trim.specs.horsepower != null ? `${trim.specs.horsepower} CV` : "—";
  }
  return trim.specs[field] || "—";
}

export default async function ComparePage({
  searchParams,
}: {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  const params = await searchParams;
  const trimIdsRaw = queryValue(params.trim_ids);
  const yearRaw = queryValue(params.year);
  const trimIds = parseTrimIds(trimIdsRaw);
  const year = Number(yearRaw);

  if (trimIds.length < 2 || !Number.isFinite(year) || year <= 0) {
    return (
      <main className="mx-auto w-full max-w-4xl px-6 py-16">
        <Link href="/" className="text-sm font-medium text-emerald-700 hover:underline">
          ← Volver a buscar
        </Link>
        <h1 className="mt-8 text-3xl font-semibold tracking-tight text-zinc-900">Comparar</h1>
        <p className="mt-4 text-zinc-600">
          {trimIds.length === 1 ? (
            <>
              Versión agregada.{" "}
              <Link href="/" className="font-medium text-emerald-700 hover:underline">
                Buscá otra
              </Link>{" "}
              y tocá Comparar.
            </>
          ) : (
            <>
              Elegí 2 o 3 versiones y un año. Ejemplo:{" "}
              <Link
                href="/compare?trim_ids=1,2&year=2019"
                className="font-medium text-emerald-700 hover:underline"
              >
                Corolla XEi vs XLi 2019
              </Link>
              .
            </>
          )}
        </p>
      </main>
    );
  }

  let data;
  try {
    data = await getCompare(trimIds, year);
  } catch (err) {
    const message =
      err instanceof ApiError
        ? err.status === 404
          ? "Versión no encontrada."
          : err.message
        : "No pudimos cargar la comparación.";
    return (
      <main className="mx-auto w-full max-w-4xl px-6 py-16">
        <Link href="/" className="text-sm font-medium text-emerald-700 hover:underline">
          ← Volver a buscar
        </Link>
        <p className="mt-8 text-zinc-600">{message}</p>
      </main>
    );
  }

  const specRows = [
    { label: "Motor", field: "engine" as const },
    { label: "Transmisión", field: "transmission" as const },
    { label: "Potencia", field: "horsepower" as const },
  ];

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10 sm:py-16">
      <Link href="/" className="text-sm font-medium text-emerald-700 hover:underline">
        ← Volver a buscar
      </Link>

      <h1 className="mt-6 text-3xl font-semibold tracking-tight text-zinc-900 sm:text-4xl">
        Comparar · {data.year}
      </h1>

      <div className="mt-8 overflow-x-auto rounded-xl border border-zinc-200 bg-white shadow-sm">
        <table className="w-full min-w-[36rem] text-left text-sm">
          <thead>
            <tr className="border-b border-zinc-200">
              <th className="px-4 py-3 font-medium text-zinc-500"> </th>
              {data.trims.map((trim) => (
                <th key={trim.trim_id} className="px-4 py-3 align-top">
                  <Link
                    href={`/trims/${trim.trim_id}?year=${data.year}`}
                    className="font-semibold text-zinc-900 hover:underline"
                  >
                    {trim.brand_name} {trim.model_name}
                  </Link>
                  <p className="mt-1 font-normal text-zinc-600">{trim.trim_name}</p>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {specRows.map((row) => (
              <tr key={row.field} className="border-b border-zinc-100">
                <th className="px-4 py-3 font-medium text-zinc-500">{row.label}</th>
                {data.trims.map((trim) => (
                  <td key={trim.trim_id} className="px-4 py-3 text-zinc-900">
                    {specCell(trim, row.field)}
                  </td>
                ))}
              </tr>
            ))}
            <tr className="border-b border-zinc-100">
              <th className="px-4 py-3 font-medium text-zinc-500">Mediana</th>
              {data.trims.map((trim) => (
                <td key={trim.trim_id} className="px-4 py-3 font-medium text-zinc-900">
                  {trim.market.median != null ? formatARSCompact(trim.market.median) : "—"}
                  <span className="ml-1 font-normal text-zinc-500">
                    ({trim.market.count})
                  </span>
                </td>
              ))}
            </tr>
            {data.features.map((feature) => (
              <tr key={feature.code} className="border-b border-zinc-100">
                <th className="px-4 py-3 font-medium text-zinc-500">
                  {feature.name}
                  <span className="ml-1 font-normal text-zinc-400">{feature.category}</span>
                </th>
                {data.trims.map((trim) => (
                  <td key={trim.trim_id} className="px-4 py-3 text-zinc-900">
                    {feature.values[String(trim.trim_id)] ?? "—"}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </main>
  );
}
