package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "hashes normal password",
			password: "correcthorse",
			wantErr:  false,
		},
		{
			name:     "hashes password with special chars",
			password: "P@ssw0rd!#$%",
			wantErr:  false,
		},
		{
			name:     "hashes empty password",
			password: "",
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPassword(tc.password)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			// Hash must differ from plaintext
			assert.NotEqual(t, tc.password, hash)
			// Two hashes of the same password must differ (bcrypt salts)
			hash2, err := HashPassword(tc.password)
			require.NoError(t, err)
			assert.NotEqual(t, hash, hash2)
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "correct_horse_battery"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		hash     string
		password string
		wantErr  bool
	}{
		{
			name:     "correct password verifies",
			hash:     hash,
			password: password,
			wantErr:  false,
		},
		{
			name:     "wrong password fails",
			hash:     hash,
			password: "wrong_password",
			wantErr:  true,
		},
		{
			name:     "empty password against real hash fails",
			hash:     hash,
			password: "",
			wantErr:  true,
		},
		{
			name:     "password against empty hash fails",
			hash:     "",
			password: password,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := VerifyPassword(tc.hash, tc.password)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
