package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"simpleconsult/setup-wizard/internal/config"
	"simpleconsult/setup-wizard/internal/provisioner"
	"simpleconsult/setup-wizard/internal/sshclient"
	"simpleconsult/setup-wizard/templates"
)

// ANSI Color Codes
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	BgBlue  = "\033[44m"
	BgGreen = "\033[42m"
	BgRed   = "\033[41m"
)

const BuildVersion = "VEFAagencyBuild v0.1.2"

func printBanner() {
	fmt.Print(Cyan + Bold)
	fmt.Println(`
╔═════════════════════════════════════════════════════════════════════════════╗
║                                                                             ║
║                     ██╗   ██╗███████╗███████╗ █████╗                        ║
║                     ██║   ██║██╔════╝██╔════╝██╔══██╗                       ║
║                     ██║   ██║█████╗  █████╗  ███████║                       ║
║                     ╚██╗ ██╔╝██╔══╝  ██╔══╝  ██╔══██║                       ║
║                      ╚████╔╝ ███████╗██║     ██║  ██║                       ║
║                       ╚═══╝  ╚══════╝╚═╝     ╚═╝  ╚═╝                       ║
║                                                                             ║
║             █████╗  ██████╗ ███████╗███╗   ██╗ ██████╗██╗   ██╗             ║
║            ██╔══██╗██╔════╝ ██╔════╝████╗  ██║██╔════╝╚██╗ ██╔╝             ║
║            ███████║██║  ███╗█████╗  ██╔██╗ ██║██║      ╚████╔╝              ║
║            ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║██║       ╚██╔╝               ║
║            ██║  ██║╚██████╔╝███████╗██║ ╚████║╚██████╗   ██║                ║
║            ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝ ╚═════╝   ╚═╝                ║
║                                                                             ║
║            ██████╗ ██╗   ██╗██╗██╗     ██████╗ ███████╗██████╗              ║
║            ██╔══██╗██║   ██║██║██║     ██╔══██╗██╔════╝██╔══██╗             ║
║            ██████╔╝██║   ██║██║██║     ██║  ██║█████╗  ██████╔╝             ║
║            ██╔══██╗██║   ██║██║██║     ██║  ██║██╔══╝  ██╔══██╗             ║
║            ██████╔╝╚██████╔╝██║███████╗██████╔╝███████╗██║  ██║             ║
║            ╚═════╝  ╚═════╝ ╚═╝╚══════╝╚═════╝ ╚══════╝╚═╝  ╚═╝             ║
║                                                                             ║
║        VEFAagencyBuild v0.1.2  •  Multi-Tenant AI VPS Setup Wizard          ║
║                      Target OS: Ubuntu 24.04 LTS                            ║
╚═════════════════════════════════════════════════════════════════════════════╝` + Reset)
	fmt.Println()
}

func printCelebrationBanner(cfg *config.Config) {
	fmt.Print(Green + Bold)
	fmt.Println(`
╔═════════════════════════════════════════════════════════════════════════════╗
║                                                                             ║
║                     ██╗   ██╗███████╗███████╗ █████╗                        ║
║                     ██║   ██║██╔════╝██╔════╝██╔══██╗                       ║
║                     ██║   ██║█████╗  █████╗  ███████║                       ║
║                     ╚██╗ ██╔╝██╔══╝  ██╔══╝  ██╔══██║                       ║
║                      ╚████╔╝ ███████╗██║     ██║  ██║                       ║
║                       ╚═══╝  ╚══════╝╚═╝     ╚═╝  ╚═╝                       ║
║                                                                             ║
║             🎉 CONGRATULATIONS! SETUP COMPLETED SUCCESSFULLY!               ║
║                                                                             ║
║         VEFA™: Virtual Environment Fellowship of America                    ║
║         Official Community & Standards: https://www.VEFA.club               ║
║         (VEFA Trademarked as of 2026 - All Rights Reserved)                 ║
║                                                                             ║
║               Your High-Density AI VPS Infrastructure is LIVE               ║
╚═════════════════════════════════════════════════════════════════════════════╝` + Reset)
	fmt.Println()
}

