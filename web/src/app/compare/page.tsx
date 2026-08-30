import Link from "next/link";

import { formatARSCompact } from "@/lib/format";

type CompareSpecs = {
  engine: string;
  transmission: string;
  horsepower: number | null;
};

type CompareMarket = {
  count: number;
  median: number | null;
  currency: string;
};

type CompareTrim = {
  trim_id: number;
  trim_name: string;
  brand_name: string;
  model_name: string;
  specs: CompareSpecs | null;
  market: CompareMarket;
};

type CompareFeature = {
  code: string;
  name: string;
  category: string;
  values: Record<string, string>;
};

type CompareResult = {
  year: number;
  trims: CompareTrim[];
  features: CompareFeature[];
};

type ApiErrorBody = {
  code: string;
  message: string;
};

function queryValue(value: string | string[] | undefined): string {
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }
  return value ?? "";
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

async function loadCompare(trimIds: string, year: string): Promise<
  { ok: true; data: CompareResult } | { ok: false; status: number; message: string }
> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
  const params = new URLSearchParams({ trim_ids: trimIds, year });
  const response = await fetch(`${apiUrl}/api/v1/compare?${params}`);
  if (!response.ok) {
    let message = response.statusText;
    try {
      const body = (await response.json()) as ApiErrorBody;
      message = body.message || message;
    } catch {
      // keep statusText
    }
    return { ok: false, status: response.status, message };
  }
  return { ok: true, data: (await response.json()) as CompareResult };
}

export default async function ComparePage({
  searchParams,
}: {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  const params = await searchParams;
  const trimIds = queryValue(params.trim_ids);
  const year = queryValue(params.year);

  const idCount = trimIds
    .split(",")
    .map((part) => part.trim())
    .filter(Boolean).length;

  if (!trimIds || !year || idCount < 2) {
    return (
      <main className="mx-auto w-full max-w-4xl px-6 py-16">
        <Link href="/" className="text-sm font-medium text-emerald-700 hover:underline">
          ← Volver a buscar
        </Link>
        <h1 className="mt-8 text-3xl font-semibold tracking-tight text-zinc-900">Comparar</h1>
        <p className="mt-4 text-zinc-600">
          {idCount === 1 ? (
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

  const result = await loadCompare(trimIds, year);
  if (!result.ok) {
    return (
      <main className="mx-auto w-full max-w-4xl px-6 py-16">
        <Link href="/" className="text-sm font-medium text-emerald-700 hover:underline">
          ← Volver a buscar
        </Link>
        <p className="mt-8 text-zinc-600">
          {result.status === 404 ? "Versión no encontrada." : result.message}
        </p>
      </main>
    );
  }

  const { data } = result;
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
