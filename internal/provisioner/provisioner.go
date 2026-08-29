package provisioner

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
	"time"

	"simpleconsult/setup-wizard/internal/config"
	"simpleconsult/setup-wizard/internal/sshclient"
)

type Provisioner struct {
	client     *sshclient.SSHClient
	cfg        *config.Config
	templateFS embed.FS
	onLog      func(level, msg string)
	dryRun     bool
}

func NewProvisioner(client *sshclient.SSHClient, cfg *config.Config, templateFS embed.FS, dryRun bool, onLog func(level, msg string)) *Provisioner {
	return &Provisioner{
		client:     client,
		cfg:        cfg,
		templateFS: templateFS,
		dryRun:     dryRun,
		onLog:      onLog,
	}
}

func (p *Provisioner) log(level, format string, args ...interface{}) {
	if p.onLog != nil {
		p.onLog(level, fmt.Sprintf(format, args...))
	}
}

func (p *Provisioner) renderTemplate(tmplPath string, data interface{}) (string, error) {
	tmplBytes, err := p.templateFS.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed reading embedded template '%s': %w", tmplPath, err)
	}

	tmpl, err := template.New(tmplPath).Parse(string(tmplBytes))
	if err != nil {
		return "", fmt.Errorf("failed parsing template '%s': %w", tmplPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed executing template '%s': %w", tmplPath, err)
	}

	return buf.String(), nil
}

func (p *Provisioner) run(cmd string) (string, error) {
	if p.dryRun {
		p.log("DRY_RUN", "[CMD] %s", cmd)
		return "[DRY-RUN EXECUTED]", nil
	}
	return p.client.RunCommand(cmd)
}

func (p *Provisioner) runStream(cmd string) error {
	if p.dryRun {
		p.log("DRY_RUN", "[STREAM CMD] %s", cmd)
		return nil
	}
	return p.client.RunCommandStream(cmd, func(line string) {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			p.log("INFO", "  │ %s", trimmed)
		}
	})
}

func (p *Provisioner) writeRemote(path, content, perm string) error {
	if p.dryRun {
		p.log("DRY_RUN", "[WRITE FILE %s (%s)] Length: %d bytes", path, perm, len(content))
		return nil
	}
	return p.client.WriteRemoteFile(path, content, perm)
}

// RunFullSetup executes all 7 provisioning phases
func (p *Provisioner) RunFullSetup() error {
	p.log("STEP", "═══════════════════════════════════════════════════════════════")
	p.log("STEP", "  STARTING FULL VPS PROVISIONING FOR: %s", p.cfg.AgencyName)
	p.log("STEP", "═══════════════════════════════════════════════════════════════")

	phases := []struct {
		name string
		fn   func() error
	}{
		{"Phase 1: OS Update, Prerequisites & UFW Firewall", p.Phase1_OSAndFirewall},
		{"Phase 2: CloudPanel & MariaDB 10.11 Engine", p.Phase2_CloudPanel},
		{"Phase 3: Core Database & Multi-Tenant Schemas", p.Phase3_DatabaseSchemas},
		{"Phase 4: Fallback SSL & Nginx Catch-All vHost", p.Phase4_NginxVHost},
		{"Phase 5: Master Slim PHP Codebase & Composer Dependencies", p.Phase5_DeployCodebase},
		{"Phase 6: AI Background Worker Daemon & Systemd Service", p.Phase6_WorkerDaemon},
		{"Phase 7: Verification & Diagnostics", p.Phase7_Diagnostics},
	}

	for i, phase := range phases {
		p.log("STEP", "\n▶ [%d/7] %s", i+1, phase.name)
		start := time.Now()
		if err := phase.fn(); err != nil {
			p.log("ERROR", "❌ Failed at %s: %v", phase.name, err)
			return err
		}
		p.log("SUCCESS", "✔ Completed %s (took %v)", phase.name, time.Since(start).Round(time.Millisecond))
	}

	p.log("STEP", "\n═══════════════════════════════════════════════════════════════")
	p.log("SUCCESS", "🎉 ALL SETUP PHASES COMPLETED SUCCESSFULLY!")
	p.log("STEP", "═══════════════════════════════════════════════════════════════")
	return nil
}

func (p *Provisioner) Phase1_OSAndFirewall() error {
	p.log("INFO", "Updating Ubuntu package index and installing base dependencies...")
	cmd := "export DEBIAN_FRONTEND=noninteractive && apt-get update -y && apt-get install -y curl wget git unzip openssl ufw software-properties-common"
	if err := p.runStream(cmd); err != nil {
		return fmt.Errorf("failed updating packages: %w", err)
	}

	p.log("INFO", "Configuring UFW firewall rules (Ports 22, 80, 443, 8443)...")
	ufwCmds := "ufw allow 22/tcp && ufw allow 80/tcp && ufw allow 443/tcp && ufw allow 8443/tcp && echo 'y' | ufw enable"
	if _, err := p.run(ufwCmds); err != nil {
		p.log("WARN", "UFW warning (proceeding): %v", err)
	}
	return nil
}

