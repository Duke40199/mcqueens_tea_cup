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
		log.Printf("error upserting cars: %v. Query: %s", err, formatQuery(query, values))
		return err
	}
	return nil
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
	if err != nil {
		log.Printf("error upserting styles: %v. Query: %s", err, formatQuery(query, values))
	}
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

func (r *CarRepository) GetSegaIDToUUIDMap(ctx context.Context) (map[int64]string, error) {
	query := `SELECT sega_id, id FROM sega_idac_cars_metadata`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]string)
	for rows.Next() {
		var segaID int64
		var id string
		if err := rows.Scan(&segaID, &id); err != nil {
			return nil, err
		}
		result[segaID] = id
	}
	return result, rows.Err()
}

func formatQuery(query string, args []any) string {
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		var val string
		switch v := arg.(type) {
		case string:
			val = fmt.Sprintf("'%s'", v)
		case int, int64, float64:
			val = fmt.Sprintf("%v", v)
		case bool:
			val = fmt.Sprintf("%v", v)
		default:
			val = fmt.Sprintf("'%v'", v)
		}
		query = strings.Replace(query, placeholder, val, 1)
	}
	return query
}

func (r *CarRepository) GetCarWithSpecsByAliases(ctx context.Context, aliasSpecMap map[string]string) (map[string]entity.CarSpecInfo, error) {
	query := `SELECT 
	 c.maker,
	 c.name AS car_name,
 	c.model_code,
 	c.base_spec,
 	c.aliases,
	cs.name AS spec_name,
 	cs.sega_id as sega_spec_id
	FROM sega_idac_cars_metadata c
	LEFT JOIN sega_idac_car_styles_metadata cs 
  	ON c.id = cs.car_id 
	WHERE `
	count := 0
	for key, value := range aliasSpecMap {
		query += fmt.Sprintf(`'%s' = ANY(aliases) AND cs.name = '%s'`, key, value)
		query += fmt.Sprintf(` OR (c.model_code='%s' AND cs.name = '%s')`, key, value)
		count++
		if count < len(aliasSpecMap) {
			query += ` OR `
		} else {
			query += `;`
		}
	}
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		log.Println("error getting base spec map:", err)
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]entity.CarSpecInfo)
	for rows.Next() {
		var maker, name, modelCode, baseSpec, specStyleName, segaSpecID string
		var aliases pq.StringArray

		if err := rows.Scan(&maker, &name, &modelCode, &baseSpec, &aliases, &specStyleName, &segaSpecID); err != nil {
			return nil, err
		}
		info := entity.CarSpecInfo{
			Maker:         maker,
			CarName:       name,
			ModelCode:     modelCode,
			BaseSpec:      baseSpec,
			SpecStyleName: specStyleName,
			SegaSpecID:    segaSpecID,
			Aliases:       aliases,
		}
		// Map model_code -> CarSpecInfo
		result[modelCode] = info
	}
	return result, rows.Err()
}
