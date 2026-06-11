package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"smieci-sms/internal/model"
)

type UserRepository interface {
	SaveUserLocation(ctx context.Context, user model.UserLocation) error
	ListUsers(ctx context.Context) ([]model.UserLocation, error)
	ListUsersWithTodayGarbage(ctx context.Context, today time.Time) ([]model.UserGarbageSchedule, error)
	GetOutdatedLocationIDs(ctx context.Context) ([]int, error)
	SaveGarbageSchedules(ctx context.Context, schedules []model.GarbageSchedule) error
}

type userRepository struct {
	db *DB
}

func NewUserRepository(db *DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) SaveUserLocation(ctx context.Context, user model.UserLocation) error {
	// basic validation
	if strings.TrimSpace(user.Name) == "" {
		return fmt.Errorf("name must not be empty")
	}
	if strings.TrimSpace(user.Phone) == "" {
		return fmt.Errorf("phone must not be empty")
	}
	if user.LocationID == 0 {
		return fmt.Errorf("location_id must not be empty or zero")
	}

	// prepare address value: if empty, store NULL
	var addrArg interface{}
	var addrForLog string
	if strings.TrimSpace(user.AddressName) == "" {
		addrArg = nil
		addrForLog = "NULL"
	} else {
		addrArg = user.AddressName
		// escape single quotes for logging representation
		escaped := strings.ReplaceAll(user.AddressName, "'", "''")
		addrForLog = fmt.Sprintf("'%s'", escaped)
	}

	// prepare log of full insert statement for visibility in cmd
	// NOTE: This is a logged representation only and not executed as-is.
	escapedName := strings.ReplaceAll(user.Name, "'", "''")
	escapedPhone := strings.ReplaceAll(user.Phone, "'", "''")
	logSQL := fmt.Sprintf("INSERT INTO user_locations (name, phone, location_id, address_name) VALUES ('%s','%s',%d,%s)", escapedName, escapedPhone, user.LocationID, addrForLog)
	fmt.Println(logSQL)

	// execute parameterized query using database/sql to remain DB-agnostic
	query := `INSERT INTO user_locations (name, phone, location_id, address_name) VALUES ($1, $2, $3, $4)`
	result, err := r.db.Conn.ExecContext(ctx, query, user.Name, user.Phone, user.LocationID, addrArg)
	if err != nil {
		// some drivers (Postgres) expect $1 style placeholders; to keep this code portable,
		// only the driver needs to match the placeholder style. ExecContext will work with the
		// configured driver for the DB in use.
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		// if driver doesn't support LastInsertId (e.g., Postgres), attempt to fall back gracefully
		if err == sql.ErrNoRows {
			return nil
		}
		// ignore inability to fetch last insert id as it's optional for callers
		return nil
	}
	user.ID = id
	return nil
}

func (r *userRepository) ListUsers(ctx context.Context) ([]model.UserLocation, error) {
	query := `SELECT id, name, phone, location_id, address_name FROM user_locations`
	rows, err := r.db.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.UserLocation
	for rows.Next() {
		var user model.UserLocation
		if err := rows.Scan(&user.ID, &user.Name, &user.Phone, &user.LocationID, &user.AddressName); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) ListUsersWithTodayGarbage(ctx context.Context, today time.Time) ([]model.UserGarbageSchedule, error) {
	query := `
	SELECT
		u.id,
		u.name,
		u.phone,
		u.location_id,
		u.address_name,
		g.location_id,
		g.date_zmieszane,
		g.date_papier,
		g.date_plastik,
		g.date_szklo,
		g.date_bio,
		g.date_zielone,
		g.date_bio_restauracyjne,
		g.date_gabaryty,
		g.last_update
	FROM user_locations u
	JOIN garbage_schedules g ON u.location_id = g.location_id
	`

	rows, err := r.db.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []model.UserGarbageSchedule
	for rows.Next() {
		var item model.UserGarbageSchedule
		var scheduleLocationID int
		if err := rows.Scan(
			&item.User.ID,
			&item.User.Name,
			&item.User.Phone,
			&item.User.LocationID,
			&item.User.AddressName,
			&scheduleLocationID,
			&item.Schedule.DateZmieszane,
			&item.Schedule.DatePapier,
			&item.Schedule.DatePlastik,
			&item.Schedule.DateSzklo,
			&item.Schedule.DateBio,
			&item.Schedule.DateZielone,
			&item.Schedule.DateBioRestauracyjne,
			&item.Schedule.DateGabaryty,
			&item.Schedule.LastUpdate,
		); err != nil {
			return nil, err
		}
		item.Schedule.LocationID = scheduleLocationID
		if item.Schedule.HasToday(today) {
			matches = append(matches, item)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}

func (r *userRepository) SaveGarbageSchedules(ctx context.Context, schedules []model.GarbageSchedule) error {
	query := `
		INSERT INTO garbage_schedules (
			location_id, date_zmieszane, date_papier, date_plastik, 
			date_szklo, date_bio, date_zielone, date_bio_restauracyjne, 
			date_gabaryty, last_update
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (location_id) 
		DO UPDATE SET 
			date_zmieszane = EXCLUDED.date_zmieszane,
			date_papier = EXCLUDED.date_papier,
			date_plastik = EXCLUDED.date_plastik,
			date_szklo = EXCLUDED.date_szklo,
			date_bio = EXCLUDED.date_bio,
			date_zielone = EXCLUDED.date_zielone,
			date_bio_restauracyjne = EXCLUDED.date_bio_restauracyjne,
			date_gabaryty = EXCLUDED.date_gabaryty,
			last_update = EXCLUDED.last_update;`

	for _, s := range schedules {
		_, err := r.db.Conn.ExecContext(ctx, query,
			s.LocationID, s.DateZmieszane, s.DatePapier, s.DatePlastik,
			s.DateSzklo, s.DateBio, s.DateZielone, s.DateBioRestauracyjne,
			s.DateGabaryty, s.LastUpdate,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *userRepository) GetOutdatedLocationIDs(ctx context.Context) ([]int, error) {
	// Selects IDs where last_update is older than exactly 3 days ago
	query := `
		SELECT ul.location_id 
		FROM user_locations ul
		LEFT JOIN garbage_schedules gs ON ul.location_id = gs.location_id
		WHERE gs.last_update IS NULL OR gs.last_update < NOW() - INTERVAL '3 days'
		GROUP BY ul.location_id;`

	rows, err := r.db.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query outdated locations: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan location id: %w", err)
		}
		ids = append(ids, id)
	}

	// Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return ids, nil
}