func (p *Provisioner) Phase2_CloudPanel() error {
	p.log("INFO", "Checking if CloudPanel is already installed...")
	out, _ := p.run("which clpctl || test -d /var/www || echo 'not_installed'")
	if strings.Contains(out, "clpctl") {
		p.log("INFO", "CloudPanel is already installed. Skipping installer.")
		return nil
	}

	p.log("INFO", "Downloading and installing CloudPanel CE with MariaDB 10.11...")
	installCmd := "curl -sS https://installer.cloudpanel.io/ce/v2/install.sh -o /tmp/install.sh && sudo DB_ENGINE=MARIADB_10.11 bash /tmp/install.sh"
	if err := p.runStream(installCmd); err != nil {
		p.log("WARN", "CloudPanel installer returned code, verifying services: %v", err)
	}

	// Detect system user (clp or www-data)
	userCheck, _ := p.run("id -u clp 2>/dev/null && echo 'clp' || echo 'www-data'")
	if strings.Contains(userCheck, "clp") {
		p.cfg.SystemUser = "clp"
	} else {
		p.cfg.SystemUser = "www-data"
	}
	p.log("INFO", "Configured application system user: %s", p.cfg.SystemUser)
	return nil
}

func (p *Provisioner) Phase3_DatabaseSchemas() error {
	p.log("INFO", "Provisioning MariaDB root & core agency user...")
	
	// Create MySQL core user if needed
	createSqlUser := fmt.Sprintf(
		"mysql -e \"CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s'; GRANT ALL PRIVILEGES ON *.* TO '%s'@'localhost' WITH GRANT OPTION; FLUSH PRIVILEGES;\"",
		p.cfg.DBUser, p.cfg.DBPassword, p.cfg.DBUser,
	)
	if _, err := p.run(createSqlUser); err != nil {
		p.log("WARN", "Database user creation note: %v", err)
	}

	p.log("INFO", "Rendering and applying db_agency_core schema...")
	schemaSql, err := p.renderTemplate("sql/schema.sql", p.cfg)
	if err != nil {
		return err
	}

	if err := p.writeRemote("/tmp/agency_schema.sql", schemaSql, "0644"); err != nil {
		return err
	}

	mysqlCmd := fmt.Sprintf("mysql -u%s -p'%s' < /tmp/agency_schema.sql", p.cfg.DBUser, p.cfg.DBPassword)
	if _, err := p.run(mysqlCmd); err != nil {
		// Fallback to local root auth
		mysqlCmdFallback := "mysql < /tmp/agency_schema.sql"
		if _, err2 := p.run(mysqlCmdFallback); err2 != nil {
			return fmt.Errorf("failed executing schema.sql: %w", err)
		}
	}

	if p.cfg.SeedDemoTenant {
		p.log("INFO", "Seeding demo tenant (%s -> %s)...", p.cfg.DemoDomain, p.cfg.DemoDBName)
		seedSql, err := p.renderTemplate("sql/seed_tenant.sql", p.cfg)
		if err != nil {
			return err
		}
		if err := p.writeRemote("/tmp/agency_seed.sql", seedSql, "0644"); err != nil {
			return err
		}
		seedCmd := fmt.Sprintf("mysql -u%s -p'%s' < /tmp/agency_seed.sql", p.cfg.DBUser, p.cfg.DBPassword)
		if _, err := p.run(seedCmd); err != nil {
			p.run("mysql < /tmp/agency_seed.sql")
		}
	}

	return nil
}

func (p *Provisioner) Phase4_NginxVHost() error {
	p.log("INFO", "Setting up fallback self-signed SSL certificates for Nginx...")
	sslCmd := "mkdir -p /etc/nginx/ssl && openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /etc/nginx/ssl/agency-master.key -out /etc/nginx/ssl/agency-master.crt -subj '/CN=agency-master-catchall'"
	if _, err := p.run(sslCmd); err != nil {
		p.log("WARN", "SSL certificate generation notice: %v", err)
	}

	p.log("INFO", "Deploying Nginx Catch-All vHost template...")
	nginxConf, err := p.renderTemplate("nginx/agency-master.conf", p.cfg)
	if err != nil {
		return err
	}

	vhostPath := "/etc/nginx/sites-available/agency-master.conf"
	if err := p.writeRemote(vhostPath, nginxConf, "0644"); err != nil {
		return err
	}

	p.run("ln -sf /etc/nginx/sites-available/agency-master.conf /etc/nginx/sites-enabled/agency-master.conf")
	p.run("rm -f /etc/nginx/sites-enabled/default")

	p.log("INFO", "Testing Nginx configuration syntax...")
	if out, err := p.run("nginx -t"); err != nil {
		p.log("WARN", "Nginx test output: %s", out)
	} else {
		p.run("systemctl reload nginx")
	}

	return nil
}