func printTooltip(title, purpose, location, format, example string) {
	fmt.Println(Dim + "┌── " + Yellow + Bold + "💡 [?] TOOLTIP: " + title + Reset + Dim)
	fmt.Printf("│  %sPurpose:%s  %s\n", White+Bold, Reset+Dim, purpose)
	if location != "" {
		fmt.Printf("│  %sWhere to find:%s  %s\n", Cyan+Bold, Reset+Dim, location)
	}
	if format != "" {
		fmt.Printf("│  %sFormat/Default:%s %s\n", Green+Bold, Reset+Dim, format)
	}
	if example != "" {
		fmt.Printf("│  %sExample:%s %s\n", Magenta+Bold, Reset+Dim, example)
	}
	fmt.Println("└────────────────────────────────────────────────────────────────────────" + Reset)
}

func promptString(reader *bufio.Reader, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s%s%s [%s%s%s]: ", Bold, label, Reset, Green, defaultVal, Reset)
	} else {
		fmt.Printf("%s%s%s: ", Bold, label, Reset)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

func promptPassword(label string) string {
	fmt.Printf("%s%s%s: ", Bold, label, Reset)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(bytePassword))
}

func promptBool(reader *bufio.Reader, label string, defaultVal bool) bool {
	defStr := "Y/n"
	if !defaultVal {
		defStr = "y/N"
	}
	fmt.Printf("%s%s%s [%s%s%s]: ", Bold, label, Reset, Green, defStr, Reset)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultVal
	}
	return input == "y" || input == "yes"
}

func pauseExit(code int) {
	fmt.Printf("\n%s%sPress [Enter] to close VEFA Agency Builder...%s\n", Yellow, Bold, Reset)
	bufio.NewReader(os.Stdin).ReadString('\n')
	os.Exit(code)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("%sCould not open browser automatically: %v%s\n", Yellow, err, Reset)
	} else {
		fmt.Printf("%s✔ Browser opened to: %s%s\n", Green, url, Reset)
	}
}

