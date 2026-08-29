<?php
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

/**
 * Tutor Trade Module Routes
 */

// Home page for Tutor micro-site
$app->get('/', function (Request $request, Response $response) {
    $tenant = $this->get('tenant');
    $domain = htmlspecialchars($tenant['domain']);
    $html = <<<HTML
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tutor Portal - {$domain}</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50 text-gray-800">
    <div class="max-w-4xl mx-auto py-12 px-4">
        <div class="bg-white p-8 rounded-2xl shadow-sm border border-gray-200 mb-8">
            <div class="inline-flex items-center gap-2 bg-indigo-50 text-indigo-700 px-3 py-1 rounded-full text-xs font-semibold uppercase mb-4">
                Trade Module: Tutor
            </div>
            <h1 class="text-3xl font-bold text-gray-900 mb-2">Volunteer Tutor Coordination</h1>
            <p class="text-gray-600">Tenant Domain: <span class="font-mono text-indigo-600 font-semibold">{$domain}</span></p>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                <h2 class="text-xl font-bold mb-4 text-gray-900">Active Volunteers</h2>
                <div id="volunteers-list" class="text-sm text-gray-500">Loading volunteers...</div>
            </div>

            <div class="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                <h2 class="text-xl font-bold mb-4 text-gray-900">AI Match & Schedule</h2>
                <p class="text-sm text-gray-600 mb-4">Trigger asynchronous AI tutor matching without blocking web workers.</p>
                <button id="generate-btn" onclick="handleGenerateScheduleClick('demo-student-101')" class="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2.5 px-4 rounded-lg transition-colors flex items-center justify-center gap-2">
                    <span>Generate Match with AI</span>
                </button>
                <div id="ai-status" class="mt-4 text-sm font-mono hidden"></div>
            </div>
        </div>
    </div>

    <script src="/js/polling.js"></script>
    <script>
        // Load volunteers list on mount
        fetch('/api/volunteers')
            .then(res => res.json())
            .then(res => {
                const list = document.getElementById('volunteers-list');
                if (res.data && res.data.length > 0) {
                    list.innerHTML = res.data.map(v => '<div class="p-2 border-b border-gray-100 flex justify-between"><span>' + v.name + '</span><span class="text-xs bg-green-100 text-green-800 px-2 py-0.5 rounded font-semibold">' + v.status + '</span></div>').join('');
                } else {
                    list.innerHTML = '<div class="text-gray-400">No volunteers registered yet.</div>';
                }
            })
            .catch(() => {
                document.getElementById('volunteers-list').innerHTML = '<div class="text-red-500">Failed to load volunteers.</div>';
            });

        async function handleGenerateScheduleClick(studentId) {
            const btn = document.getElementById('generate-btn');
            const statusDiv = document.getElementById('ai-status');
            btn.disabled = true;
            btn.classList.add('opacity-50');
            statusDiv.classList.remove('hidden');
            statusDiv.innerHTML = '<div class="text-indigo-600 animate-pulse">⏳ Dispatching job to queue...</div>';

            try {
                const trigger = await fetch('/api/tutors/generate-schedule', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ student_id: studentId })
                });

                const initData = await trigger.json();
                statusDiv.innerHTML = '<div class="text-sky-600">⚡ Job accepted! Polling for AI result...</div>';

                const finalResult = await pollJobStatus(initData.poll_url);
                statusDiv.innerHTML = '<div class="text-green-600 font-bold">✅ AI Match Complete!</div><pre class="mt-2 bg-gray-900 text-gray-100 p-2 rounded text-xs overflow-x-auto">' + JSON.stringify(finalResult, null, 2) + '</pre>';
            } catch (err) {
                statusDiv.innerHTML = '<div class="text-red-500">❌ Error: ' + err.message + '</div>';
            } finally {
                btn.disabled = false;
                btn.classList.remove('opacity-50');
            }
        }
    </script>
</body>
</html>
HTML;
    $response->getBody()->write($html);
    return $response->withHeader('Content-Type', 'text/html');
});

// Read Isolated Tenant Data
$app->get('/api/volunteers', function (Request $request, Response $response) {
    $db = $this->get('db'); // Isolated tenant database connection
    $stmt = $db->query("SELECT id, name, email, status FROM volunteers WHERE status = 'active'");
    $volunteers = $stmt->fetchAll();
    
    $response->getBody()->write(json_encode(['data' => $volunteers]));
    return $response->withHeader('Content-Type', 'application/json');
});

// Dispatch Async AI Job (Non-Blocking)
$app->post('/api/tutors/generate-schedule', function (Request $request, Response $response) {
    $coreDb = $this->get('core_db');
    $tenant = $this->get('tenant');
    
    $parsedBody = $request->getParsedBody();
    if (!is_array($parsedBody)) {
        $raw = (string)$request->getBody();
        $parsedBody = json_decode($raw, true) ?? [];
    }

    $jobId = sprintf(
        '%04x%04x-%04x-%04x-%04x-%04x%04x%04x',
        mt_rand(0, 0xffff), mt_rand(0, 0xffff),
        mt_rand(0, 0xffff),
        mt_rand(0, 0x0fff) | 0x4000,
        mt_rand(0, 0x3fff) | 0x8000,
        mt_rand(0, 0xffff), mt_rand(0, 0xffff), mt_rand(0, 0xffff)
    );

    $payload = json_encode([
        'theme' => 'volunteer_coordination',
        'action' => 'generate_tutor_schedule',
        'target_student' => $parsedBody['student_id'] ?? 'student_generic',
        'tenant_domain' => $tenant['domain']
    ]);

    // Insert into shared Global AI Queue
    $stmt = $coreDb->prepare("
        INSERT INTO ai_job_queue (id, tenant_id, status, payload, created_at)
        VALUES (:id, :tenant_id, 'pending', :payload, NOW())
    ");
    $stmt->execute(['id' => $jobId, 'tenant_id' => $tenant['id'], 'payload' => $payload]);

    // Return 202 Accepted with polling URL
    $uiResponse = json_encode([
        'message' => 'Job queued successfully.',
        'job_id' => $jobId,
        'poll_url' => "/api/jobs/status/{$jobId}"
    ]);

    $response->getBody()->write($uiResponse);
    return $response->withHeader('Content-Type', 'application/json')->withStatus(202);
});

// The Status Polling Endpoint (Used by JS)
$app->get('/api/jobs/status/{id}', function (Request $request, Response $response, array $args) {
    $coreDb = $this->get('core_db');
    $stmt = $coreDb->prepare("SELECT status, response, created_at FROM ai_job_queue WHERE id = :id");
    $stmt->execute(['id' => $args['id']]);
    $job = $stmt->fetch();

    if (!$job) {
        $response->getBody()->write(json_encode(['error' => 'Job not found']));
        return $response->withHeader('Content-Type', 'application/json')->withStatus(404);
    }

    $data = [
        'status' => $job['status'],
        'response' => $job['response'] ? json_decode($job['response'], true) : null,
        'created_at' => $job['created_at']
    ];

    $response->getBody()->write(json_encode($data));
    return $response->withHeader('Content-Type', 'application/json');
});
