import Link from "next/link";

import { HomeSearch } from "@/components/HomeSearch";

export default function Home() {
  return (
    <div className="flex flex-1 flex-col">
      <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col px-6 py-16 sm:py-24">
        <div className="mb-10 max-w-2xl">
          <p className="mb-3 text-sm font-semibold uppercase tracking-[0.2em] text-emerald-700">
            Argentina
          </p>
          <h1 className="text-4xl font-semibold tracking-tight text-zinc-900 sm:text-5xl">
            AutoCompare
          </h1>
          <p className="mt-4 text-lg leading-8 text-zinc-600">
            ¿Está barato? Compará el precio publicado con el mercado real y las
            referencias oficiales antes de comprar.
          </p>
          <p className="mt-4 text-sm text-zinc-500">
            Esto no es un clasificado. Son ejemplos del catálogo de demostración.
          </p>
          {/* seed ids from migrations/002 */}
          <ul className="mt-6 space-y-2 text-emerald-800">
            <li>
              <Link href="/trims/1?year=2019">Corolla XEi 2.0 CVT 2019</Link>
            </li>
            <li>
              <Link href="/trims/2?year=2019">Corolla XLi 1.8 CVT 2019</Link>
            </li>
            <li>
              <Link href="/trims/4?year=2019">Golf Comfortline 2019</Link>
            </li>
          </ul>
        </div>

        <HomeSearch />
      </main>
    </div>
  );
}
