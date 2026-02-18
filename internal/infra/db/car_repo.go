package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"McQueens_Tea_Cup/internal/adapter/postgres"
	"McQueens_Tea_Cup/internal/domain/entity"

	"github.com/lib/pq"
)

type CarRepository struct {
	DB *sql.DB
}

func NewCarRepository(db *sql.DB) postgres.CarRepository {
	return &CarRepository{DB: db}
}

func (r *CarRepository) UpsertCars(ctx context.Context, cars []entity.CarMetadata) error {
	if len(cars) == 0 {
		return nil
	}

	query := `INSERT INTO sega_idac_cars_metadata (id, sega_id, name, model_code, maker, base_spec, style_ids) VALUES `
	values := []any{}
	placeholders := []string{}
	// Batch insert
	for i, car := range cars {
		placeholders = append(placeholders,
			fmt.Sprintf("($%d,$%d, $%d, $%d, $%d, $%d, $%d)", i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7))
		values = append(values, car.ID, car.SegaCarID, car.Name, car.ModelCode, car.Maker, car.BaseStyleName, pq.Int64Array(car.CarStyleIDs))
	}
	query += strings.Join(placeholders, ",")
	query += ` ON CONFLICT (sega_id) DO UPDATE SET name = EXCLUDED.name, maker = EXCLUDED.maker, base_spec = EXCLUDED.base_spec;`

	_, err := r.DB.ExecContext(ctx, query, values...)
	if err != nil {
		log.Println("error upserting cars:", err)
		return err
	}
	return err
}

func (r *CarRepository) UpsertCarStyles(ctx context.Context, styles []entity.CarStyleMetadata) error {
	if len(styles) == 0 {
		return nil
	}
	query := `INSERT INTO sega_idac_car_styles_metadata (id, sega_id, name, car_id) VALUES `
	values := []any{}
	placeholders := []string{}

	for i, style := range styles {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		values = append(values, style.ID, style.StyleCarID, style.RouteStyleName, style.CarID)
	}

	query += strings.Join(placeholders, ",")
	query += ` ON CONFLICT (sega_id) DO UPDATE SET name = EXCLUDED.name;`

	_, err := r.DB.ExecContext(ctx, query, values...)
	return err
}

// GetBaseSpecMap returns a map of (model_code OR alias) -> CarSpecInfo
// e.g. "FL5" -> {ModelCode: "FL5", BaseSpec: "tech"}, "CZ4Aエボ10" -> {ModelCode: "CZ4A", BaseSpec: "speed"}
func (r *CarRepository) GetBaseSpecMap(ctx context.Context) (map[string]entity.CarSpecInfo, error) {
	query := `SELECT name, maker, model_code, base_spec, COALESCE(aliases, '{}') FROM sega_idac_cars_metadata WHERE base_spec != ''`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		log.Println("error getting base spec map:", err)
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]entity.CarSpecInfo)
	for rows.Next() {
		var name, maker, modelCode, baseSpec string
		var aliases pq.StringArray
		if err := rows.Scan(&name, &maker, &modelCode, &baseSpec, &aliases); err != nil {
			return nil, err
		}
		info := entity.CarSpecInfo{Maker: maker, CarName: name, ModelCode: modelCode, BaseSpec: baseSpec}
		// Map model_code -> CarSpecInfo
		result[modelCode] = info
		// Map each alias -> same CarSpecInfo (resolves to canonical model_code)
		for _, alias := range aliases {
			result[alias] = info
		}
	}
	return result, rows.Err()
}
