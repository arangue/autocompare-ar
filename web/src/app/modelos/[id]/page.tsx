import Link from "next/link";

import { ApiError, getModelOffer } from "@/lib/api";
import { formatDateART } from "@/lib/format";

export const metadata = { title: "Modelo | AutoCompare" };

export default async function ModeloPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const modelId = Number(id);
  if (!Number.isFinite(modelId) || modelId <= 0) {
    return (
      <main className="mx-auto w-full max-w-3xl px-6 py-16">
        <Link href="/modelos" className="text-sm font-medium text-emerald-700 hover:underline">
          ← Modelos
        </Link>
        <p className="mt-8 text-zinc-600">Modelo no encontrado.</p>
      </main>
    );
  }

  let data;
  try {
    data = await getModelOffer(modelId);
  } catch (err) {
    const message =
      err instanceof ApiError
        ? err.status === 404
          ? "Modelo no encontrado."
          : err.message
        : "No pudimos cargar este modelo.";
    return (
      <main className="mx-auto w-full max-w-3xl px-6 py-16">
        <Link href="/modelos" className="text-sm font-medium text-emerald-700 hover:underline">
          ← Modelos
        </Link>
        <p className="mt-8 text-zinc-600">{message}</p>
      </main>
    );
  }

  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-10 sm:py-16">
      <Link href="/modelos" className="text-sm font-medium text-emerald-700 hover:underline">
        ← Modelos
      </Link>
      <p className="mt-6 text-sm font-semibold uppercase tracking-wide text-emerald-700">
        {data.brand_name}
      </p>
      <h1 className="mt-1 text-3xl font-semibold tracking-tight text-zinc-900 sm:text-4xl">
        {data.model_name}
      </h1>
      <p className="mt-3 text-zinc-600">
        {data.count} avisos usados
        {data.observed_at ? ` · corte ${formatDateART(data.observed_at)}` : ""}
      </p>
      <p className="mt-2 text-sm text-zinc-500">
        Elegí versión y año. La mediana y el trato viven ahí, no en el modelo
        entero.
      </p>

      {data.trims.length === 0 ? (
        <p className="mt-8 text-sm text-zinc-600">Sin avisos usados para este modelo.</p>
      ) : (
        <ul className="mt-8 space-y-4">
          {data.trims.map((trim) => (
            <li
              key={trim.trim_id}
              className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm"
            >
              <p className="font-medium text-zinc-900">{trim.trim_name}</p>
              {trim.generation_name !== data.model_name ? (
                <p className="mt-1 text-sm text-zinc-500">{trim.generation_name}</p>
              ) : null}
              <p className="mt-1 text-sm text-zinc-500">{trim.count} avisos</p>
              <div className="mt-3 flex flex-wrap gap-2">
                {trim.years.map((year) => (
                  <Link
                    key={year.year}
                    href={`/trims/${trim.trim_id}?year=${year.year}`}
                    className="rounded-lg border border-zinc-300 bg-white px-3 py-1.5 text-sm text-zinc-800 hover:border-emerald-600 hover:text-emerald-800"
                  >
                    {year.year}
                    <span className="ml-1 text-zinc-500">({year.count})</span>
                  </Link>
                ))}
              </div>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
