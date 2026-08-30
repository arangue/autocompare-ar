import type {
  ApiErrorBody,
  CompareResult,
  DealAssessment,
  Listing,
  ReferencePrice,
  TrimDetail,
  TrimMarketSummary,
  TrimSearchResult,
} from "./types";

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

function apiBase(): string {
  if (typeof window !== "undefined") {
    return "";
  }
  return process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
}

async function fetchJSON<T>(path: string): Promise<T> {
  const response = await fetch(`${apiBase()}${path}`, {
    headers: { Accept: "application/json" },
  });

  if (!response.ok) {
    let body: ApiErrorBody | null = null;
    try {
      body = (await response.json()) as ApiErrorBody;
    } catch {
      body = null;
    }
    throw new ApiError(
      response.status,
      body?.code ?? "HTTP_ERROR",
      body?.message ?? response.statusText,
    );
  }

  return response.json() as Promise<T>;
}

export function searchTrims(q: string, limit = 20): Promise<TrimSearchResult[]> {
  const params = new URLSearchParams({ q, limit: String(limit) });
  return fetchJSON<TrimSearchResult[]>(`/api/v1/search/trims?${params}`);
}

export function getTrim(trimId: number, year?: number): Promise<TrimDetail> {
  const params = year ? `?year=${year}` : "";
  return fetchJSON<TrimDetail>(`/api/v1/trims/${trimId}${params}`);
}

export function getMarket(trimId: number, year: number): Promise<TrimMarketSummary> {
  return fetchJSON<TrimMarketSummary>(
    `/api/v1/trims/${trimId}/market?year=${year}`,
  );
}

export function getDeal(
  trimId: number,
  year: number,
  price: number,
): Promise<DealAssessment> {
  const params = new URLSearchParams({
    year: String(year),
    price: String(price),
  });
  return fetchJSON<DealAssessment>(`/api/v1/trims/${trimId}/deal?${params}`);
}

export function getListings(
  trimId: number,
  year: number,
  limit = 50,
): Promise<Listing[]> {
  const params = new URLSearchParams({
    year: String(year),
    limit: String(limit),
  });
  return fetchJSON<Listing[]>(`/api/v1/trims/${trimId}/listings?${params}`);
}

export function getReferences(
  trimId: number,
  year: number,
): Promise<ReferencePrice[]> {
  return fetchJSON<ReferencePrice[]>(
    `/api/v1/trims/${trimId}/references?year=${year}`,
  );
}

export function getCompare(trimIds: number[], year: number): Promise<CompareResult> {
  const params = new URLSearchParams({
    trim_ids: trimIds.join(","),
    year: String(year),
  });
  return fetchJSON<CompareResult>(`/api/v1/compare?${params}`);
}
