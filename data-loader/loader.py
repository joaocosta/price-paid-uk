#!/usr/bin/env python3
import sqlite3
import pandas as pd
import sys
import os
import time

"""
# UK Price Paid Data Ingestion & Sync

This script handles the initial 31M-row load of UK property sales data and monthly incremental updates.

## Data Attribution
- Source: HM Land Registry
- Licence: Open Government Licence v3.0 (https://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/)
- Documentation: https://www.gov.uk/guidance/about-the-price-paid-data

## Features
- Chunked Processing: Reads CSVs in chunks (default: 250k) to optimize memory.
- SQLite Write Tuning: Uses exclusive locking and memory journaling for 10x+ speed boost.
- Incremental Sync: Processes record statuses (A: Add, C: Change, D: Delete).
- Deferred Indexing: Drops indexes before bulk load and rebuilds them after to save time.

## Usage
Initial Load: python sync_prices.py pp-complete.csv
Monthly Update: python sync_prices.py pp-monthly.csv
"""

# Column headers based on GOV.UK documentation
HEADERS = [
    "Transaction_ID", "Price", "Date", "Postcode", "Property_Type",
    "Old_New", "Duration", "PAON", "SAON", "Street",
    "Locality", "Town_City", "District", "County", "PPD_Category", "Status"
]

DB_NAME = "prices.db"

def get_connection():
    conn = sqlite3.connect(DB_NAME)
    # Speed PRAGMAs for bulk loading
    conn.execute("PRAGMA synchronous = OFF;")
    conn.execute("PRAGMA journal_mode = MEMORY;")
    conn.execute("PRAGMA locking_mode = EXCLUSIVE;")
    conn.execute("PRAGMA cache_size = -2000000;") # 2GB cache
    return conn

def init_db(conn):
    """Initializes the SQLite database and creates the necessary schema."""
    cursor = conn.cursor()
    cursor.execute("""
    CREATE TABLE IF NOT EXISTS land_registry_prices (
        Transaction_ID TEXT PRIMARY KEY,
        Price INTEGER,
        Date TEXT,
        Postcode TEXT,
        Property_Type TEXT,
        Old_New TEXT,
        Duration TEXT,
        PAON TEXT,
        SAON TEXT,
        Street TEXT,
        Locality TEXT,
        Town_City TEXT,
        District TEXT,
        County TEXT,
        PPD_Category TEXT
    )
    """)
    conn.commit()

def drop_indexes(conn):
    print("Dropping indexes for bulk load speed...")
    cursor = conn.cursor()
    cursor.execute("DROP INDEX IF EXISTS idx_postcode;")
    cursor.execute("DROP INDEX IF EXISTS idx_date;")
    cursor.execute("DROP INDEX IF EXISTS idx_town_city;")
    cursor.execute("DROP INDEX IF EXISTS idx_district;")
    conn.commit()

def normalize_cities(conn):
    print("Normalizing city names (e.g., ST. ALBANS -> ST ALBANS)...")
    cursor = conn.cursor()
    city_mappings = [
        "UPDATE land_registry_prices SET Town_City = 'BURY ST EDMUNDS' WHERE Town_City = 'BURY ST. EDMUNDS';",
        "UPDATE land_registry_prices SET Town_City = 'CHALFONT ST GILES' WHERE Town_City = 'CHALFONT ST. GILES';",
        "UPDATE land_registry_prices SET Town_City = 'HINTON ST GEORGE' WHERE Town_City = 'HINTON ST. GEORGE';",
        "UPDATE land_registry_prices SET Town_City = 'LYTHAM ST ANNES' WHERE Town_City = 'LYTHAM ST. ANNES';",
        "UPDATE land_registry_prices SET Town_City = 'OTTERY ST MARY' WHERE Town_City = 'OTTERY ST. MARY';",
        "UPDATE land_registry_prices SET Town_City = 'ST AGNES' WHERE Town_City = 'ST. AGNES';",
        "UPDATE land_registry_prices SET Town_City = 'ST ALBANS' WHERE Town_City = 'ST. ALBANS';",
        "UPDATE land_registry_prices SET Town_City = 'ST ASAPH' WHERE Town_City = 'ST. ASAPH';",
        "UPDATE land_registry_prices SET Town_City = 'ST AUSTELL' WHERE Town_City = 'ST. AUSTELL';",
        "UPDATE land_registry_prices SET Town_City = 'ST BEES' WHERE Town_City = 'ST. BEES';",
        "UPDATE land_registry_prices SET Town_City = 'ST COLUMB' WHERE Town_City = 'ST. COLUMB';",
        "UPDATE land_registry_prices SET Town_City = 'ST HELENS' WHERE Town_City = 'ST. HELENS';",
        "UPDATE land_registry_prices SET Town_City = 'ST IVES' WHERE Town_City = 'ST. IVES';",
        "UPDATE land_registry_prices SET Town_City = 'ST LEONARDS-ON-SEA' WHERE Town_City = 'ST. LEONARDS-ON-SEA';",
        "UPDATE land_registry_prices SET Town_City = 'ST NEOTS' WHERE Town_City = 'ST. NEOTS';"
    ]
    for sql in city_mappings:
        cursor.execute(sql)
    conn.commit()

