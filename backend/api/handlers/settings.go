package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"sunspear/api/middleware"
	"sunspear/config"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type SettingsHandler struct {
	cfg *config.Config
	db  *sql.DB
}

func NewSettingsHandler(cfg *config.Config, db *sql.DB) *SettingsHandler {
	return &SettingsHandler{cfg: cfg, db: db}
}

// GetSettings returns all settings as a JSON object
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT key, value FROM settings")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		settings[key] = value
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, settings)
}

// UpdateSettings upserts multiple settings at once
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var settings map[string]string
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Upsert each key-value pair
	for key, value := range settings {
		_, err := h.db.Exec(
			`INSERT INTO settings (key, value, updated_at)
			 VALUES (?, ?, CURRENT_TIMESTAMP)
			 ON CONFLICT(key) DO UPDATE SET
			 value = excluded.value,
			 updated_at = CURRENT_TIMESTAMP`,
			key, value,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "Settings updated",
	})
}

// ListUsers returns all users (without password hashes)
func (h *SettingsHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, username, created_at FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var username, createdAt string
		if err := rows.Scan(&id, &username, &createdAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, map[string]interface{}{
			"id":         id,
			"username":   username,
			"created_at": createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

// CreateUser creates a new user with hashed password
func (h *SettingsHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert user
	result, err := h.db.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		req.Username,
		string(passwordHash),
	)
	if err != nil {
		// Check for duplicate username (SQLite UNIQUE constraint violation)
		if err.Error() == "UNIQUE constraint failed: users.username" {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newID, _ := result.LastInsertId()

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":       newID,
		"username": req.Username,
	})
}

// DeleteUser deletes a user, but prevents deleting the last user
func (h *SettingsHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if this is the last user
	var count int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count <= 1 {
		http.Error(w, "Cannot delete the last user", http.StatusBadRequest)
		return
	}

	// Delete the user
	result, err := h.db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "User deleted",
	})
}

// ChangePassword changes a user's password
func (h *SettingsHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get target user ID from URL
	vars := mux.Vars(r)
	targetIDStr := vars["id"]
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get requesting user's ID from context
	requestingID := r.Context().Value(middleware.UserIDKey).(int)

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Always verify current password (own or other user's password change)
	var passwordHash string
	err = h.db.QueryRow("SELECT password_hash FROM users WHERE id = ?", requestingID).Scan(&passwordHash)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update password
	result, err := h.db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(newPasswordHash), targetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "Password changed",
	})
}
