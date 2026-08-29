<?php
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

/**
 * Agency Core Portal & Webhook Routes
 */

// Centralized Portal Home / API Status
$app->get('/', function (Request $request, Response $response) {
    $html = <<<HTML
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Agency Core Portal</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-900 text-white min-h-screen flex items-center justify-center p-6">
    <div class="max-w-md w-full bg-slate-800 p-8 rounded-xl border border-slate-700 shadow-2xl text-center">
        <div class="inline-flex p-3 bg-sky-500/10 text-sky-400 rounded-full mb-4">
            <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
        </div>
        <h1 class="text-2xl font-bold mb-2">Agency Multi-Tenant Portal</h1>
        <p class="text-slate-400 text-sm mb-6">Centralized Identity Provider & MCP Orchestration Hub</p>
        <div class="bg-slate-900/80 p-4 rounded-lg text-left text-xs font-mono text-slate-300 space-y-2 border border-slate-700/50">
            <div class="flex justify-between"><span class="text-slate-500">Core Status:</span><span class="text-emerald-400 font-semibold">ONLINE</span></div>
            <div class="flex justify-between"><span class="text-slate-500">FPM Pool:</span><span class="text-sky-400">Consolidated</span></div>
            <div class="flex justify-between"><span class="text-slate-500">MCP Relay:</span><span class="text-emerald-400">Active</span></div>
        </div>
    </div>
</body>
</html>
HTML;
    $response->getBody()->write($html);
    return $response->withHeader('Content-Type', 'text/html');
});

// MCP Webhook Relay Receiver (Catches AI results from MCP orchestrator)
$app->post('/api/webhooks/mcp-relay', function (Request $request, Response $response) {
    $expectedSecret = '{{.MCPSecretKey}}';
    $authHeader = $request->getHeaderLine('Authorization');
    
    if (!empty($expectedSecret) && $authHeader !== "Bearer {$expectedSecret}") {
        $response->getBody()->write(json_encode(['error' => 'Unauthorized']));
        return $response->withHeader('Content-Type', 'application/json')->withStatus(401);
    }

    $payload = $request->getParsedBody();
    if (!is_array($payload)) {
        $raw = (string)$request->getBody();
        $payload = json_decode($raw, true) ?? [];
    }

    $jobId = $payload['job_id'] ?? null;
    $status = $payload['status'] ?? 'completed';
    $result = $payload['result'] ?? null;

    if (!$jobId) {
        $response->getBody()->write(json_encode(['error' => 'Missing job_id']));
        return $response->withHeader('Content-Type', 'application/json')->withStatus(400);
    }

    $coreDb = $this->get('core_db');

    // 1. Mark AI Queue Job Status & Response
    $stmt = $coreDb->prepare("UPDATE ai_job_queue SET status = :status, response = :response WHERE id = :id");
    $stmt->execute([
        'status' => $status,
        'response' => json_encode($result),
        'id' => $jobId
    ]);

    // 2. Materialize Data into Isolated Tenant Database if applicable
    if ($status === 'completed' && isset($result['matches']) && is_array($result['matches'])) {
        $tenantStmt = $coreDb->prepare("
            SELECT t.db_name FROM ai_job_queue q
            JOIN tenants t ON q.tenant_id = t.id
            WHERE q.id = :job_id LIMIT 1
        ");
        $tenantStmt->execute(['job_id' => $jobId]);
        $dbName = $tenantStmt->fetchColumn();

        if ($dbName) {
            $dbHost = getenv('DB_HOST') ?: '127.0.0.1';
            $dbUser = getenv('DB_USER') ?: '{{.DBUser}}';
            $dbPass = getenv('DB_PASS') ?: '{{.DBPassword}}';
            
            try {
                $tenantDb = new PDO("mysql:host={$dbHost};dbname={$dbName};charset=utf8mb4", $dbUser, $dbPass);
                $insert = $tenantDb->prepare("
                    INSERT INTO tutor_sessions (id, student_name, student_id, volunteer_id, ai_prep_notes, scheduled_for)
                    VALUES (UUID(), :student_name, :student_id, :volunteer_id, :ai_prep_notes, :scheduled_for)
                ");

                foreach ($result['matches'] as $match) {
                    $insert->execute([
                        'student_name' => $match['student_name'] ?? 'Assigned Student',
                        'student_id' => $match['student_id'] ?? null,
                        'volunteer_id' => $match['volunteer_id'] ?? null,
                        'ai_prep_notes' => $match['ai_prep_notes'] ?? '',
                        'scheduled_for' => $match['scheduled_time'] ?? date('Y-m-d H:i:s', strtotime('+1 day'))
                    ]);
                }
            } catch (Exception $ex) {
                error_log("Failed to materialize data to tenant DB [{$dbName}]: " . $ex->getMessage());
            }
        }
    }

    $response->getBody()->write(json_encode(['status' => 'acknowledged', 'job_id' => $jobId]));
    return $response->withHeader('Content-Type', 'application/json')->withStatus(200);
});

// List all tenants (Protected Core Endpoint)
$app->get('/api/admin/tenants', function (Request $request, Response $response) {
    $coreDb = $this->get('core_db');
    $stmt = $coreDb->query("SELECT id, domain, trade_module, db_name, created_at FROM tenants ORDER BY created_at DESC");
    $tenants = $stmt->fetchAll();
    
    $response->getBody()->write(json_encode(['data' => $tenants]));
    return $response->withHeader('Content-Type', 'application/json');
});
