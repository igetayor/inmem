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
      this.container.innerHTML = `
        <div class="empty-stream">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
          </svg>
          <span>No invocation events recorded yet. Trigger downstream traffic below.</span>
        </div>
      `;
      return;
    }

    this.container.innerHTML = history.slice().reverse().map(item => {
      let badgeClass = "badge-success";
      let statusIcon = "●";

      if (item.status === "SHORT_CIRCUITED") {
        badgeClass = "badge-danger";
        statusIcon = "✕";
      } else if (item.status === "FALLBACK") {
        badgeClass = "badge-warning";
        statusIcon = "⚠";
      }

      return `
        <div class="log-entry ${item.status.toLowerCase()}">
          <span class="log-time">${item.timestamp}</span>
          <span class="log-pill ${badgeClass}">
            <span class="pill-dot">${statusIcon}</span>
            ${item.status}
          </span>
          <span class="log-msg">${item.message}</span>
          <span class="log-meta">${item.state}</span>
        </div>
      `;
    }).join('');
  }
}