func runInteractiveQuestionnaire(cfg *config.Config, configFile string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(BgBlue + White + Bold + " STEP 1 of 5: AGENCY & BRANDING " + Reset)
	fmt.Println()

	printTooltip(
		"Agency Name",
		"Identifies your agency instance across the admin dashboard and notifications.",
		"Internal branding / Agency Portal identifier",
		"Text (e.g. VEFA Agency)",
		"VEFA Agency Global",
	)
	cfg.AgencyName = promptString(reader, "Enter Agency Name", cfg.AgencyName)
	cfg.SaveToFile(configFile)
	fmt.Println()

	fmt.Println(BgBlue + White + Bold + " STEP 2 of 5: HOSTINGER VPS & SSH CREDENTIALS " + Reset)
	fmt.Println()

	printTooltip(
		"Hostinger VPS IP Address",
		"The public IPv4 address of your Ubuntu 24.04 Virtual Private Server.",
		"Hostinger hPanel → VPS Management → Server Details → IP Address",
		"Valid IPv4 address (e.g. 195.35.40.12)",
		"185.199.108.153",
	)
	for {
		cfg.VPSHost = promptString(reader, "Enter Hostinger VPS IP Address", cfg.VPSHost)
		if cfg.VPSHost != "" {
			break
		}
		fmt.Println(Red + "  ⚠ VPS IP address cannot be empty." + Reset)
	}
	cfg.SaveToFile(configFile)
	fmt.Println()

	printTooltip(
		"SSH Port",
		"The port your VPS SSH service listens on. Standard default is 22.",
		"Hostinger hPanel → VPS Management → Access Details",
		"Integer (Default: 22)",
		"22",
	)
	portStr := promptString(reader, "Enter SSH Port", strconv.Itoa(cfg.SSHPort))
	if p, err := strconv.Atoi(portStr); err == nil {
		cfg.SSHPort = p
	}
	cfg.SaveToFile(configFile)
	fmt.Println()

	printTooltip(
		"SSH Username",
		"The administrative login user on your Ubuntu VPS. 'root' is standard for fresh servers.",
		"Hostinger hPanel → VPS Access Details",
		"root (or sudo user like 'ubuntu')",
		"root",
	)
	cfg.SSHUser = promptString(reader, "Enter SSH Username", cfg.SSHUser)
	cfg.SaveToFile(configFile)
	fmt.Println()

	printTooltip(
		"SSH Authentication Method",
		"Choose whether to authenticate with your root password or an SSH private key file.",
		"Hostinger hPanel → Set Root Password, or your ~/.ssh/id_rsa key",
		"P for Password / K for Key file",
		"P",
	)
	authChoice := promptString(reader, "Authentication method (P = Password, K = Private Key)", "P")
	if strings.ToUpper(authChoice) == "K" {
		printTooltip(
			"SSH Private Key Path",
			"Local path on this PC to your private SSH key.",
			"e.g. C:\\Users\\admin\\.ssh\\id_rsa",
			"Valid file path",
			"C:\\Users\\admin\\.ssh\\id_ed25519",
		)
		cfg.SSHKeyPath = promptString(reader, "Enter Private Key Path", cfg.SSHKeyPath)
		cfg.SSHPassword = ""
	} else {
		printTooltip(
			"SSH Root Password",
			"The root password for your VPS (characters will be hidden for security).",
			"Hostinger hPanel → VPS Management → SSH Password",
			"Secret text (hidden input)",
			"********",
		)
		cfg.SSHPassword = promptPassword("Enter SSH Root Password")
		cfg.SSHKeyPath = ""
	}
	cfg.SaveToFile(configFile)
	fmt.Println()

	fmt.Println(BgBlue + White + Bold + " STEP 3 of 5: DOMAIN NAMES " + Reset)
	fmt.Println()

	printTooltip(
		"Central Agency Portal Domain",
		"The primary domain for your agency's central IdP portal and admin dashboard.",
		"Your DNS Manager (Point A-Record to VPS IP)",
		"FQDN (Fully Qualified Domain Name)",
		"portal.youragency.com",
	)
	cfg.PortalDomain = promptString(reader, "Enter Portal Domain", cfg.PortalDomain)
	cfg.SaveToFile(configFile)
	fmt.Println()

	printTooltip(
		"Master Asset Domain",
		"Single domain serving global CSS, JS libraries, and logos cached across all 200 sites.",
		"Your DNS Manager (Point A-Record to VPS IP)",
		"FQDN",
		"assets.youragency.com",
	)
	cfg.AssetDomain = promptString(reader, "Enter Master Asset Domain", cfg.AssetDomain)
	cfg.SaveToFile(configFile)
	fmt.Println()

	fmt.Println(BgBlue + White + Bold + " STEP 4 of 5: DATABASE & MCP AI ORCHESTRATOR " + Reset)
	fmt.Println()

	printTooltip(
		"MariaDB Database Password",
		"The password for the MariaDB core agency user (db_core_user).",
		"Generated secure password for MySQL / MariaDB",
		"Alphanumeric + Special Characters",
		"SecretAgencyDbPass_2026!",
	)
	cfg.DBPassword = promptString(reader, "Enter MariaDB Core Password", cfg.DBPassword)
	cfg.SaveToFile(configFile)
	fmt.Println()

	printTooltip(
		"External MCP API Endpoint",
		"The Model Context Protocol orchestrator that handles asynchronous AI tasks.",
		"Your AI orchestrator gateway URL",
		"HTTPS URL",
		"https://api.consultdevin.com/mcp",
	)
	cfg.MCPEndpoint = promptString(reader, "Enter MCP API Endpoint", cfg.MCPEndpoint)
	cfg.SaveToFile(configFile)
	fmt.Println()

	printTooltip(
		"MCP Master Secret Key",
		"Bearer token used by the background daemon and incoming webhooks to verify authenticity.",
		"Generated API token / MCP API Key",
		"Secret Token String",
		"sk_live_mcp_secret_token_abc123",
	)
	cfg.MCPSecretKey = promptString(reader, "Enter MCP Secret Key", cfg.MCPSecretKey)
	cfg.SaveToFile(configFile)
	fmt.Println()

	fmt.Println(BgBlue + White + Bold + " STEP 5 of 5: SAMPLE TENANT SEEDING " + Reset)
	fmt.Println()

	printTooltip(
		"Seed Demo Tenant",
		"Automatically registers a demo tenant (demo.youragency.com) with the Tutor trade module for immediate testing.",
		"Sample Database Seed",
		"Yes / No",
		"Yes",
	)
	cfg.SeedDemoTenant = promptBool(reader, "Seed demo tenant ('demo.youragency.com')?", cfg.SeedDemoTenant)
	if cfg.SeedDemoTenant {
		cfg.DemoDomain = promptString(reader, "Demo Tenant Domain", cfg.DemoDomain)
	}
	cfg.SaveToFile(configFile)
	fmt.Println()
}

