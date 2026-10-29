/**
 * Component: MetricsCards renders aggregate request counters and trip thresholds.
 */
export class MetricsCards {
  constructor(containerId) {
    this.container = document.getElementById(containerId);
  }

  update(telemetry) {
    if (!this.container || !telemetry) return;

    const total = telemetry.total_requests || 0;
    const failures = telemetry.failed_requests || 0;
    const tripRate = total > 0 ? ((failures / total) * 100).toFixed(1) : 0;

    this.container.innerHTML = `
      <div class="metric-card">
        <div class="metric-label">Failure Threshold</div>
        <div class="metric-value" style="color: ${telemetry.failure_count >= telemetry.max_failures ? '#ef4444' : 'inherit'};">
          ${telemetry.failure_count} <span style="font-size: 1rem; color: var(--text-dim);">/ ${telemetry.max_failures}</span>
        </div>
        <div style="font-size: 0.8rem; color: var(--text-muted);">Consecutive failure trip bar</div>
      </div>

      <div class="metric-card">
        <div class="metric-label">Total Invocations</div>
        <div class="metric-value">${total}</div>
        <div style="font-size: 0.8rem; color: var(--text-muted);">${telemetry.successful_requests} successful calls</div>
      </div>

      <div class="metric-card">
        <div class="metric-label">Failure Rate</div>
        <div class="metric-value" style="color: ${tripRate > 20 ? '#ef4444' : '#10b981'};">${tripRate}%</div>
        <div style="font-size: 0.8rem; color: var(--text-muted);">${failures} downstream errors</div>
      </div>

      <div class="metric-card">
        <div class="metric-label">Circuits Tripped</div>
        <div class="metric-value" style="color: #f59e0b;">${telemetry.tripped_count}</div>
        <div style="font-size: 0.8rem; color: var(--text-muted);">Cascade protective actions</div>
      </div>
    `;
  }
}
