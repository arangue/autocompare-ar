import type { TrimDetail } from "./types";

export function isYearInRange(
  year: number,
  yearFrom: number | null,
  yearTo: number | null,
): boolean {
  if (yearFrom != null && year < yearFrom) {
    return false;
  }
  if (yearTo != null && year > yearTo) {
    return false;
  }
  return yearFrom != null || yearTo != null;
}

export function defaultYear(trim: TrimDetail, queryYear?: number): number {
  if (queryYear != null && isYearInRange(queryYear, trim.year_from, trim.year_to)) {
    return queryYear;
  }
  if (trim.year_from != null && trim.year_to != null) {
    return Math.floor((trim.year_from + trim.year_to) / 2);
  }
  if (trim.year_from != null) {
    return trim.year_from;
  }
  return new Date().getFullYear();
}

export function yearOptions(
  yearFrom: number | null,
  yearTo: number | null,
): number[] {
  if (yearFrom == null) {
    return [];
  }
  const end = yearTo ?? new Date().getFullYear();
  const years: number[] = [];
  for (let year = end; year >= yearFrom; year -= 1) {
    years.push(year);
  }
  return years;
}
