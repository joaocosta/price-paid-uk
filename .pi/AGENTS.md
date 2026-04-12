# UK Price Paid Data Project: Master Index

This project handles 31M+ records of UK property sales data, providing a high-performance web dashboard with monthly incremental updates.

## 📄 Data Source & Attribution
This project uses **HM Land Registry Price Paid Data** which is licensed under the **Open Government Licence v3.0**.

- **Dataset Collection**: [HM Land Registry: Price Paid Data](https://www.gov.uk/government/collections/price-paid-data)
- **Column Explanations**: [Guidance on Price Paid Data Fields](https://www.gov.uk/guidance/about-the-price-paid-data#explanations-of-column-headers-in-the-ppd)

---

## 🗺️ Project Navigation
- [Data Ingestion & Sync](sync_prices.py): In-code documentation for 31M-row load and monthly updates.
- [Architecture & Tech Stack](.pi/docs/ARCHITECTURE.md): Core design and backend/frontend strategy.
- [API & Backend](.pi/docs/API.md): FastAPI service endpoints and SQL optimization.
- **Frontend & Visualization**[.pi/docs/FRONTEND.md]: Dashboard UI, Plotly charts, and mapping (Includes House Number and Road Name).

## 🚀 Quick Commands
- **Initial Load**: `python sync_prices.py pp-complete.csv`
- **Monthly Update**: `python sync_prices.py pp-monthly.csv`
- **Start Web Service**: `uvicorn app.main:app --reload`

## 🛠️ Current Status
- [x] Initial Project Setup
- [ ] Ingestion Script (Incremental Sync)
- [ ] SQLite Schema Optimization
- [ ] FastAPI Backend
- [ ] Interactive Dashboard

---

### 📜 Agent Protocol
1.  **Read the Master Index**: Before starting work, read this **Master Index** and then load the specific `.pi/docs/*.md` file related to your current task.
2.  **Follow the License**: Always ensure the frontend and all public-facing outputs explicitly attribute the data source to **HM Land Registry** and link to the **Open Government Licence v3.0**.
3.  **Context Management**: **Do not** attempt to read all documentation at once to save context.
