# UK Property Price Paid Explorer

A high-performance explorer for the HM Land Registry Price Paid Data, covering over 31 million property transactions in England and Wales since 1995.

## Features

- **Data Ingestion**: Python-based loader with SQLite optimizations (WAL mode, memory journaling, deferred indexing).
- **Go Backend**: Gin-powered API providing lightning-fast queries over millions of rows.
- **Modern Web Interface**: Responsive dashboard built with Tailwind CSS, Alpine.js, and Plotly for data visualization.
- **Address History**: Deep dive into specific properties to see their historical sales and price appreciation.

## Project Structure

- `data-loader/`: Python scripts to download and ingest CSV data into SQLite.
- `backend/`: Go API server.
- `web/`: Frontend dashboard.

## Attribution

Contains public sector information licensed under the [Open Government Licence v3.0](https://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/).

Data Source: [HM Land Registry Price Paid Data](https://www.gov.uk/guidance/about-the-price-paid-data).