func connectWithInteractiveRecovery(cfg *config.Config, configFile string) (*sshclient.SSHClient, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s▶ Attempting SSH connection to %s@%s:%d...%s\n", Cyan+Bold, cfg.SSHUser, cfg.VPSHost, cfg.SSHPort, Reset)

		client, err := sshclient.Connect(cfg.VPSHost, cfg.SSHPort, cfg.SSHUser, cfg.SSHPassword, cfg.SSHKeyPath)
		if err == nil {
			fmt.Printf("%s✔ SSH Connection Established Successfully!%s\n\n", Green+Bold, Reset)
			cfg.SaveToFile(configFile)
			return client, nil
		}

		errMsg := err.Error()
		fmt.Println()
		fmt.Println(Red + Bold + "┌────────────────────────────────────────────────────────────────────────" + Reset)
		fmt.Printf("%s│  ❌ SSH Connection Failed: %s%s\n", Red+Bold, White+Bold, errMsg)
		fmt.Println(Red + Bold + "├────────────────────────────────────────────────────────────────────────" + Reset)
		fmt.Printf("%s│  Target:   %s%s@%s:%d\n", White, Cyan, cfg.SSHUser, cfg.VPSHost, cfg.SSHPort)
		if cfg.SSHKeyPath != "" {
			fmt.Printf("%s│  Auth:     %sPrivate Key (%s)\n", White, Cyan, cfg.SSHKeyPath)
		} else {
			fmt.Printf("%s│  Auth:     %sPassword Authentication\n", White, Cyan)
		}

		if strings.Contains(strings.ToLower(errMsg), "handshake") || strings.Contains(strings.ToLower(errMsg), "unable to authenticate") {
			fmt.Printf("%s│  Diagnostic: %sAuthentication credentials were rejected by the server.\n", Yellow+Bold, Reset)
			fmt.Printf("%s│              Check your root password or SSH key file in Hostinger hPanel.\n", Reset)
		} else if strings.Contains(strings.ToLower(errMsg), "timeout") || strings.Contains(strings.ToLower(errMsg), "refused") || strings.Contains(strings.ToLower(errMsg), "no route") {
			fmt.Printf("%s│  Diagnostic: %sNetwork connection could not be established.\n", Yellow+Bold, Reset)
			fmt.Printf("%s│              Verify VPS IP address and check that your VPS is powered ON.\n", Reset)
		}
		fmt.Println(Red + Bold + "└────────────────────────────────────────────────────────────────────────" + Reset)
		fmt.Println()

		fmt.Println(Yellow + Bold + "RECOVERY OPTIONS (Settings are safely preserved):" + Reset)
		fmt.Println("  [1] 🔁 Retry connection (Try again with current settings)")
		fmt.Printf("  [2] 🔑 Update SSH Password\n")
		fmt.Printf("  [3] 🌐 Update VPS IP Address (currently: %s%s%s)\n", Cyan, cfg.VPSHost, Reset)
		fmt.Printf("  [4] 👤 Update SSH Username & Port (currently: %s%s:%d%s)\n", Cyan, cfg.SSHUser, cfg.SSHPort, Reset)
		fmt.Println("  [5] 📁 Switch to SSH Private Key file")
		fmt.Println("  [6] 📝 Re-run guided setup questionnaire")
		fmt.Println("  [7] 🚪 Save and Exit")
		fmt.Println()

		choice := promptString(reader, "Select recovery option (1-7)", "1")

		switch strings.TrimSpace(choice) {
		case "1":
			continue
		case "2":
			printTooltip("SSH Root Password", "Enter the correct root password for your VPS.", "Hostinger hPanel → Access Details", "Secret text", "********")
			cfg.SSHPassword = promptPassword("Enter SSH Root Password")
			cfg.SSHKeyPath = ""
			cfg.SaveToFile(configFile)
			fmt.Printf("%s✔ Password updated and saved.%s\n\n", Green, Reset)
		case "3":
			printTooltip("VPS IP Address", "Public IPv4 address of your Hostinger VPS.", "Hostinger hPanel → Server Details", "IPv4 address", "185.199.108.153")
			cfg.VPSHost = promptString(reader, "Enter VPS IP Address", cfg.VPSHost)
			cfg.SaveToFile(configFile)
			fmt.Printf("%s✔ IP Address updated and saved.%s\n\n", Green, Reset)
		case "4":
			cfg.SSHUser = promptString(reader, "Enter SSH Username", cfg.SSHUser)
			portStr := promptString(reader, "Enter SSH Port", strconv.Itoa(cfg.SSHPort))
			if p, err := strconv.Atoi(portStr); err == nil {
				cfg.SSHPort = p
			}
			cfg.SaveToFile(configFile)
			fmt.Printf("%s✔ SSH User and Port updated and saved.%s\n\n", Green, Reset)
		case "5":
			printTooltip("SSH Private Key Path", "Path to private key on your PC.", "e.g. C:\\Users\\admin\\.ssh\\id_rsa", "File path", "C:\\Users\\admin\\.ssh\\id_ed25519")
			cfg.SSHKeyPath = promptString(reader, "Enter Private Key Path", cfg.SSHKeyPath)
			cfg.SSHPassword = ""
			cfg.SaveToFile(configFile)
			fmt.Printf("%s✔ Private key path updated and saved.%s\n\n", Green, Reset)
		case "6":
			runInteractiveQuestionnaire(cfg, configFile)
		case "7":
			cfg.SaveToFile(configFile)
			fmt.Printf("%sConfiguration preserved in '%s'. Exiting.%s\n", Yellow, configFile, Reset)
			return nil, fmt.Errorf("setup aborted by user")
		default:
			fmt.Println(Yellow + "Invalid choice. Retrying connection..." + Reset)
		}
	}
}

