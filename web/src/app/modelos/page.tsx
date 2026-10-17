import Link from "next/link";

import { ApiError, listModelOffer } from "@/lib/api";
import { formatDateART } from "@/lib/format";

export const metadata = { title: "Modelos | AutoCompare" };

export default async function ModelosPage() {
  let data;
  try {
    data = await listModelOffer();
  } catch (err) {
    const message =
      err instanceof ApiError
        ? err.message
        : "No pudimos cargar los modelos con oferta.";
    return (
      <main className="mx-auto w-full max-w-3xl px-6 py-16">
        <h1 className="text-3xl font-semibold tracking-tight text-zinc-900">
          Modelos con oferta
        </h1>
        <p className="mt-4 text-zinc-600">{message}</p>
      </main>
    );
  }

  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-10 sm:py-16">
      <h1 className="text-3xl font-semibold tracking-tight text-zinc-900 sm:text-4xl">
        Modelos con oferta
      </h1>
      <p className="mt-3 text-zinc-600">
        Catálogo canónico, ordenado por avisos usados activos. El precio de
        mercado está en cada versión y año; acá no hay mediana de modelo.
      </p>
      <p className="mt-2 text-sm text-zinc-500">
        {data.observed_at ? `Corte ${formatDateART(data.observed_at)}` : "Sin corte"}
        {data.models.length > 0 ? ` · ${data.models.length} modelos` : ""}
      </p>

      {data.models.length === 0 ? (
        <p className="mt-8 text-sm text-zinc-600">Todavía no hay avisos usados.</p>
      ) : (
        <ul className="mt-8 divide-y divide-zinc-200 overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-sm">
          {data.models.map((model) => (
            <li key={model.model_id}>
              <Link
                href={`/modelos/${model.model_id}`}
                className="flex items-baseline justify-between gap-4 px-4 py-4 transition hover:bg-zinc-50"
              >
                <span>
                  <span className="font-medium text-zinc-900">
                    {model.brand_name} {model.model_name}
                  </span>
                </span>
                <span className="shrink-0 text-sm text-zinc-500">
                  {model.count} avisos
                </span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
