# Frontend & Visualization

The frontend is a responsive dashboard for exploring 31M+ records of property data.

## 🧱 UI Layout
- **Global Filters**: Top-level filters for Date Range, Property Type, and Location (Town/District).
- **KPI Summary**: Real-time updating cards for Total Volume, Average Price, and Median Price.
- **Charts**: Interactive Plotly.js charts for Price Distribution and Regional Comparisons.
- **Data Table**: A server-side paginated table for browsing individual transactions.

## 🏠 Transaction Details (Data Table)
The data table includes the following columns:
1.  **Price**: Bold formatted currency.
2.  **Date**: Sale completion date.
3.  **Address**: Combined from PAON (House Number/Name), Street (Road Name), and Postcode.
    - *Note*: Road Name and House Number are displayed for context but are **not** currently filterable to maintain query performance.
4.  **Property Type**: Detached, Semi, Flat, etc.
5.  **Location**: Town/City and District.

## ⚡ Interaction Pattern
- **Filter Change**: Triggers an HTMX request or Fetch call to the API.
- **Update**: KPI cards and Charts re-render based on the new filtered JSON data.
- **Pagination**: Table requests the next 50 records from the `/search` endpoint.
