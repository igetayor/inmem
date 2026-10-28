/**
 * Component: RequestStream renders the real-time execution log.
 */
export class RequestStream {
  constructor(containerId) {
    this.container = document.getElementById(containerId);
  }

  update(history = []) {
    if (!this.container) return;

    if (history.length === 0) {
      this.container.innerHTML = '<div style="color: var(--text-dim); text-align: center; padding: 2rem;">No request activity logged yet. Dispatch a test request below.</div>';
      return;
    }

    this.container.innerHTML = history.slice().reverse().map(item => {
      let badgeStyle = "background: rgba(16,185,129,0.2); color: #34d399;";
      let label = item.status;

      if (item.status === "SHORT_CIRCUITED") {
        badgeStyle = "background: rgba(239,68,68,0.2); color: #f87171;";
      } else if (item.status === "FALLBACK") {
        badgeStyle = "background: rgba(245,158,11,0.2); color: #fbbf24;";
      }

      return `
        <div class="log-item ${item.status}">
          <span style="color: var(--text-dim); font-size: 0.75rem;">${item.timestamp}</span>
          <span style="padding: 0.15rem 0.4rem; border-radius: 4px; font-size: 0.7rem; font-weight: 700; ${badgeStyle}">
            ${label}
          </span>
          <span style="flex: 1; color: var(--text-main);">${item.message}</span>
          <span style="font-size: 0.75rem; color: var(--text-dim);">[State: ${item.state}]</span>
        </div>
      `;
    }).join('');
  }
}
