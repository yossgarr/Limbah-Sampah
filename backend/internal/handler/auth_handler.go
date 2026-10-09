package handler

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

const (
	RoleUser       = "user"
	RoleObserver   = "observer"
	RoleSuperadmin = "superadmin"

	maxFailedLogins = 5
	tokenTTL        = 8 * time.Hour
)

type Handler struct {
	DB        *sqlx.DB
	JWTSecret []byte
}

type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (h *Handler) issueToken(userID, role string) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}).SignedString(h.JWTSecret)
}

// Authenticate memvalidasi "Authorization: Bearer <jwt>" lalu menyimpan
// user_id dan role di c.Locals.
func (h *Handler) Authenticate(c *fiber.Ctx) error {
	raw := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	if raw == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token tidak ditemukan"})
	}

	var cl claims
	token, err := jwt.ParseWithClaims(raw, &cl, func(*jwt.Token) (any, error) {
		return h.JWTSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token tidak valid atau kedaluwarsa"})
	}

	c.Locals("user_id", cl.Subject)
	c.Locals("role", cl.Role)
	return c.Next()
}

// RequireRole membatasi route hanya untuk role tertentu.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Anda tidak memiliki akses untuk aksi ini"})
	}
}

// ExternalLogin: login email+password untuk akun eksternal.
// Semua kegagalan memakai pesan generik agar email tidak bocor (FR-AUTH-11).
func (h *Handler) ExternalLogin(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	genericErr := fiber.Map{"error": "Email atau password salah"}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(genericErr)
	}

	var u struct {
		ID           string `db:"id"`
		Email        string `db:"email"`
		PasswordHash string `db:"password_hash"`
		Role         string `db:"role"`
		IsActive     bool   `db:"is_active"`
		Expired      bool   `db:"expired"`
		Locked       bool   `db:"locked"`
	}
	err := h.DB.Get(&u, `
		SELECT id, email, COALESCE(password_hash, '') AS password_hash, role, is_active,
		       (berlaku_hingga IS NOT NULL AND berlaku_hingga < NOW()) AS expired,
		       (locked_until IS NOT NULL AND locked_until > NOW())     AS locked
		FROM users
		WHERE auth_source = 'external' AND LOWER(email) = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(fiber.StatusUnauthorized).JSON(genericErr)
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Terjadi kesalahan server"})
	}

	if !u.IsActive || u.Expired || u.Locked || u.PasswordHash == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(genericErr)
	}

	match, err := argon2id.ComparePasswordAndHash(req.Password, u.PasswordHash)
	if err != nil || !match {
		// Kunci 15 menit setelah 5x gagal berturut-turut
		_, _ = h.DB.Exec(`
			UPDATE users SET
				locked_until = CASE WHEN failed_login_count + 1 >= $2 THEN NOW() + INTERVAL '15 minutes' ELSE locked_until END,
				failed_login_count = CASE WHEN failed_login_count + 1 >= $2 THEN 0 ELSE failed_login_count + 1 END
			WHERE id = $1`, u.ID, maxFailedLogins)
		return c.Status(fiber.StatusUnauthorized).JSON(genericErr)
	}

	_, _ = h.DB.Exec(`UPDATE users SET failed_login_count = 0, locked_until = NULL, last_login = NOW() WHERE id = $1`, u.ID)

	token, err := h.issueToken(u.ID, u.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat token"})
	}

	return c.JSON(fiber.Map{
		"access_token": token,
		"expires_in":   int(tokenTTL.Seconds()),
		"user_id":      u.ID,
		"email":        u.Email,
		"role":         u.Role,
	})
}
