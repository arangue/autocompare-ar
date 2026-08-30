export type TrimSearchResult = {
  trim_id: number;
  trim_name: string;
  brand_name: string;
  model_name: string;
  generation_name: string;
  year_from: number | null;
  year_to: number | null;
};

export type VehicleSpec = {
  engine: string;
  displacement_cc: number | null;
  horsepower: number | null;
  transmission: string;
  fuel: string;
  doors: number | null;
  seats: number | null;
  consumption_city: number | null;
  consumption_highway: number | null;
};

export type VehicleFeature = {
  code: string;
  name: string;
  category: string;
  value: string;
};

export type TrimDetail = {
  trim_id: number;
  trim_name: string;
  brand_id: number;
  brand_name: string;
  model_id: number;
  model_name: string;
  generation_id: number;
  generation_name: string;
  year_from: number | null;
  year_to: number | null;
  requested_year: number | null;
  specs: VehicleSpec | null;
  features: VehicleFeature[];
};

export type Listing = {
  id: number;
  source: string;
  external_id: string;
  trim_id: number;
  year: number;
  km: number | null;
  price: number;
  currency: string;
  location: string | null;
  url: string | null;
  last_seen_at: string;
};

export type TrimMarketSummary = {
  trim_id: number;
  year: number;
  count: number;
  median: number | null;
  minimum: number | null;
  maximum: number | null;
  p25: number | null;
  p75: number | null;
  currency: string;
};

export type ReferencePrice = {
  id: number;
  trim_id: number;
  year: number;
  source: string;
  kind: string;
  price: number;
  currency: string;
  observed_at: string;
};

export type DealMarketSnapshot = {
  count: number;
  median: number | null;
  p25: number | null;
  p75: number | null;
};

export type DealAssessment = {
  trim_id: number;
  year: number;
  price: number;
  currency: string;
  status: "ok" | "insufficient_data";
  band: "below" | "near" | "above" | "insufficient_data";
  delta_ars: number | null;
  delta_pct: number | null;
  market: DealMarketSnapshot;
  disclaimer: string;
};

export type ApiErrorBody = {
  code: string;
  message: string;
};