// runLiveDiagnostics performs interactive 4-point verification
func runLiveDiagnostics(client *sshclient.SSHClient, cfg *config.Config) {
	fmt.Println()
	fmt.Println(Cyan + Bold + "╔═════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║               🧪 LIVE SYSTEM DIAGNOSTIC & VERIFICATION SUITE                ║")
	fmt.Println("╚═════════════════════════════════════════════════════════════════════════════╝" + Reset)
	fmt.Println()

	if client == nil {
		fmt.Println(Yellow + "[!] Running in dry-run mode: Mock diagnostics passed." + Reset)
		return
	}

	// 1. Nginx & FPM Pool Check
	fmt.Print("  [1/4] Checking Nginx & Single PHP-FPM Worker Pool... ")
	out, err := client.RunCommand("systemctl is-active nginx && (systemctl is-active php8.3-fpm || systemctl is-active php8.2-fpm)")
	if err == nil && strings.Contains(out, "active") {
		fmt.Println(Green + Bold + "✔ PASS (Nginx & PHP-FPM Online)" + Reset)
	} else {
		fmt.Printf("%s⚠ WARNING: %s%s\n", Yellow, strings.TrimSpace(out), Reset)
	}

	// 2. Database Integrity Check
	fmt.Print("  [2/4] Verifying MariaDB & Core Tables (tenants, ai_job_queue)... ")
	dbCheckCmd := fmt.Sprintf("mysql -u%s -p'%s' -e 'USE %s; SELECT COUNT(*) FROM tenants; SELECT COUNT(*) FROM ai_job_queue;' 2>&1", cfg.DBUser, cfg.DBPassword, cfg.CoreDBName)
	dbOut, dbErr := client.RunCommand(dbCheckCmd)
	if dbErr == nil {
		fmt.Println(Green + Bold + "✔ PASS (Database Connected & Tables Ready)" + Reset)
	} else {
		fmt.Printf("%s⚠ WARNING: %s%s\n", Yellow, strings.TrimSpace(dbOut), Reset)
	}

	// 3. Multi-Tenant Slim PHP Front Controller Router Test
	fmt.Printf("  [3/4] Testing Dynamic Domain Routing (Host: %s)... ", cfg.DemoDomain)
	httpCmd := fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' -H 'Host: %s' http://127.0.0.1/api/health || echo '000'", cfg.DemoDomain)
	httpOut, _ := client.RunCommand(httpCmd)
	httpCode := strings.TrimSpace(httpOut)
	if httpCode == "200" || httpCode == "404" {
		fmt.Printf("%s✔ PASS (Slim Router Responding with HTTP %s)%s\n", Green+Bold, httpCode, Reset)
	} else {
		fmt.Printf("%s⚠ Response HTTP %s%s\n", Yellow, httpCode, Reset)
	}

	// 4. Background AI Daemon Loop Verification
	fmt.Print("  [4/4] Verifying Background AI Worker Daemon (agency-worker.service)... ")
	workerOut, workerErr := client.RunCommand("systemctl is-active agency-worker")
	if workerErr == nil && strings.Contains(workerOut, "active") {
		fmt.Println(Green + Bold + "✔ PASS (Daemon Active & Polling Queue)" + Reset)
	} else {
		fmt.Printf("%s⚠ WARNING: %s%s\n", Yellow, strings.TrimSpace(workerOut), Reset)
	}

	fmt.Println()
	fmt.Println(Green + Bold + "🎉 All core diagnostic checks completed!" + Reset)
	fmt.Println()
}

