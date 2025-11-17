package repository

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/powerlifting-coach-app/user-service/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (keycloak_id, email, name, user_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query, user.KeycloakID, user.Email, user.Name, user.UserType).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if user.UserType == models.UserTypeAthlete {
		if err := r.createAthleteProfile(user.ID); err != nil {
			return fmt.Errorf("failed to create athlete profile: %w", err)
		}
	} else if user.UserType == models.UserTypeCoach {
		if err := r.createCoachProfile(user.ID); err != nil {
			return fmt.Errorf("failed to create coach profile: %w", err)
		}
	}

	return nil
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	query := `SELECT id, keycloak_id, email, name, user_type, created_at, updated_at FROM users WHERE id = $1`
	
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.Name, &user.UserType,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByKeycloakID(keycloakID string) (*models.User, error) {
	query := `SELECT id, keycloak_id, email, name, user_type, created_at, updated_at FROM users WHERE keycloak_id = $1`
	
	user := &models.User{}
	err := r.db.QueryRow(query, keycloakID).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.Name, &user.UserType,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, keycloak_id, email, name, user_type, created_at, updated_at FROM users WHERE email = $1`
	
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.Name, &user.UserType,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) createAthleteProfile(userID uuid.UUID) error {
	query := `
		INSERT INTO athlete_profiles (user_id, training_frequency)
		VALUES ($1, 3)`
	
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *UserRepository) createCoachProfile(userID uuid.UUID) error {
	query := `
		INSERT INTO coach_profiles (user_id, years_experience)
		VALUES ($1, 0)`
	
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *UserRepository) GetAthleteProfile(userID uuid.UUID) (*models.AthleteProfile, error) {
	query := `
		SELECT id, user_id, weight_kg, experience_level, competition_date, access_code,
		       access_code_expires_at, squat_max_kg, bench_max_kg, deadlift_max_kg,
		       training_frequency, goals, injuries,
		       COALESCE(bio, '') as bio,
		       COALESCE(target_weight_class, '') as target_weight_class,
		       COALESCE(preferred_federation, '') as preferred_federation,
		       created_at, updated_at
		FROM athlete_profiles WHERE user_id = $1`

	profile := &models.AthleteProfile{}
	var bio, weightClass, federation string
	err := r.db.QueryRow(query, userID).Scan(
		&profile.ID, &profile.UserID, &profile.WeightKg, &profile.ExperienceLevel,
		&profile.CompetitionDate, &profile.AccessCode, &profile.AccessCodeExpiresAt,
		&profile.SquatMaxKg, &profile.BenchMaxKg, &profile.DeadliftMaxKg,
		&profile.TrainingFrequency, &profile.Goals, &profile.Injuries,
		&bio, &weightClass, &federation,
		&profile.CreatedAt, &profile.UpdatedAt,
	)
	if bio != "" {
		profile.Bio = &bio
	}
	if weightClass != "" {
		profile.TargetWeightClass = &weightClass
	}
	if federation != "" {
		profile.PreferredFederation = &federation
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("athlete profile not found")
		}
		return nil, fmt.Errorf("failed to get athlete profile: %w", err)
	}

	return profile, nil
}

func (r *UserRepository) UpdateAthleteProfile(userID uuid.UUID, req models.UpdateAthleteProfileRequest) error {
	query := `
		UPDATE athlete_profiles SET
			weight_kg = COALESCE($2, weight_kg),
			experience_level = COALESCE($3, experience_level),
			competition_date = COALESCE($4, competition_date),
			squat_max_kg = COALESCE($5, squat_max_kg),
			bench_max_kg = COALESCE($6, bench_max_kg),
			deadlift_max_kg = COALESCE($7, deadlift_max_kg),
			training_frequency = COALESCE($8, training_frequency),
			goals = COALESCE($9, goals),
			injuries = COALESCE($10, injuries)
		WHERE user_id = $1`

	_, err := r.db.Exec(query, userID, req.WeightKg, req.ExperienceLevel, req.CompetitionDate,
		req.SquatMaxKg, req.BenchMaxKg, req.DeadliftMaxKg, req.TrainingFrequency,
		req.Goals, req.Injuries)
	
	if err != nil {
		return fmt.Errorf("failed to update athlete profile: %w", err)
	}

	return nil
}

func (r *UserRepository) GetCoachProfile(userID uuid.UUID) (*models.CoachProfile, error) {
	query := `
		SELECT id, user_id, bio, certifications, years_experience, specializations,
		       hourly_rate, created_at, updated_at
		FROM coach_profiles WHERE user_id = $1`
	
	profile := &models.CoachProfile{}
	err := r.db.QueryRow(query, userID).Scan(
		&profile.ID, &profile.UserID, &profile.Bio, pq.Array(&profile.Certifications),
		&profile.YearsExperience, pq.Array(&profile.Specializations),
		&profile.HourlyRate, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("coach profile not found")
		}
		return nil, fmt.Errorf("failed to get coach profile: %w", err)
	}

	return profile, nil
}

func (r *UserRepository) UpdateCoachProfile(userID uuid.UUID, req models.UpdateCoachProfileRequest) error {
	query := `
		UPDATE coach_profiles SET
			bio = COALESCE($2, bio),
			certifications = COALESCE($3, certifications),
			years_experience = COALESCE($4, years_experience),
			specializations = COALESCE($5, specializations),
			hourly_rate = COALESCE($6, hourly_rate)
		WHERE user_id = $1`

	_, err := r.db.Exec(query, userID, req.Bio, pq.Array(req.Certifications),
		req.YearsExperience, pq.Array(req.Specializations), req.HourlyRate)
	
	if err != nil {
		return fmt.Errorf("failed to update coach profile: %w", err)
	}

	return nil
}

func (r *UserRepository) GenerateAccessCode(userID uuid.UUID, expiresInWeeks *int) (string, error) {
	rand.Seed(time.Now().UnixNano())
	
	for {
		code := ""
		for i := 0; i < 6; i++ {
			code += strconv.Itoa(rand.Intn(10))
		}

		var expiresAt *time.Time
		if expiresInWeeks != nil && *expiresInWeeks > 0 && *expiresInWeeks <= 12 {
			expiry := time.Now().AddDate(0, 0, (*expiresInWeeks)*7)
			expiresAt = &expiry
		}

		query := `
			UPDATE athlete_profiles 
			SET access_code = $2, access_code_expires_at = $3
			WHERE user_id = $1`
		
		_, err := r.db.Exec(query, userID, code, expiresAt)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				continue
			}
			return "", fmt.Errorf("failed to generate access code: %w", err)
		}

		return code, nil
	}
}

func (r *UserRepository) GetAthleteByAccessCode(accessCode string) (*models.User, error) {
	query := `
		SELECT u.id, u.keycloak_id, u.email, u.name, u.user_type, u.created_at, u.updated_at
		FROM users u
		JOIN athlete_profiles ap ON u.id = ap.user_id
		WHERE ap.access_code = $1 
		AND (ap.access_code_expires_at IS NULL OR ap.access_code_expires_at > NOW())`
	
	user := &models.User{}
	err := r.db.QueryRow(query, accessCode).Scan(
		&user.ID, &user.KeycloakID, &user.Email, &user.Name, &user.UserType,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired access code")
		}
		return nil, fmt.Errorf("failed to get athlete by access code: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GrantCoachAccess(coachID, athleteID uuid.UUID, accessCode string) error {
	query := `
		INSERT INTO coach_athlete_access (coach_id, athlete_id, access_code)
		VALUES ($1, $2, $3)
		ON CONFLICT (coach_id, athlete_id) 
		DO UPDATE SET access_code = $3, granted_at = NOW(), is_active = TRUE`
	
	_, err := r.db.Exec(query, coachID, athleteID, accessCode)
	if err != nil {
		return fmt.Errorf("failed to grant coach access: %w", err)
	}

	return nil
}

func (r *UserRepository) GetCoachAthletes(coachID uuid.UUID) ([]models.User, error) {
	query := `
		SELECT u.id, u.keycloak_id, u.email, u.name, u.user_type, u.created_at, u.updated_at
		FROM users u
		JOIN coach_athlete_access caa ON u.id = caa.athlete_id
		WHERE caa.coach_id = $1 AND caa.is_active = TRUE`
	
	rows, err := r.db.Query(query, coachID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coach athletes: %w", err)
	}
	defer rows.Close()

	var athletes []models.User
	for rows.Next() {
		var athlete models.User
		err := rows.Scan(
			&athlete.ID, &athlete.KeycloakID, &athlete.Email, &athlete.Name,
			&athlete.UserType, &athlete.CreatedAt, &athlete.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan athlete: %w", err)
		}
		athletes = append(athletes, athlete)
	}

	return athletes, nil
}