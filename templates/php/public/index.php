<?php
use DI\Container;
use Slim\Factory\AppFactory;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

require __DIR__ . '/../vendor/autoload.php';

// 1. Initialize Dependency Injection Container
$container = new Container();
AppFactory::setContainer($container);
$app = AppFactory::create();

// 2. Intercept Hostname from Nginx
$host = $_SERVER['HTTP_HOST'] ?? 'unknown';
$host = explode(':', $host)[0]; // Strip port if dev environment

// Load environment configuration
$dbHost = getenv('DB_HOST') ?: '127.0.0.1';
$dbUser = getenv('DB_USER') ?: '{{.DBUser}}';
$dbPass = getenv('DB_PASS') ?: '{{.DBPassword}}';
$coreDbName = getenv('DB_CORE_NAME') ?: '{{.CoreDBName}}';

// 3. Connect to Global Core Database
try {
    $coreDb = new PDO("mysql:host={$dbHost};dbname={$coreDbName};charset=utf8mb4", $dbUser, $dbPass);
    $coreDb->setAttribute(PDO::ATTR_DEFAULT_FETCH_MODE, PDO::FETCH_ASSOC);
    $coreDb->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
} catch (PDOException $e) {
    http_response_code(500);
    header('Content-Type: application/json');
    echo json_encode(['error' => 'Core Database Connection Failed', 'details' => $e->getMessage()]);
    exit;
}

// 4. Look up Tenant by Domain
$stmt = $coreDb->prepare("SELECT id, trade_module, db_name, api_keys FROM tenants WHERE domain = :domain LIMIT 1");
$stmt->execute(['domain' => $host]);
$tenant = $stmt->fetch();

// Check if this is the Master Portal Domain
$portalDomain = '{{.PortalDomain}}';
if (!$tenant && ($host === $portalDomain || $host === 'localhost' || $host === '127.0.0.1')) {
    // Master Agency Core System Routes
    $tenant = [
        'id' => '00000000-0000-0000-0000-000000000000',
        'trade_module' => 'Core',
        'db_name' => $coreDbName,
        'domain' => $host
    ];
}

if (!$tenant) {
    http_response_code(404);
    header('Content-Type: application/json');
    echo json_encode([
        'error' => 'Domain not configured',
        'requested_host' => $host,
        'message' => 'Please register this domain in the Agency Admin Portal.'
    ]);
    exit;
}

// 5. Inject Tenant Context & Core DB into Container
$container->set('tenant', function() use ($tenant) { return $tenant; });
$container->set('core_db', function() use ($coreDb) { return $coreDb; });

// 6. Inject ISOLATED Tenant Database dynamically
$container->set('db', function() use ($tenant, $dbHost, $dbUser, $dbPass) {
    $dsn = "mysql:host={$dbHost};dbname=" . $tenant['db_name'] . ";charset=utf8mb4";
    $tenantDb = new PDO($dsn, $dbUser, $dbPass);
    $tenantDb->setAttribute(PDO::ATTR_DEFAULT_FETCH_MODE, PDO::FETCH_ASSOC);
    $tenantDb->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    return $tenantDb;
});

// 7. Load Trade Module Routes
$moduleName = ucfirst($tenant['trade_module']);
$modulePath = __DIR__ . "/../app/Trades/{$moduleName}/routes.php";

if ($moduleName === 'Core') {
    $modulePath = __DIR__ . "/../app/Core/routes.php";
}

if (file_exists($modulePath)) {
    require $modulePath;
} else {
    http_response_code(500);
    header('Content-Type: application/json');
    echo json_encode(['error' => "Module logic missing for trade module: {$moduleName}"]);
    exit;
}

// Global Health Route
$app->get('/api/health', function (Request $request, Response $response) use ($tenant) {
    $data = [
        'status' => 'healthy',
        'tenant' => $tenant['domain'],
        'trade_module' => $tenant['trade_module'],
        'timestamp' => date('c')
    ];
    $response->getBody()->write(json_encode($data));
    return $response->withHeader('Content-Type', 'application/json');
});

$app->addRoutingMiddleware();
$app->addErrorMiddleware(true, true, true);
$app->run();
