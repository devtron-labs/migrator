package main

import (
	"reflect"
	"testing"

	"go.uber.org/zap"
)

func TestMigrateConfig_Valid(t *testing.T) {
	type fields struct {
		DatabaseUrl   string
		TargetVersion uint
		DbType        string
		UserName      string
		Password      string
		Host          string
		Port          string
		DbName        string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := MigrateConfig{
				DatabaseUrl:   tt.fields.DatabaseUrl,
				TargetVersion: tt.fields.TargetVersion,
				DbType:        tt.fields.DbType,
				UserName:      tt.fields.UserName,
				Password:      tt.fields.Password,
				Host:          tt.fields.Host,
				Port:          tt.fields.Port,
				DbName:        tt.fields.DbName,
			}
			if got := cfg.Valid(); got != tt.want {
				t.Errorf("MigrateConfig.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMigrateConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    *MigrateConfig
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMigrateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMigrateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMigrateConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMigrateUtil(t *testing.T) {
	type args struct {
		cfg    *MigrateConfig
		logger *zap.SugaredLogger
	}
	tests := []struct {
		name string
		args args
		want *MigrateUtil
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMigrateUtil(tt.args.cfg, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMigrateUtil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMigrateUtil_Migrate(t *testing.T) {
	mcfg, _:=GetMigrateConfig()
	type fields struct {
		config *MigrateConfig
		logger *zap.SugaredLogger
	}
	type args struct {
		sourceLocation string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantVersion uint
		wantErr     bool
	}{
		{
			name: "test-all",
			wantErr:false,
			args:args{sourceLocation:"file:///Users/nishant/go/src/devtron.ai/orchestrator/scripts/sql"},
			fields:fields{config:mcfg,logger:nil},
			wantVersion:57,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util := MigrateUtil{
				config: tt.fields.config,
				logger: tt.fields.logger,
			}
			gotVersion, err := util.Migrate(tt.args.sourceLocation)
			if (err != nil) != tt.wantErr {
				t.Errorf("MigrateUtil.Migrate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("MigrateUtil.Migrate() = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

func TestMigrateUtil_migrateUp(t *testing.T) {
	type fields struct {
		config *MigrateConfig
		logger *zap.SugaredLogger
	}
	type args struct {
		location string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantVersion uint
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util := MigrateUtil{
				config: tt.fields.config,
				logger: tt.fields.logger,
			}
			gotVersion, err := util.migrateUp(tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("MigrateUtil.migrateUp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("MigrateUtil.migrateUp() = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

func TestMigrateUtil_migrateGoTo(t *testing.T) {
	type fields struct {
		config *MigrateConfig
		logger *zap.SugaredLogger
	}
	type args struct {
		location    string
		goToVersion uint
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantVersion uint
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util := MigrateUtil{
				config: tt.fields.config,
				logger: tt.fields.logger,
			}
			gotVersion, err := util.migrateGoTo(tt.args.location, tt.args.goToVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("MigrateUtil.migrateGoTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("MigrateUtil.migrateGoTo() = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}
