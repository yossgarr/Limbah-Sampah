package models

import (
	"time"
	"github.com/google/uuid"
)

type Role string
type AuthSource string

const (
	RoleUser       Role = "user"       // Bisa melapor, lihat data sendiri
	RoleObserver   Role = "observer"   // Lihat agregat
	RoleSuperadmin Role = "superadmin" // Kelola user & referensi[cite: 1]

	AuthSourceSSO      AuthSource = "sso"
	AuthSourceExternal AuthSource = "external"
)

// User mematuhi aturan minimisasi data (PRD 2.3)[cite: 1]
type User struct {
	ID KarbonIKN      uuid.UUID  `json:"id_internal" db:"id_internal"` // ID Pseudonim[cite: 1]
	AuthSource        AuthSource `json:"auth_source" db:"auth_source"`
	
	// Khusus SSO
	KeycloakSub       *string    `json:"-" db:"keycloak_sub"` // Tidak pernah di-expose[cite: 1]
	KeycloakIss       *string    `json:"-" db:"keycloak_iss"`
	
	// Khusus Akun Eksternal (Pengamat)
	Email             *string    `json:"email,omitempty" db:"email"` // Hanya untuk eksternal[cite: 1]
	PasswordHash      *string    `json:"-" db:"password_hash"`       // Argon2id[cite: 1]
	ValidUntil        *time.Time `json:"berlaku_hingga,omitempty" db:"berlaku_hingga"` // Maks 12 bulan[cite: 1]
	FailedLoginCount  int        `json:"-" db:"failed_login_count"`  // Maks 5 kali gagal[cite: 1]
	LockedUntil       *time.Time `json:"-" db:"locked_until"`        // Terkunci 15 menit[cite: 1]

	// Data Bersama
	Role              Role       `json:"role" db:"role"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	LastLogin         *time.Time `json:"last_login" db:"last_login"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`

	// Atribut Analitik (Disalin dari SSO, PRD 2.4)[cite: 1]
	SatuanKerja       *string    `json:"satuan_kerja,omitempty" db:"satuan_kerja"`
	UnitKerja         *string    `json:"unit_kerja,omitempty" db:"unit_kerja"`
	StatusPegawai     *string    `json:"status_pegawai,omitempty" db:"status_pegawai"`
	GolonganGroup     *string    `json:"golongan_group,omitempty" db:"golongan_group"`
	AgeBand           *string    `json:"age_band,omitempty" db:"age_band"`
}