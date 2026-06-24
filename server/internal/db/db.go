package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath, migrationsPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode and foreign key constraints
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run migrations
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"sqlite3", driver)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully.")

	// Run data cleaning and uniformization on startup for existing databases
	log.Println("Running store name cleanup and uniformization...")
	if err := CleanExistingStores(db); err != nil {
		log.Printf("Warning: failed to clean existing stores: %v", err)
	}
	if err := CleanExistingVisits(db); err != nil {
		log.Printf("Warning: failed to clean existing visits: %v", err)
	}
	if err := CleanExistingAgents(db); err != nil {
		log.Printf("Warning: failed to clean existing agents: %v", err)
	}

	return db, nil
}

func CleanStoreName(name string) string {
	// 1. Remove all double quotes
	name = strings.ReplaceAll(name, "\"", "")
	// 2. Trim spaces
	name = strings.TrimSpace(name)
	// 3. Remove case-insensitive " di " ownership indications
	lowerName := strings.ToLower(name)
	if idx := strings.Index(lowerName, " di "); idx != -1 {
		name = name[:idx]
	}
	// 4. Uniform to uppercase
	return strings.ToUpper(strings.TrimSpace(name))
}

func GetOfficialAgentName(name string) string {
	name = strings.ToUpper(strings.TrimSpace(name))
	name = strings.Join(strings.Fields(name), " ")

	mappings := map[string]string{
		"CUNDARI S.A.S DI CUNDARI GIANFRANCO": "CUNDARI GIANFRANCO",
		"DI MENTO DARIO":                      "DARIO DI MENTO",
		"ERNESTO GALLIANI SAS":                "GALLIANI ERNESTO",
		"LDS RAPPRESENTANZE SRL SEMPLIFICATA": "LUIGI DE SENA",
		"MAROCCHI SAS DI ANDREA MAROCCHI & C.": "MAROCCHI ANDREA",
		"PERU GIOVANNI MARTINO":               "PERU GIAN MARTINO",
		"SG RAPPRESENTANZE S.R.L.S.":          "COSSU SAVERIO",
		"ERRE VI - SOCIETA' A RESPONSABILITA' LIMITATA SEMPLIFICATA": "GIULIANINI GIANCARLO",
		"FAZZARI GINO": "NOVELLI DAVIDE",
	}

	if official, ok := mappings[name]; ok {
		return official
	}
	return name
}

func CleanExistingStores(db *sql.DB) error {
	rows, err := db.Query("SELECT id, name, province, province_name, address, region, region_code, city, agent_name FROM stores")
	if err != nil {
		return fmt.Errorf("failed to select stores for cleanup: %w", err)
	}
	defer rows.Close()

	type Store struct {
		id           int
		name         string
		province     string
		provinceName string
		address      string
		region       string
		regionCode   string
		city         string
		agentName    string
	}

	var list []Store
	for rows.Next() {
		var s Store
		if err := rows.Scan(&s.id, &s.name, &s.province, &s.provinceName, &s.address, &s.region, &s.regionCode, &s.city, &s.agentName); err != nil {
			return fmt.Errorf("failed to scan store row for cleanup: %w", err)
		}
		list = append(list, s)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start cleanup transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE stores 
		SET name = ?, province = ?, province_name = ?, address = ?, region = ?, region_code = ?, city = ?, agent_name = ?
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	for _, s := range list {
		cName := CleanStoreName(s.name)
		cProv := strings.ToUpper(strings.TrimSpace(s.province))
		cProvName := strings.ToUpper(strings.TrimSpace(s.provinceName))
		cAddr := strings.ToUpper(strings.TrimSpace(s.address))
		cReg := strings.ToUpper(strings.TrimSpace(s.region))
		cRegCode := strings.ToUpper(strings.TrimSpace(s.regionCode))
		cCity := strings.ToUpper(strings.TrimSpace(s.city))
		cAgent := GetOfficialAgentName(s.agentName)

		_, err = stmt.Exec(cName, cProv, cProvName, cAddr, cReg, cRegCode, cCity, cAgent, s.id)
		if err != nil {
			return fmt.Errorf("failed to update store %d: %w", s.id, err)
		}
	}

	return tx.Commit()
}

func CleanExistingVisits(db *sql.DB) error {
	rows, err := db.Query("SELECT id, agent FROM visits")
	if err != nil {
		return fmt.Errorf("failed to select visits for cleanup: %w", err)
	}
	defer rows.Close()

	type Visit struct {
		id    string
		agent string
	}

	var list []Visit
	for rows.Next() {
		var v Visit
		if err := rows.Scan(&v.id, &v.agent); err != nil {
			return fmt.Errorf("failed to scan visit for cleanup: %w", err)
		}
		list = append(list, v)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start visits cleanup transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE visits SET agent = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare visits update statement: %w", err)
	}
	defer stmt.Close()

	for _, v := range list {
		cAgent := strings.ToUpper(strings.TrimSpace(v.agent))
		_, err = stmt.Exec(cAgent, v.id)
		if err != nil {
			return fmt.Errorf("failed to update visit %d: %w", v.id, err)
		}
	}

	return tx.Commit()
}

func CleanExistingAgents(db *sql.DB) error {
	rows, err := db.Query("SELECT id, zona, agente FROM agents")
	if err != nil {
		return fmt.Errorf("failed to select agents for cleanup: %w", err)
	}
	defer rows.Close()

	type Agent struct {
		id     int
		zona   string
		agente string
	}

	var list []Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.id, &a.zona, &a.agente); err != nil {
			return fmt.Errorf("failed to scan agent for cleanup: %w", err)
		}
		list = append(list, a)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start agents cleanup transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE agents SET zona = ?, agente = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare agents update statement: %w", err)
	}
	defer stmt.Close()

	for _, a := range list {
		cZona := strings.ToUpper(strings.TrimSpace(a.zona))
		cAgente := strings.ToUpper(strings.TrimSpace(a.agente))
		_, err = stmt.Exec(cZona, cAgente, a.id)
		if err != nil {
			return fmt.Errorf("failed to update agent %d: %w", a.id, err)
		}
	}

	return tx.Commit()
}
