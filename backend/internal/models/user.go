package models

import (
	"github.com/google/uuid"
	"time"
)

type Role string
type AuthSource string

const (
	RoleUser       Role = "user"       // Bisa melapor, lihat data sendiri
	RoleObserver   Role = "observer"   // Lihat agregat
	RoleSuperadmin Role = "superadmin" // Kelola user & referensi

	AuthSourceSSO      AuthSource = "sso"
	AuthSourceExternal AuthSource = "external"
)

// User mematuhi aturan minimisasi data (PRD 2.3)
type User struct {
	ID         uuid.UUID  `json:"id_internal" db:"id_internal"` // ID Pseudonim
	AuthSource AuthSource `json:"auth_source" db:"auth_source"`

	// Khusus SSO
	KeycloakSub *string `json:"-" db:"keycloak_sub"` // Tidak pernah di-expose
	KeycloakIss *string `json:"-" db:"keycloak_iss"`

	// Khusus Akun Eksternal (Pengamat)
	Email            *string    `json:"email,omitempty" db:"email"`                   // Hanya untuk eksternal
	PasswordHash     *string    `json:"-" db:"password_hash"`                         // Argon2id
	ValidUntil       *time.Time `json:"berlaku_hingga,omitempty" db:"berlaku_hingga"` // Maks 12 bulan
	FailedLoginCount int        `json:"-" db:"failed_login_count"`                    // Maks 5 kali gagal
	LockedUntil      *time.Time `json:"-" db:"locked_until"`                          // Terkunci 15 menit

	// Data Bersama
	Role      Role       `json:"role" db:"role"`
	IsActive  bool       `json:"is_active" db:"is_active"`
	LastLogin *time.Time `json:"last_login" db:"last_login"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`

	// Atribut Analitik (Disalin dari SSO, PRD 2.4)
	SatuanKerja   *string `json:"satuan_kerja,omitempty" db:"satuan_kerja"`
	UnitKerja     *string `json:"unit_kerja,omitempty" db:"unit_kerja"`
	StatusPegawai *string `json:"status_pegawai,omitempty" db:"status_pegawai"`
	GolonganGroup *string `json:"golongan_group,omitempty" db:"golongan_group"`
	AgeBand       *string `json:"age_band,omitempty" db:"age_band"`
}
