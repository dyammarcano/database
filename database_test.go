//go:build windows

package database

import (
	"testing"
)

func TestDatabase_Connect(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "Test Oracle Connection",
			config: &Config{
				DriverName: "sqlite",
				User:       "test",
				Pass:       "test",
				Host:       "localhost",
				Name:       "oracleDB",
				Port:       1521,
				FilePath:   "C:\\arqprod_local\\mydatabase.sqlite",
			},
			wantErr: false, // assuming that an Oracle DB is run on localhost
		},
		{
			name:    "Test Unrecognized Database",
			config:  &Config{DriverName: "unrecognized"},
			wantErr: true,
		},
	}

	db := NewDatabase()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Connect(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
