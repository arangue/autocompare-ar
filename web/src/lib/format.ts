export function formatARS(amount: number): string {
  return formatMoney(amount, "ARS");
}

export function formatMoney(amount: number, currency: string): string {
  return new Intl.NumberFormat("es-AR", {
    style: "currency",
    currency: currency === "USD" ? "USD" : "ARS",
    maximumFractionDigits: 0,
  }).format(amount);
}

export function formatARSCompact(amount: number): string {
  return formatMoneyCompact(amount, "ARS");
}

export function formatMoneyCompact(amount: number, currency: string): string {
  const code = currency === "USD" ? "USD" : "ARS";
  if (code === "ARS" && amount >= 1_000_000) {
    const millions = amount / 1_000_000;
    const formatted =
      millions % 1 === 0
        ? millions.toFixed(0)
        : millions.toFixed(1).replace(".", ",");
    return `$${formatted}M`;
  }
  return formatMoney(amount, code);
}

export function formatKm(km: number): string {
  return `${new Intl.NumberFormat("es-AR").format(km)} km`;
}
