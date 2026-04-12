package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

const (
	Attribution = "Contains public sector information licensed under the Open Government Licence v3.0."
	LicenseURL  = "https://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
	DBPath      = "../prices.db"
	SlowQuery   = 3 * time.Second
)

func logSlowQuery(query string, args []interface{}, duration time.Duration) {
	if duration >= SlowQuery {
		sql := query
		for _, arg := range args {
			var val string
			switch v := arg.(type) {
			case string:
				val = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
			default:
				val = fmt.Sprintf("%v", v)
			}
			sql = strings.Replace(sql, "?", val, 1)
		}
		log.Printf("[SLOW QUERY] %v\n%s;\n", duration, sql)
	}
}

// PriceRecord represents the database row for a price paid transaction
type PriceRecord struct {
	TransactionID string  `json:"Transaction_ID"`
	Price         int     `json:"Price"`
	Date          string  `json:"Date"`
	Postcode      *string `json:"Postcode"`
	PropertyType  string  `json:"Property_Type"`
	TownCity      string  `json:"Town_City"`
	District      string  `json:"District"`
	PAON          *string `json:"PAON"`
	SAON          *string `json:"SAON"`
	Street        *string `json:"Street"`
}

var db *sql.DB

func getSQLiteCompileOptions(db *sql.DB) ([]string, error) {
	rows, err := db.Query("PRAGMA compile_options")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []string
	for rows.Next() {
		var option string
		if err := rows.Scan(&option); err != nil {
			return nil, err
		}
		options = append(options, option)
	}

	sort.Strings(options)
	return options, rows.Err()
}

func summarizeSQLiteCompileOptions(options []string) []string {
	var summary []string
	for _, option := range options {
		if strings.HasPrefix(option, "ENABLE_") {
			summary = append(summary, option)
			continue
		}

		switch {
		case option == "DIRECT_OVERFLOW_READ":
			summary = append(summary, option)
		case strings.HasPrefix(option, "THREADSAFE="):
			summary = append(summary, option)
		case strings.HasPrefix(option, "TEMP_STORE="):
			summary = append(summary, option)
		case strings.HasPrefix(option, "DEFAULT_SYNCHRONOUS="):
			summary = append(summary, option)
		case strings.HasPrefix(option, "DEFAULT_WAL_SYNCHRONOUS="):
			summary = append(summary, option)
		}
	}

	return summary
}

func hasSQLiteCompileOption(options []string, target string) bool {
	for _, option := range options {
		if option == target {
			return true
		}
	}
	return false
}

func main() {
	var err error
	// Open SQLite in Read-Only mode for safety
	db, err = sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", DBPath))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var sqliteVersion string
	if err := db.QueryRow("SELECT sqlite_version()").Scan(&sqliteVersion); err != nil {
		log.Fatal(err)
	}
	log.Printf("Using SQLite library version %s", sqliteVersion)

	compileOptions, err := getSQLiteCompileOptions(db)
	if err != nil {
		log.Fatal(err)
	}
	if !hasSQLiteCompileOption(compileOptions, "ENABLE_STAT4") {
		log.Fatal("SQLite library is missing ENABLE_STAT4; refusing to start because postcode query planning depends on it")
	}
	log.Printf("SQLite compile options: %s", strings.Join(summarizeSQLiteCompileOptions(compileOptions), ", "))

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":     "UK Price Paid Data API (Go) is active",
			"attribution": Attribution,
			"license":     LicenseURL,
		})
	})

	r.GET("/stats", getStats)
	r.GET("/search", searchPrices)
	r.GET("/cities", getCities)
	r.GET("/history", getHistory)

	// Run server
	fmt.Println("Server starting on :8000...")
	r.Run(":8000")
}