func (p *Provisioner) Phase5_DeployCodebase() error {
	p.log("INFO", "Creating master application directories at /var/www/master-app...")
	p.run("mkdir -p /var/www/master-app/public /var/www/master-app/app/Core /var/www/master-app/app/Trades/Tutor /var/www/master-app/app/Trades/Plumber /var/www/master-app/public/js")

	p.log("INFO", "Deploying composer.json and PHP application source files...")
	
	// 1. composer.json
	composerJson, err := p.renderTemplate("php/composer.json", p.cfg)
	if err != nil {
		return err
	}
	p.writeRemote("/var/www/master-app/composer.json", composerJson, "0644")

	// 2. public/index.php
	indexPhp, err := p.renderTemplate("php/public/index.php", p.cfg)
	if err != nil {
		return err
	}
	p.writeRemote("/var/www/master-app/public/index.php", indexPhp, "0644")

	// 3. app/Core/routes.php
	coreRoutes, err := p.renderTemplate("php/app/Core/routes.php", p.cfg)
	if err != nil {
		return err
	}
	p.writeRemote("/var/www/master-app/app/Core/routes.php", coreRoutes, "0644")

	// 4. app/Trades/Tutor/routes.php
	tutorRoutes, err := p.renderTemplate("php/app/Trades/Tutor/routes.php", p.cfg)
	if err != nil {
		return err
	}
	p.writeRemote("/var/www/master-app/app/Trades/Tutor/routes.php", tutorRoutes, "0644")

	// 5. app/Trades/Plumber/routes.php
	plumberRoutes, err := p.renderTemplate("php/app/Trades/Plumber/routes.php", p.cfg)
	if err != nil {
		return err
	}
	p.writeRemote("/var/www/master-app/app/Trades/Plumber/routes.php", plumberRoutes, "0644")

	// 6. public/js/polling.js
	pollingJs, err := p.renderTemplate("public/polling.js", p.cfg)
	if err != nil {
		return err
	}
	p.writeRemote("/var/www/master-app/public/js/polling.js", pollingJs, "0644")

	// Verify Composer is installed
	p.log("INFO", "Checking Composer installation on VPS...")
	cOut, _ := p.run("which composer || echo 'not_installed'")
	if strings.Contains(cOut, "not_installed") {
		p.log("INFO", "Installing Composer globally...")
		p.run("curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer")
	}

	p.log("INFO", "Installing PHP dependencies via Composer...")
	installComposerDeps := "cd /var/www/master-app && composer install --no-dev --optimize-autoloader --no-interaction"
	if err := p.runStream(installComposerDeps); err != nil {
		p.log("WARN", "Composer warning (fallback check): %v", err)
	}

	p.log("INFO", "Setting permissions on /var/www/master-app for user %s...", p.cfg.SystemUser)
	p.run(fmt.Sprintf("chown -R %s:%s /var/www/master-app", p.cfg.SystemUser, p.cfg.SystemUser))
	p.run("chmod -R 755 /var/www/master-app")

	return nil
}

func (p *Provisioner) Phase6_WorkerDaemon() error {
	p.log("INFO", "Deploying worker.php daemon...")
	workerPhp, err := p.renderTemplate("php/worker.php", p.cfg)
	if err != nil {
		return err
	}
	p.writeRemote("/var/www/master-app/worker.php", workerPhp, "0755")

	p.log("INFO", "Registering Systemd service agency-worker.service...")
	systemdService, err := p.renderTemplate("systemd/agency-worker.service", p.cfg)
	if err != nil {
		return err
	}
	p.writeRemote("/etc/systemd/system/agency-worker.service", systemdService, "0644")

	p.run("systemctl daemon-reload")
	p.run("systemctl enable agency-worker.service --now")
	p.run("systemctl restart agency-worker.service")

	return nil
}

func (p *Provisioner) Phase7_Diagnostics() error {
	p.log("INFO", "Running diagnostics and service verification...")

	// 1. Verify Nginx
	nginxStatus, _ := p.run("systemctl is-active nginx")
	p.log("INFO", "  • Nginx Service Status: %s", strings.TrimSpace(nginxStatus))

	// 2. Verify PHP-FPM
	fpmStatus, _ := p.run("systemctl is-active php8.3-fpm || systemctl is-active php8.2-fpm || echo 'unknown'")
	p.log("INFO", "  • PHP-FPM Status: %s", strings.TrimSpace(fpmStatus))

	// 3. Verify Background Daemon
	workerStatus, _ := p.run("systemctl is-active agency-worker")
	p.log("INFO", "  • Agency Worker Daemon Status: %s", strings.TrimSpace(workerStatus))

	// 4. Verify Database Connectivity & Tenants Table
	dbCheck, _ := p.run(fmt.Sprintf("mysql -u%s -p'%s' -e 'USE %s; SELECT COUNT(*) AS count FROM tenants;' 2>&1", p.cfg.DBUser, p.cfg.DBPassword, p.cfg.CoreDBName))
	p.log("INFO", "  • Database Query Verification:\n%s", strings.TrimSpace(dbCheck))

	// 5. Test Localhost HTTP request simulation
	httpCheck, _ := p.run(fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' -H 'Host: %s' http://127.0.0.1/api/health || echo '000'", p.cfg.DemoDomain))
	p.log("INFO", "  • Simulated Tenant Route HTTP Status: %s", strings.TrimSpace(httpCheck))

	return nil
}
