import Link from "next/link";

export function SiteNav() {
  return (
    <header className="border-b border-zinc-200 bg-white">
      <div className="mx-auto flex w-full max-w-5xl items-center justify-between px-6 py-3">
        <Link href="/" className="text-sm font-semibold tracking-tight text-zinc-900">
          AutoCompare
        </Link>
        <nav className="flex gap-4 text-sm">
          <Link href="/" className="text-zinc-600 hover:text-zinc-900">
            Buscar
          </Link>
          <Link href="/modelos" className="text-zinc-600 hover:text-zinc-900">
            Modelos
          </Link>
        </nav>
      </div>
    </header>
  );
}