func getStats(c *gin.Context) {
	town := c.Query("town")
	postcode := c.Query("postcode")
	year := c.Query("year")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")

	query := "SELECT count(*), avg(Price) FROM land_registry_prices WHERE 1=1"
	var args []interface{}

	if town != "" {
		query += " AND Town_City = ?"
		args = append(args, town)
	}
	if postcode != "" {
        query += " AND Postcode LIKE ?"
        args = append(args, postcode+"%")
	}
	if year != "" {
		query += " AND Date LIKE ?"
		args = append(args, year+"%")
	}
	if startDate != "" {
		query += " AND Date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND Date <= ?"
		args = append(args, endDate)
	}
	if minPrice != "" {
		query += " AND Price >= ?"
		args = append(args, minPrice)
	}
	if maxPrice != "" {
		query += " AND Price <= ?"
		args = append(args, maxPrice)
	}

	var count int
	var avg sql.NullFloat64
	start := time.Now()
	err := db.QueryRow(query, args...).Scan(&count, &avg)
	logSlowQuery(query, args, time.Since(start))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_count":  count,
		"avg_price":    avg.Float64,
		"median_price": 0.0,
		"attribution":  Attribution,
	})
}

func searchPrices(c *gin.Context) {
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")
	town := c.Query("town")
	postcode := c.Query("postcode")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := "SELECT Transaction_ID, Price, Date, Postcode, Property_Type, Town_City, District, PAON, SAON, Street FROM land_registry_prices WHERE 1=1"
	var args []interface{}

	if town != "" {
		query += " AND Town_City = ?"
		args = append(args, town)
	}
	if postcode != "" {
        query += " AND Postcode LIKE ?"
        args = append(args, postcode+"%")
	}
	if minPrice != "" {
		query += " AND Price >= ?"
		args = append(args, minPrice)
	}
	if maxPrice != "" {
		query += " AND Price <= ?"
		args = append(args, maxPrice)
	}
	if startDate != "" {
		query += " AND Date >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND Date <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY Date DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	start := time.Now()
	rows, err := db.Query(query, args...)
	logSlowQuery(query, args, time.Since(start))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var records []PriceRecord = []PriceRecord{} // Initialize as empty slice not nil
	for rows.Next() {
		var r PriceRecord
		err := rows.Scan(&r.TransactionID, &r.Price, &r.Date, &r.Postcode, &r.PropertyType, &r.TownCity, &r.District, &r.PAON, &r.SAON, &r.Street)
		if err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		records = append(records, r)
	}

	c.JSON(http.StatusOK, records)
}

func getCities(c *gin.Context) {
	rows, err := db.Query("SELECT DISTINCT Town_City FROM land_registry_prices ORDER BY Town_City LIMIT 50000")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	cities := []string{}
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err == nil {
			cities = append(cities, city)
		}
	}
	c.JSON(http.StatusOK, cities)
}

func getHistory(c *gin.Context) {
	postcode := c.Query("postcode")
	paon := c.Query("paon")
	saon := c.Query("saon")
	street := c.Query("street")

	if postcode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "postcode is required"})
		return
	}

	query := `
		SELECT Transaction_ID, Price, Date, Postcode, Property_Type, Town_City, District, PAON, SAON, Street 
		FROM land_registry_prices 
		WHERE Postcode = ?`
	var args []interface{}
	args = append(args, postcode)

	if paon != "" {
		query += " AND PAON = ?"
		args = append(args, paon)
	} else {
		query += " AND (PAON IS NULL OR PAON = '')"
	}

	if saon != "" {
		query += " AND SAON = ?"
		args = append(args, saon)
	} else {
		query += " AND (SAON IS NULL OR SAON = '')"
	}

	if street != "" {
		query += " AND Street = ?"
		args = append(args, street)
	} else {
		query += " AND (Street IS NULL OR Street = '')"
	}

	query += " ORDER BY Date DESC"

	start := time.Now()
	rows, err := db.Query(query, args...)
	logSlowQuery(query, args, time.Since(start))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var records []PriceRecord = []PriceRecord{}
	for rows.Next() {
		var r PriceRecord
		err := rows.Scan(&r.TransactionID, &r.Price, &r.Date, &r.Postcode, &r.PropertyType, &r.TownCity, &r.District, &r.PAON, &r.SAON, &r.Street)
		if err != nil {
			continue
		}
		records = append(records, r)
	}

	c.JSON(http.StatusOK, records)
}
