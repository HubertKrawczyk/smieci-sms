package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"smieci-sms/internal/model"
)

type UserRepository interface {
	SaveUserLocation(ctx context.Context, user model.UserLocation) error
	DeleteUserLocationByChatID(ctx context.Context, chatID int64) error
	ListUsers(ctx context.Context) ([]model.UserLocation, error)
	ListUsersWithTodayGarbage(ctx context.Context, today time.Time) ([]model.UserGarbageSchedule, error)
	GetOutdatedLocationIDs(ctx context.Context) ([]string, error)
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
	if user.LocationID == "" {
		return fmt.Errorf("location_id must not be empty")
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
	logSQL := fmt.Sprintf("INSERT INTO user_locations (chat_id, location_id, name, phone, address_name) VALUES (%d,'%s','%s','%s',%s)", user.ChatID, user.LocationID, escapedName, escapedPhone, addrForLog)
	fmt.Println(logSQL)

	// execute in transaction to keep locations and notifications in sync
	tx, err := r.db.Conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// execute parameterized query using RETURNING id to remain DB-agnostic & fetch generated ID
	query := `INSERT INTO user_locations (chat_id, location_id, name, phone, address_name) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var insertedID int64
	err = tx.QueryRowContext(ctx, query, user.ChatID, user.LocationID, user.Name, user.Phone, addrArg).Scan(&insertedID)
	if err != nil {
		return fmt.Errorf("failed to insert user location: %w", err)
	}

	// insert each notification preference setting
	for _, setting := range user.NotificationSettings {
		if strings.TrimSpace(setting) == "" {
			continue
		}
		prefQuery := `INSERT INTO user_notifications (user_location_id, notification_time) VALUES ($1, $2)`
		_, err := tx.ExecContext(ctx, prefQuery, insertedID, setting)
		if err != nil {
			return fmt.Errorf("failed to insert notification preference: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	user.ID = insertedID
	return nil
}

func (r *userRepository) DeleteUserLocationByChatID(ctx context.Context, chatID int64) error {
	query := `DELETE FROM user_locations WHERE chat_id = $1`
	_, err := r.db.Conn.ExecContext(ctx, query, chatID)
	return err
}

func (r *userRepository) ListUsers(ctx context.Context) ([]model.UserLocation, error) {
	query := `
		SELECT ul.id, ul.chat_id, ul.name, ul.phone, ul.location_id, ul.address_name, COALESCE(un.notification_time, '')
		FROM user_locations ul
		LEFT JOIN user_notifications un ON ul.id = un.user_location_id
	`
	rows, err := r.db.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userMap := make(map[int64]*model.UserLocation)
	var orderedIDs []int64

	for rows.Next() {
		var id int64
		var chatID int64
		var name string
		var phone string
		var locationID string
		var addressName string
		var notificationTime string

		if err := rows.Scan(&id, &chatID, &name, &phone, &locationID, &addressName, &notificationTime); err != nil {
			return nil, err
		}

		user, exists := userMap[id]
		if !exists {
			user = &model.UserLocation{
				ID:          id,
				ChatID:      chatID,
				Name:        name,
				Phone:       phone,
				LocationID:  locationID,
				AddressName: addressName,
			}
			userMap[id] = user
			orderedIDs = append(orderedIDs, id)
		}
		if notificationTime != "" {
			user.NotificationSettings = append(user.NotificationSettings, notificationTime)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	users := make([]model.UserLocation, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		users = append(users, *userMap[id])
	}
	return users, nil
}

func (r *userRepository) ListUsersWithTodayGarbage(ctx context.Context, today time.Time) ([]model.UserGarbageSchedule, error) {
	query := `
	SELECT
		u.id,
		u.chat_id,
		u.name,
		u.phone,
		u.location_id,
		u.address_name,
		COALESCE(un.notification_time, ''),
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
	LEFT JOIN user_notifications un ON u.id = un.user_location_id
	`

	rows, err := r.db.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type key struct {
		userID     int64
		locationID string
	}
	matchMap := make(map[key]*model.UserGarbageSchedule)
	var orderedKeys []key

	for rows.Next() {
		var id int64
		var chatID int64
		var name string
		var phone string
		var locationID string
		var addressName string
		var notificationTime string
		var scheduleLocationID string
		var sched model.GarbageSchedule

		if err := rows.Scan(
			&id,
			&chatID,
			&name,
			&phone,
			&locationID,
			&addressName,
			&notificationTime,
			&scheduleLocationID,
			&sched.DateZmieszane,
			&sched.DatePapier,
			&sched.DatePlastik,
			&sched.DateSzklo,
			&sched.DateBio,
			&sched.DateZielone,
			&sched.DateBioRestauracyjne,
			&sched.DateGabaryty,
			&sched.LastUpdate,
		); err != nil {
			return nil, err
		}
		sched.LocationID = scheduleLocationID

		k := key{userID: id, locationID: locationID}
		item, exists := matchMap[k]
		if !exists {
			item = &model.UserGarbageSchedule{
				User: model.UserLocation{
					ID:          id,
					ChatID:      chatID,
					Name:        name,
					Phone:       phone,
					LocationID:  locationID,
					AddressName: addressName,
				},
				Schedule: sched,
			}
			matchMap[k] = item
			orderedKeys = append(orderedKeys, k)
		}
		if notificationTime != "" {
			item.User.NotificationSettings = append(item.User.NotificationSettings, notificationTime)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var matches []model.UserGarbageSchedule
	for _, k := range orderedKeys {
		item := matchMap[k]
		if item.Schedule.HasToday(today) {
			matches = append(matches, *item)
		}
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

func (r *userRepository) GetOutdatedLocationIDs(ctx context.Context) ([]string, error) {
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

	var ids []string
	for rows.Next() {
		var id string
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
