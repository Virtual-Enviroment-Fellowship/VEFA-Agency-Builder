/**
 * Resilient Frontend Polling Utility with Exponential Backoff and Jitter
 */

const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

async function pollJobStatus(pollUrl, maxAttempts = 15) {
    let attempt = 0;
    let baseDelay = 1500; // start at 1.5 seconds

    while (attempt < maxAttempts) {
        try {
            const response = await fetch(pollUrl);
            if (!response.ok) {
                console.warn(`Polling HTTP status: ${response.status}`);
            } else {
                const data = await response.json();

                // Terminal states
                if (data.status === 'completed') {
                    return data.response;
                }
                if (data.status === 'failed') {
                    throw new Error(data.response?.error || 'AI Job reported failure.');
                }

                console.log(`[Job Status] ${data.status} (Attempt ${attempt + 1}/${maxAttempts})`);
            }
        } catch (error) {
            console.warn("Polling network hiccup:", error.message);
        }

        attempt++;

        // Exponential backoff (max 8s) + random jitter (0-400ms) to avoid thundering herds
        const backoff = Math.min(baseDelay * Math.pow(1.4, attempt), 8000);
        const jitter = Math.random() * 400;
        await sleep(backoff + jitter);
    }

    throw new Error("Job polling timed out after " + maxAttempts + " attempts.");
}
