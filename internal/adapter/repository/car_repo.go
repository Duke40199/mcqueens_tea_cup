package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
	"McQueens_Tea_Cup/pkg/logger"
	"McQueens_Tea_Cup/pkg/utils"

	"github.com/lib/pq"
)

type CarRepository struct {
	DB             *sql.DB
	carsTable      string
	carStylesTable string
}

func NewCarRepository(db *sql.DB, tables config.DatabaseTablesConfig) port.CarRepository {
	return &CarRepository{
		DB:             db,
		carsTable:      quoteIdent(tables.IDACCarsMetadata),
		carStylesTable: quoteIdent(tables.IDACCarStylesMetadata),
	}
}

func (r *CarRepository) UpsertCars(ctx context.Context, cars []entity.CarMetadata) error {
	if len(cars) == 0 {
		return nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // no-op once Commit succeeds

	// Chunk so we never exceed Postgres's 65535 bind-parameter limit (7 params/row).
	for _, batch := range utils.ChunkSlice(cars, pgMaxBulkRows) {
		query := fmt.Sprintf(`INSERT INTO %s (id, sega_id, name, model_code, maker, base_spec, style_ids) VALUES `, r.carsTable)
		values := make([]any, 0, len(batch)*7)
		placeholders := make([]string, 0, len(batch))
		for i, car := range batch {
			placeholders = append(placeholders,
				fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d)", i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7))
			values = append(values, car.ID, car.SegaCarID, car.Name, car.ModelCode, car.Maker, car.BaseStyleName, pq.Int64Array(car.CarStyleIDs))
		}
		query += strings.Join(placeholders, ",")
		query += ` ON CONFLICT (sega_id) DO UPDATE SET name = EXCLUDED.name, maker = EXCLUDED.maker, base_spec = EXCLUDED.base_spec;`

		if _, err := tx.ExecContext(ctx, query, values...); err != nil {
			logger.Error(ctx, "error upserting cars", err)
			return err
		}
	}
	return tx.Commit()
}

func (r *CarRepository) UpsertCarStyles(ctx context.Context, styles []entity.CarStyleMetadata) error {
	if len(styles) == 0 {
		return nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // no-op once Commit succeeds

	// Chunk so we never exceed Postgres's 65535 bind-parameter limit (4 params/row).
	for _, batch := range utils.ChunkSlice(styles, pgMaxBulkRows) {
		query := fmt.Sprintf(`INSERT INTO %s (id, sega_id, name, car_id) VALUES `, r.carStylesTable)
		values := make([]any, 0, len(batch)*4)
		placeholders := make([]string, 0, len(batch))
		for i, style := range batch {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
			values = append(values, style.ID, style.StyleCarID, style.RouteStyleName, style.CarID)
		}
		query += strings.Join(placeholders, ",")
		query += ` ON CONFLICT (sega_id) DO UPDATE SET name = EXCLUDED.name;`

		if _, err := tx.ExecContext(ctx, query, values...); err != nil {
			logger.Error(ctx, "error upserting car styles", err)
			return err
		}
	}
	return tx.Commit()
}

// GetBaseSpecMap returns a map of (model_code OR alias) -> CarSpecInfo
// e.g. "FL5" -> {ModelCode: "FL5", BaseSpec: "tech"}, "CZ4Aエボ10" -> {ModelCode: "CZ4A", BaseSpec: "speed"}
func (r *CarRepository) GetBaseSpecMap(ctx context.Context) (map[string]entity.CarSpecInfo, error) {
	query := fmt.Sprintf(`SELECT name, maker, model_code, base_spec, COALESCE(aliases, '{}') FROM %s WHERE base_spec != ''`, r.carsTable)
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		logger.Error(ctx, "error getting base spec map", err)
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
	query := fmt.Sprintf(`SELECT sega_id, id FROM %s`, r.carsTable)
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

func (r *CarRepository) GetCarWithSpecsByAliases(ctx context.Context, aliasSpecMap map[string]string) (map[string]entity.CarSpecInfo, error) {
	query := fmt.Sprintf(`SELECT
	 c.maker,
	 c.name AS car_name,
 	c.model_code,
 	c.base_spec,
 	c.aliases,
	cs.name AS spec_name,
 	cs.sega_id as sega_spec_id
	FROM %s c
	LEFT JOIN %s cs
  	ON c.id = cs.car_id
	WHERE `, r.carsTable, r.carStylesTable)
	if len(aliasSpecMap) == 0 {
		return make(map[string]entity.CarSpecInfo), nil
	}
	// Build the WHERE clause with bound parameters to prevent SQL injection.
	// Each group reuses two placeholders ($key, $value) across both conditions,
	// and is parenthesized so the AND/OR precedence groups as intended.
	conditions := make([]string, 0, len(aliasSpecMap))
	args := make([]any, 0, len(aliasSpecMap)*2)
	for key, value := range aliasSpecMap {
		keyPlaceholder := fmt.Sprintf("$%d", len(args)+1)
		valuePlaceholder := fmt.Sprintf("$%d", len(args)+2)
		conditions = append(conditions, fmt.Sprintf(
			`((%[1]s = ANY(aliases) AND cs.name = %[2]s) OR (c.model_code = %[1]s AND cs.name = %[2]s))`,
			keyPlaceholder, valuePlaceholder,
		))
		args = append(args, key, value)
	}
	query += strings.Join(conditions, " OR ")
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error(ctx, "error getting car with specs by aliases", err)
		return nil, err
	}
	defer rows.Close()
	// key: sega_spec_id, since it's unique to each car
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
		result[segaSpecID] = info
	}
	return result, rows.Err()
}

func (r *CarRepository) GetListCarWithAggregatedSpecs(ctx context.Context) ([]*entity.CarMetadata, error) {
	query := fmt.Sprintf(`SELECT
    c.sega_id AS sega_car_id,
    c.maker,
    c.name AS car_name,
	c.model_code,
    array_agg(cs.sega_id) AS spec_ids,
    array_agg(cs.name) AS spec_names
    FROM %s c
    LEFT JOIN %s cs
    ON c.id = cs.car_id
	GROUP BY c.id
	ORDER BY c.maker, c.name ASC;`, r.carsTable, r.carStylesTable)
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		logger.Error(ctx, "error getting list car with aggregated specs", err)
		return nil, err
	}
	defer rows.Close()

	result := make([]*entity.CarMetadata, 0)
	for rows.Next() {
		var segaCarID int64
		var maker, name, modelCode string
		var aggregatedCarSpecIDs, aggregatedCarSpecNames pq.StringArray
		if err := rows.Scan(&segaCarID, &maker, &name, &modelCode, &aggregatedCarSpecIDs, &aggregatedCarSpecNames); err != nil {
			return nil, err
		}
		info := entity.CarMetadata{
			SegaCarID: segaCarID,
			Maker:     maker,
			Name:      name,
			ModelCode: modelCode,
			SpecNames: aggregatedCarSpecNames,
			SpecIDs:   aggregatedCarSpecIDs,
		}
		// Map model_code -> CarSpecInfo
		result = append(result, &info)
	}
	return result, rows.Err()
}
