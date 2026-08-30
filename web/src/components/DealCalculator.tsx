import type { DealAssessment } from "@/lib/types";
import { formatARS } from "@/lib/format";

type DealBadgeProps = {
  deal: DealAssessment | null;
  loading: boolean;
};

export function DealBadge({ deal, loading }: DealBadgeProps) {
  if (loading) {
    return (
      <p className="text-sm text-zinc-500" aria-live="polite">
        Calculando…
      </p>
    );
  }

  if (!deal) {
    return null;
  }

  if (deal.status === "insufficient_data") {
    return (
      <div
        className="rounded-xl border border-zinc-200 bg-zinc-50 px-4 py-3 text-sm text-zinc-700"
        aria-live="polite"
      >
        <p className="font-medium">Sin datos de mercado</p>
        <p className="mt-1 text-zinc-600">{deal.disclaimer}</p>
      </div>
    );
  }

  const delta = deal.delta_ars ?? 0;
  const pct = deal.delta_pct ?? 0;

  if (deal.band === "below") {
    return (
      <div
        className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-emerald-900"
        aria-live="polite"
      >
        <p className="text-lg font-semibold">
          🟢 {formatARS(delta)} debajo del mercado
        </p>
        <p className="mt-1 text-sm text-emerald-800">
          {pct.toFixed(1).replace(".", ",")}% por debajo de la mediana
        </p>
        <p className="mt-2 text-xs text-emerald-700">{deal.disclaimer}</p>
      </div>
    );
  }

  if (deal.band === "above") {
    return (
      <div
        className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-red-900"
        aria-live="polite"
      >
        <p className="text-lg font-semibold">
          🔴 {formatARS(Math.abs(delta))} por encima del mercado
        </p>
        <p className="mt-1 text-sm text-red-800">
          {Math.abs(pct).toFixed(1).replace(".", ",")}% por encima de la mediana
        </p>
        <p className="mt-2 text-xs text-red-700">{deal.disclaimer}</p>
      </div>
    );
  }

  return (
    <div
      className="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-amber-950"
      aria-live="polite"
    >
      <p className="text-lg font-semibold">🟡 Precio dentro del rango de mercado</p>
      <p className="mt-1 text-sm text-amber-800">
        Entre el percentil 25 y 75 de las publicaciones
      </p>
      <p className="mt-2 text-xs text-amber-700">{deal.disclaimer}</p>
    </div>
  );
}

type DealCalculatorProps = {
  priceInput: string;
  onPriceChange: (value: string) => void;
  deal: DealAssessment | null;
  loading: boolean;
};

export function DealCalculator({
  priceInput,
  onPriceChange,
  deal,
  loading,
}: DealCalculatorProps) {
  return (
    <section className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm">
      <h2 className="text-sm font-semibold uppercase tracking-wide text-zinc-500">
        ¿Está barato?
      </h2>
      <label htmlFor="asking-price" className="mt-4 block text-sm text-zinc-600">
        Precio publicado (ARS)
      </label>
      <input
        id="asking-price"
        type="text"
        inputMode="numeric"
        value={priceInput}
        onChange={(event) => onPriceChange(event.target.value)}
        placeholder="Ej. 21500000"
        className="mt-2 w-full rounded-xl border border-zinc-300 px-4 py-3 text-base outline-none ring-emerald-600/30 focus:border-emerald-600 focus:ring-2"
      />
      <div className="mt-4">
        <DealBadge deal={deal} loading={loading} />
      </div>
    </section>
  );
}
