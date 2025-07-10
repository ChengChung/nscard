package utils

import "slices"

type AdminConfig struct {
	AdminUsers []string `yaml:"admin_users"`
}

var cfg *AdminConfig

func IsAdmin(user string) bool {
	if cfg == nil {
		return false
	}
	return slices.Contains(cfg.AdminUsers, user)
}

func InitAdminConfig(adminCfg AdminConfig) {
	cfg = &adminCfg
}
