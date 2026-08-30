import { Suspense } from "react";

import { VehicleSheet } from "@/components/VehicleSheet";

export default function TrimPage() {
  return (
    <Suspense
      fallback={
        <main className="mx-auto w-full max-w-3xl px-6 py-16">
          <p className="text-zinc-600">Cargando versión…</p>
        </main>
      }
    >
      <VehicleSheet />
    </Suspense>
  );
}