// showDnsGuide displays copy-paste ready DNS records
func showDnsGuide(cfg *config.Config) {
	fmt.Println()
	fmt.Println(Cyan + Bold + "╔═════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   📋 DNS CONFIGURATION & SSL GUIDE                          ║")
	fmt.Println("╚═════════════════════════════════════════════════════════════════════════════╝" + Reset)
	fmt.Println()
	fmt.Println("Add the following A-Records in your Domain Registrar (GoDaddy, Namecheap, Cloudflare, etc.):")
	fmt.Println()
	fmt.Printf("  ┌──────────────────────────────┬────────┬─────────────────────────────┐\n")
	fmt.Printf("  │ %-28s │ %-6s │ %-27s │\n", "HOSTNAME / SUBDOMAIN", "TYPE", "TARGET IP (VPS)")
	fmt.Printf("  ├──────────────────────────────┼────────┼─────────────────────────────┤\n")
	fmt.Printf("  │ %-28s │ A      │ %-27s │\n", cfg.PortalDomain, cfg.VPSHost)
	fmt.Printf("  │ %-28s │ A      │ %-27s │\n", cfg.AssetDomain, cfg.VPSHost)
	if cfg.SeedDemoTenant {
		fmt.Printf("  │ %-28s │ A      │ %-27s │\n", cfg.DemoDomain, cfg.VPSHost)
	}
	fmt.Printf("  └──────────────────────────────┴────────┴─────────────────────────────┘\n")
	fmt.Println()
	fmt.Println(White + Bold + "SSL Certificate Generation:" + Reset)
	fmt.Printf("  1. Log into CloudPanel: %shttps://%s:8443%s\n", Cyan, cfg.VPSHost, Reset)
	fmt.Println("  2. Navigate to your Site → Security Tab.")
	fmt.Println("  3. Click 'New Let's Encrypt Certificate' and click Create.")
	fmt.Println()
}

// exportSetupReport creates a persistent summary file
func exportSetupReport(cfg *config.Config) {
	filename := "setup-summary.txt"
	content := fmt.Sprintf(`================================================================================
VEFA Agency Multi-Tenant AI Platform - Setup Summary Report
Build: %s
Generated: %s
Official Community: https://www.VEFA.club (VEFA Trademarked 2026)
================================================================================

[AGENCY & ACCESS]
Agency Name:            %s
VPS Target IP:          %s (Port %d)
SSH Administrative User:%s
Portal Domain:          https://%s
Asset Domain:           https://%s
Demo Tenant Domain:     https://%s

[CONTROL PANEL]
CloudPanel URL:         https://%s:8443
Installation User:      %s

[DATABASE INFRASTRUCTURE]
Core Database Name:     %s
Core Database User:     %s
Demo Tenant DB Name:    %s
MariaDB Engine:         MariaDB 10.11

[AI & MCP ORCHESTRATION]
External MCP Endpoint:  %s
Webhook Relay Receiver: https://%s/api/webhooks/mcp-relay
Background Daemon:      agency-worker.service (/var/www/master-app/worker.php)

[SERVER PATHS]
Master Codebase:        /var/www/master-app
Public Root:            /var/www/master-app/public
Trade Modules:          /var/www/master-app/app/Trades
Core System Routes:     /var/www/master-app/app/Core
================================================================================
`, BuildVersion, time.Now().Format(time.RFC1123), cfg.AgencyName, cfg.VPSHost, cfg.SSHPort, cfg.SSHUser,
		cfg.PortalDomain, cfg.AssetDomain, cfg.DemoDomain, cfg.VPSHost, cfg.SystemUser,
		cfg.CoreDBName, cfg.DBUser, cfg.DemoDBName, cfg.MCPEndpoint, cfg.PortalDomain)

	if err := os.WriteFile(filename, []byte(content), 0644); err == nil {
		fmt.Printf("%s✔ Setup report successfully exported to '%s'%s\n\n", Green+Bold, filename, Reset)
	} else {
		fmt.Printf("%sCould not write report file: %v%s\n\n", Red, err, Reset)
	}
}

