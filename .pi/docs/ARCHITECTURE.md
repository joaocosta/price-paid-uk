# Architecture & Tech Stack

This project is built for high-performance reading of 31M+ records on a standard machine using SQLite and Python.

## 🗄️ Database: SQLite
- **Why**: SQLite is excellent for read-heavy analytical datasets that fit on a single disk.
- **Mode**: Use `WAL` (Write-Ahead Logging) mode for concurrent reads/writes and `PRAGMA journal_mode=WAL;`.
- **Primary Key**: `Transaction_ID` (UUID from the dataset).
- **Indexing Strategy**:
  - `Postcode`: For location-based lookups.
  - `Town_City`: For filtering by city.
  - `Date`: For time-series analysis.
  - `Price`: For price range filtering.

## 🐹 Backend: Go & Gin
- **Performance**: High-performance Go API for concurrent querying of SQLite.
- **Gin Gonic**: Used as the web framework for speed and simplicity.
- **Standard SQL**: Using Go's standard `database/sql` library with the `go-sqlite3` driver.
- **Pagination**: All endpoints return paginated results by default.
- **Embedding**: Potential to embed the `web/` directory into the Go binary for single-file deployment.

## 📁 Repository Structure
- `backend/`: Go source code and dependencies.
- `data-loader/`: Python-based ingestion logic (optimized with Pandas).
- `web/`: Vanilla HTML/JS frontend.
- `prices.db`: SQLite database at the root.

## 🎨 Frontend: Tailwind CSS & Plotly.js
- **UI Framework**: Tailwind CSS for a modern, responsive layout.
- **Data Attribution (MANDATORY)**: The UI must comply with the [Open Government Licence v3.0](https://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/).
    - **Acknowledgement**: Include the statement: *"Contains public sector information licensed under the Open Government Licence v3.0."*
    - **Link**: Provide a functional link to the license.
    - **Non-Endorsement**: Do not use official GOV.UK or HM Land Registry logos or suggest official endorsement.
    - **Disclaimer**: State that the data is provided "as is" and the licensor is not liable for any errors or omissions.
- **Charts**: Plotly.js for interactive, zoomable data visualizations.
- **HTMX**: Use HTMX for dynamic content loading (e.g., updating the data table without a full page refresh).
- **Mapping**: Leaflet.js or Mapbox for geospatial heatmaps.
