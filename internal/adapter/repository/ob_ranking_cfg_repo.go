package repository

import (
	"database/sql"
	"fmt"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"

	_ "github.com/lib/pq"
)

type OBRankingCfgRepository struct {
	DB        *sql.DB
	tableName string
}

// NewOBRankingCfgRepository returns the struct that satisfies AliasRepository
func NewOBRankingCfgRepository(db *sql.DB, tables config.DatabaseTablesConfig) database.OBRankingCfgRepository {
	return &OBRankingCfgRepository{
		DB:        db,
		tableName: tables.OBRankingCfg,
	}
}

// GetByAliasKey fetches alias from DB
func (o *OBRankingCfgRepository) GetBySegaID(key string) (*entity.OBRankingCfg, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE sega_id = $1`, o.tableName)

	row := o.DB.QueryRow(query, key)

	var cfg entity.OBRankingCfg
	err := row.Scan(&cfg.ID, &cfg.SegaID, &cfg.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err // Not found
		}
		// Log error in a real app
		fmt.Println("DB Error:", err)
		return nil, err
	}

	return &cfg, nil
}

func (o *OBRankingCfgRepository) GetRankingCfgMap() (map[string]entity.OBRankingCfg, error) {
	query := fmt.Sprintf(`SELECT * FROM %s`, o.tableName)

	rows, err := o.DB.Query(query)
	if err != nil {
		// Log error in a real app
		fmt.Println("DB Error:", err)
		return nil, err
	}
	defer rows.Close()

	var cfgMap = make(map[string]entity.OBRankingCfg)
	for rows.Next() {
		var cfg entity.OBRankingCfg
		err = rows.Scan(&cfg.ID, &cfg.Name, &cfg.SegaID, &cfg.Emoji)
		if err != nil {
			fmt.Println("Failed to scan:", err)
			return nil, err
		}
		cfgMap[cfg.SegaID] = cfg
	}
	return cfgMap, nil
}
