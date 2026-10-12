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
        </div>

        <HomeSearch />
      </main>
    </div>
  );
}
