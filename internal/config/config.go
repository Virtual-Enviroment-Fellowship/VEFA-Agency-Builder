package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	AgencyName     string `json:"agency_name"`
	VPSHost        string `json:"vps_host"`
	SSHPort        int    `json:"ssh_port"`
	SSHUser        string `json:"ssh_user"`
	SSHPassword    string `json:"ssh_password,omitempty"`
	SSHKeyPath     string `json:"ssh_key_path,omitempty"`
	PortalDomain   string `json:"portal_domain"`
	AssetDomain    string `json:"asset_domain"`
	DBUser         string `json:"db_user"`
	DBPassword     string `json:"db_password"`
	CoreDBName     string `json:"core_db_name"`
	DemoDBName     string `json:"demo_db_name"`
	DemoDomain     string `json:"demo_domain"`
	DemoTenantID   string `json:"demo_tenant_id"`
	MCPEndpoint    string `json:"mcp_endpoint"`
	MCPSecretKey   string `json:"mcp_secret_key"`
	SystemUser     string `json:"system_user"`
	SeedDemoTenant bool   `json:"seed_demo_tenant"`
}

func NewDefaultConfig() *Config {
	return &Config{
		AgencyName:     "VEFA Agency",
		VPSHost:        "",
		SSHPort:        22,
		SSHUser:        "root",
		PortalDomain:   "portal.youragency.com",
		AssetDomain:    "assets.youragency.com",
		DBUser:         "db_core_user",
		DBPassword:     "ChangeMeSecurePassword123!",
		CoreDBName:     "db_agency_core",
		DemoDBName:     "db_tenant_demo",
		DemoDomain:     "demo.youragency.com",
		DemoTenantID:   "demo-tenant-101",
		MCPEndpoint:    "https://api.consultdevin.com/mcp",
		MCPSecretKey:   "sk_agency_secret_mcp_token_9981",
		SystemUser:     "clp",
		SeedDemoTenant: true,
	}
}

func (c *Config) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0600)
}

func LoadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
