export function formatARS(amount: number): string {
  return new Intl.NumberFormat("es-AR", {
    style: "currency",
    currency: "ARS",
    maximumFractionDigits: 0,
  }).format(amount);
}

export function formatARSCompact(amount: number): string {
  if (amount >= 1_000_000) {
    const millions = amount / 1_000_000;
    const formatted =
      millions % 1 === 0
        ? millions.toFixed(0)
        : millions.toFixed(1).replace(".", ",");
    return `$${formatted}M`;
  }
  return formatARS(amount);
}

export function formatKm(km: number): string {
  return `${new Intl.NumberFormat("es-AR").format(km)} km`;
}