def create_indexes(conn):
    print("Creating indexes (this may take a few minutes for 31M rows)...")
    cursor = conn.cursor()
    cursor.execute("CREATE INDEX IF NOT EXISTS idx_postcode ON land_registry_prices(Postcode);")
    cursor.execute("CREATE INDEX IF NOT EXISTS idx_date ON land_registry_prices(Date);")
    cursor.execute("CREATE INDEX IF NOT EXISTS idx_town_city ON land_registry_prices(Town_City);")
    cursor.execute("CREATE INDEX IF NOT EXISTS idx_district ON land_registry_prices(District);")
    conn.commit()

def sync_data(file_path):
    """Processes the CSV in chunks and syncs it with the SQLite database."""
    if not os.path.exists(file_path):
        print(f"Error: File {file_path} not found.")
        return

    conn = get_connection()
    init_db(conn)

    drop_indexes(conn)

    cursor = conn.cursor()
    csv_chunk_size = 250000
    start_time = time.time()
    total_processed = 0

    print(f"Starting optimized sync of {file_path}...")
 
    # SQL for Upsert
    sql = """INSERT OR REPLACE INTO land_registry_prices
             (Transaction_ID, Price, Date, Postcode, Property_Type,
              Old_New, Duration, PAON, SAON, Street,
              Locality, Town_City, District, County, PPD_Category)
             VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"""

    for chunk in pd.read_csv(file_path, names=HEADERS, chunksize=csv_chunk_size, low_memory=False):
        # 1. Handle Deletions (Status 'D')
        deletions = chunk[chunk['Status'] == 'D']['Transaction_ID'].tolist()
        if deletions:
            cursor.executemany("DELETE FROM land_registry_prices WHERE Transaction_ID = ?",
                             [(tid,) for tid in deletions])

        # 2. Prepare Additions/Changes (Status 'A' or 'C')
        upsert_data = chunk[chunk['Status'] != 'D'].copy()
 
        # Fast Date conversion (slice first 16 chars: YYYY-MM-DD HH:MM)
        upsert_data['Date'] = upsert_data['Date'].str.slice(0, 16)

        # Replace NaN with None
        upsert_data = upsert_data.where(pd.notnull(upsert_data), None)

        # Columns to insert (everything except Status)
        cols = [c for c in HEADERS if c != "Status"]

        # Batch execute
        cursor.executemany(sql, upsert_data[cols].values.tolist())

        total_processed += len(chunk)
        elapsed = time.time() - start_time
        rows_per_sec = total_processed / elapsed
        print(f"Processed {total_processed:,} rows... ({rows_per_sec:,.0f} rows/sec)")

        # Commit every chunk
        conn.commit()

    normalize_cities(conn)
    create_indexes(conn)

    print(f"Sync complete! Total rows processed: {total_processed:,}")
    print(f"Total time: {time.time() - start_time:.2f}s")

    print("Finalizing (VACUUM & ANALYZE)...")
    conn.execute("VACUUM;")
    conn.execute("ANALYZE;")
    conn.close()

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: sync_prices.py <path_to_csv>")
    else:
        sync_data(sys.argv[1])
