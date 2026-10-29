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
    const isOverThreshold = telemetry.failure_count >= telemetry.max_failures;

    this.container.innerHTML = `
      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">Consecutive Failures</span>
          <span class="stat-indicator ${isOverThreshold ? 'danger' : 'neutral'}"></span>
        </div>
        <div class="stat-value ${isOverThreshold ? 'text-danger' : ''}">
          ${telemetry.failure_count}
          <span class="stat-sub-val">/ ${telemetry.max_failures} threshold</span>
        </div>
        <div class="stat-footer">Trips breaker when full</div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">Total Invocations</span>
          <span class="stat-indicator success"></span>
        </div>
        <div class="stat-value">${total.toLocaleString()}</div>
        <div class="stat-footer">${telemetry.successful_requests.toLocaleString()} successful calls</div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">Failure Rate</span>
          <span class="stat-indicator ${tripRate > 20 ? 'danger' : 'success'}"></span>
        </div>
        <div class="stat-value ${tripRate > 20 ? 'text-danger' : 'text-success'}">${tripRate}%</div>
        <div class="stat-footer">${failures.toLocaleString()} downstream errors</div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">Protective Trips</span>
          <span class="stat-indicator warning"></span>
        </div>
        <div class="stat-value text-warning">${telemetry.tripped_count}</div>
        <div class="stat-footer">Cascade prevention activations</div>
      </div>
    `;
  }
}
