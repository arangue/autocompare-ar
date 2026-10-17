#!/usr/bin/env python3
"""Collect Mercado Libre asking prices for one search path.

The path is a brand (`ford`) or brand-model (`fiat-siena`). The CSV is a
price feed. It does not assign trims. The worker matches each title to the
carsbase catalog.
"""

import argparse
import random
import sys
import time
import unicodedata

import pandas as pd
import requests
from bs4 import BeautifulSoup

# Mercado Libre serves about 48 listings per page and 404s past page 42.
DEFAULT_PAGES = 42
SKIP_EXIT = 2

HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
        "AppleWebKit/537.36 (KHTML, like Gecko) "
        "Chrome/122.0.0.0 Safari/537.36"
    ),
    "Accept-Language": "es-AR,es;q=0.9",
    "Accept": (
        "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"
    ),
    "Referer": "https://www.mercadolibre.com.ar/",
}


def slug(text: str) -> str:
    text = unicodedata.normalize("NFD", text)
    text = "".join(c for c in text if unicodedata.category(c) != "Mn")
    parts: list[str] = []
    hyphen = False
    for c in text.casefold():
        if c.isalnum():
            parts.append(c)
            hyphen = False
        elif not hyphen:
            parts.append("-")
            hyphen = True
    return "".join(parts).strip("-")


def page_url(path: str, page: int) -> str:
    base = f"https://autos.mercadolibre.com.ar/{path}"
    if page == 0:
        return base
    return f"{base}_Desde_{page * 48 + 1}"


def text_or_blank(elem) -> str:
    if elem is None:
        return ""
    return elem.get_text(strip=True)


def listing_currency(price_root) -> str:
    """USD if the card says dólares / US$; otherwise ARS."""
    if price_root is None:
        return "ARS"
    amount = price_root.select_one(".andes-money-amount")
    label = ""
    if amount is not None:
        label = (amount.get("aria-label") or "").casefold()
    if "dólar" in label or "dolar" in label:
        return "USD"
    if "peso" in label:
        return "ARS"
    symbol = text_or_blank(
        price_root.select_one(".andes-money-amount__currency-symbol")
    ).upper().replace(" ", "")
    if symbol in {"US$", "U$S", "USD", "U$D"}:
        return "USD"
    return "ARS"


def scrape(path: str, pages: int) -> list[dict] | None:
    rows = []
    for page in range(pages):
        url = page_url(path, page)
        print(f"Scraping page {page + 1}: {url}")
        response = requests.get(url, headers=HEADERS, timeout=30)
        if response.status_code == 404:
            print(f"404 skip: {url}")
            if page == 0:
                return None
            break
        if response.status_code != 200:
            print(f"Failed to fetch page. Status code: {response.status_code}")
            if page == 0:
                return None
            break

        soup = BeautifulSoup(response.text, "html.parser")
        items = soup.select("li.ui-search-layout__item")
        if not items:
            print("No listings found. Mercado Libre markup may have changed.")
            if page == 0:
                return None
            break

        for item in items:
            title_elem = item.select_one("a.poly-component__title")
            price_root = item.select_one(".poly-component__price")
            price_elem = None
            if price_root is not None:
                price_elem = price_root.select_one(".andes-money-amount__fraction")
            attributes = [
                li.get_text(strip=True)
                for li in item.select("ul.poly-attributes_list > li")
                if li.get_text(strip=True)
            ]
            location_elem = item.select_one("span.poly-component__location")
            rows.append(
                {
                    "Title": text_or_blank(title_elem),
                    "Price": text_or_blank(price_elem),
                    "Currency": listing_currency(price_root),
                    "Year": attributes[0] if attributes else "",
                    "Kilometers": attributes[1] if len(attributes) > 1 else "",
                    "Location": text_or_blank(location_elem),
                    "URL": (
                        title_elem["href"]
                        if title_elem is not None and title_elem.has_attr("href")
                        else ""
                    ),
                }
            )

        print(f"  listings so far: {len(rows)}")
        if page + 1 < pages:
            time.sleep(random.uniform(2, 4))
    return rows


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "path",
        help="search path on autos.mercadolibre.com.ar, for example ford or fiat-siena",
    )
    parser.add_argument(
        "--pages",
        type=int,
        default=DEFAULT_PAGES,
        help=f"pages to scrape, 48 listings each (default {DEFAULT_PAGES})",
    )
    parser.add_argument(
        "--output",
        default="",
        help="CSV path (default mercadolibre_<path>_cars.csv)",
    )
    args = parser.parse_args()
    if args.pages < 1:
        parser.error("--pages must be at least 1")

    path = slug(args.path) if args.path else ""
    if not path:
        parser.error("path is empty")

    rows = scrape(path, args.pages)
    if not rows:
        print(f"skip {path}: no listings")
        sys.exit(SKIP_EXIT)

    output = args.output or f"mercadolibre_{path}_cars.csv"
    df = pd.DataFrame(
        rows,
        columns=["Title", "Price", "Currency", "Year", "Kilometers", "Location", "URL"],
    )
    df.to_csv(output, index=False)
    print(f"Successfully scraped {len(df)} listings into {output}.")


if __name__ == "__main__":
    main()