// cleanUpServer removes remote temporary install files
func cleanUpServer(client *sshclient.SSHClient) {
	if client == nil {
		return
	}
	fmt.Print(Cyan + "Cleaning up temporary installer scripts on VPS... " + Reset)
	client.RunCommand("rm -f /tmp/install.sh /tmp/agency_schema.sql /tmp/agency_seed.sql")
	fmt.Println(Green + "✔ Cleaned!" + Reset)
}

func main() {
	printBanner()

	configFile := flag.String("config", "agency-config.json", "Path to load/save configuration JSON")
	dryRun := flag.Bool("dry-run", false, "Simulate setup without modifying remote server")
	autoYes := flag.Bool("auto", false, "Non-interactive mode using existing config file")
	flag.Parse()

	cfg := config.NewDefaultConfig()
	reader := bufio.NewReader(os.Stdin)

	// Attempt to load existing config
	hasExistingConfig := false
	if _, err := os.Stat(*configFile); err == nil {
		loadedCfg, err := config.LoadFromFile(*configFile)
		if err == nil && loadedCfg.VPSHost != "" {
			cfg = loadedCfg
			hasExistingConfig = true
			fmt.Printf("%s[i] Loaded existing configuration from '%s'%s\n", Cyan, *configFile, Reset)
			fmt.Printf("    Agency: %s%s%s | VPS: %s%s%s | Portal: %s%s%s\n\n",
				White+Bold, cfg.AgencyName, Reset,
				White+Bold, cfg.VPSHost, Reset,
				White+Bold, cfg.PortalDomain, Reset)
		}
	}

	if !*autoYes {
		if hasExistingConfig {
			fmt.Printf("%sUse existing saved configuration for %s%s%s (%s)? [Y/n/e (e=edit)]: %s",
				Bold, Cyan, cfg.AgencyName, Reset+Bold, cfg.VPSHost, Reset)
			action, _ := reader.ReadString('\n')
			action = strings.TrimSpace(strings.ToLower(action))

			if action == "n" || action == "e" || action == "edit" {
				runInteractiveQuestionnaire(cfg, *configFile)
			}
		} else {
			runInteractiveQuestionnaire(cfg, *configFile)
		}
	}

	cfg.SaveToFile(*configFile)
	fmt.Printf("%s✔ Configuration saved locally to '%s'%s\n\n", Green, *configFile, Reset)

	// Review Summary
	fmt.Println(Yellow + Bold + "CONFIGURATION REVIEW:" + Reset)
	fmt.Printf("  • Agency:         %s%s%s\n", White+Bold, cfg.AgencyName, Reset)
	fmt.Printf("  • VPS Target:     %s%s:%d%s (%s)\n", White+Bold, cfg.VPSHost, cfg.SSHPort, Reset, cfg.SSHUser)
	fmt.Printf("  • Portal Domain:  %s%s%s\n", White+Bold, cfg.PortalDomain, Reset)
	fmt.Printf("  • Asset Domain:   %s%s%s\n", White+Bold, cfg.AssetDomain, Reset)
	fmt.Printf("  • MCP Endpoint:   %s%s%s\n", White+Bold, cfg.MCPEndpoint, Reset)
	if *dryRun {
		fmt.Println(Magenta + Bold + "\n[!] RUNNING IN DRY-RUN MODE: No remote commands will alter the VPS." + Reset)
	}
	fmt.Println()

	if !*autoYes {
		if !promptBool(reader, "Connect to VPS and begin automated setup?", true) {
			fmt.Println(Yellow + "Setup aborted by user." + Reset)
			pauseExit(0)
		}
	}

	var client *sshclient.SSHClient
	var err error

	if !*dryRun {
		client, err = connectWithInteractiveRecovery(cfg, *configFile)
		if err != nil {
			pauseExit(0)
		}
		defer client.Close()
	}

	prov := provisioner.NewProvisioner(client, cfg, templates.FS, *dryRun, func(level, msg string) {
		switch level {
		case "STEP":
			fmt.Printf("%s%s%s\n", Cyan+Bold, msg, Reset)
		case "SUCCESS":
			fmt.Printf("%s%s%s\n", Green+Bold, msg, Reset)
		case "WARN":
			fmt.Printf("%s%s%s\n", Yellow, msg, Reset)
		case "ERROR":
			fmt.Printf("%s%s%s\n", Red+Bold, msg, Reset)
		case "DRY_RUN":
			fmt.Printf("%s%s%s\n", Magenta, msg, Reset)
		default:
			fmt.Println(msg)
		}
	})

	for {
		if err := prov.RunFullSetup(); err != nil {
			fmt.Printf("\n%s❌ Provisioning encounter an issue: %v%s\n", Red+Bold, err, Reset)
			fmt.Println()
			fmt.Println(Yellow + Bold + "PROVISIONING RECOVERY:" + Reset)
			fmt.Println("  [1] 🔁 Retry provisioning")
			fmt.Println("  [2] 🔧 Edit configuration / credentials")
			fmt.Println("  [3] 🚪 Exit")
			fmt.Println()

			choice := promptString(reader, "Select option (1-3)", "1")
			if choice == "1" {
				continue
			} else if choice == "2" {
				runInteractiveQuestionnaire(cfg, *configFile)
				continue
			} else {
				pauseExit(1)
			}
		}
		break
	}

	// Celebratory Banner & VEFA Announcement
	printCelebrationBanner(cfg)

	// Summary Details
	fmt.Println(White + Bold + "📋 SYSTEM SUMMARY:" + Reset)
	fmt.Println(Dim + "─────────────────────────────────────────────────────────────────────────────" + Reset)
	fmt.Printf("  • %sCloudPanel Portal:%s      https://%s:8443\n", Bold, Reset, cfg.VPSHost)
	fmt.Printf("  • %sMaster App Location:%s    /var/www/master-app\n", Bold, Reset)
	fmt.Printf("  • %sCore Database:%s          %s (User: %s)\n", Bold, Reset, cfg.CoreDBName, cfg.DBUser)
	fmt.Printf("  • %sAI Worker Daemon:%s       agency-worker.service (Systemd: Active)\n", Bold, Reset)
	fmt.Printf("  • %sOfficial VEFA Portal:%s   https://www.VEFA.club\n", Bold, Reset)
	fmt.Println(Dim + "─────────────────────────────────────────────────────────────────────────────" + Reset)
	fmt.Println()

	// Interactive Post-Setup Loop
	for {
		fmt.Println(Yellow + Bold + "WHAT WOULD YOU LIKE TO DO NEXT?" + Reset)
		fmt.Printf("  [1] 🌐 Open CloudPanel in your Browser (https://%s:8443)\n", cfg.VPSHost)
		fmt.Println("  [2] 🧪 Run End-to-End System Tests (Live Verification)")
		fmt.Println("  [3] 📋 View DNS Configuration & SSL Guide")
		fmt.Println("  [4] 📄 Export Setup Report (setup-summary.txt)")
		fmt.Println("  [5] 🌐 Visit www.VEFA.club in Browser")
		fmt.Println("  [6] 🧹 Clean up temporary installer files and exit")
		fmt.Println("  [7] 🚪 Exit VEFA Agency Builder")
		fmt.Println()

		choice := promptString(reader, "Select an option (1-7)", "1")

		switch strings.TrimSpace(choice) {
		case "1":
			openBrowser(fmt.Sprintf("https://%s:8443", cfg.VPSHost))
			fmt.Println()
		case "2":
			runLiveDiagnostics(client, cfg)
		case "3":
			showDnsGuide(cfg)
		case "4":
			exportSetupReport(cfg)
		case "5":
			openBrowser("https://www.VEFA.club")
			fmt.Println()
		case "6":
			cleanUpServer(client)
			fmt.Println(Green + "Thank you for using VEFA Agency Builder! Visit https://www.VEFA.club for updates." + Reset)
			pauseExit(0)
		case "7":
			fmt.Println(Green + "Thank you for using VEFA Agency Builder! Visit https://www.VEFA.club for updates." + Reset)
			pauseExit(0)
		default:
			fmt.Println(Yellow + "Invalid choice. Please choose 1-7." + Reset)
		}
	}
}
