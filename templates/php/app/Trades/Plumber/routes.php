<?php
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

/**
 * Plumber Trade Module Routes
 */

$app->get('/', function (Request $request, Response $response) {
    $tenant = $this->get('tenant');
    $domain = htmlspecialchars($tenant['domain']);
    $html = <<<HTML
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Plumbing Services - {$domain}</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 text-gray-800 p-8">
    <div class="max-w-2xl mx-auto bg-white p-8 rounded-xl shadow border border-gray-200">
        <h1 class="text-2xl font-bold text-blue-800 mb-2">Emergency Plumbing Dispatch</h1>
        <p class="text-gray-600 mb-4">Domain: {$domain}</p>
        <div class="bg-blue-50 p-4 rounded-lg text-blue-900 text-sm">
            Plumbing Module active. Database connection isolated for this tenant.
        </div>
    </div>
</body>
</html>
HTML;
    $response->getBody()->write($html);
    return $response->withHeader('Content-Type', 'text/html');
});
