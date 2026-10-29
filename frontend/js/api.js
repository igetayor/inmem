/**
 * API Client for interacting with the Go Circuit Breaker server.
 */
export const Api = {
  async getStatus() {
    try {
      const res = await fetch('/api/status');
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return await res.json();
    } catch (err) {
      console.warn("Failed to fetch circuit breaker status:", err);
      return null;
    }
  },

  async execute({ simulate_failure = false, service_name = "Core Payment Pod" }) {
    const res = await fetch('/api/execute', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ simulate_failure, service_name })
    });
    return await res.json();
  },

  async reset() {
    const res = await fetch('/api/reset', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    });
    return await res.json();
  }
};
