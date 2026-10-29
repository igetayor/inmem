/**
 * Component: StateDial renders the prominent circuit state indicator and cool-down timer.
 */
export class StateDial {
  constructor(containerId) {
    this.container = document.getElementById(containerId);
  }

  update(telemetry) {
    if (!this.container || !telemetry) return;

    const state = telemetry.state;
    let stateColor = "#10b981";
    let stateBorder = "rgba(16, 185, 129, 0.25)";
    let stateBg = "rgba(16, 185, 129, 0.08)";
    let subtext = "Healthy. Upstream RPC calls passing through directly without interception.";
    let iconSvg = `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M20 6L9 17l-5-5"/></svg>`;

    if (state === "OPEN") {
      stateColor = "#f43f5e";
      stateBorder = "rgba(244, 63, 94, 0.3)";
      stateBg = "rgba(244, 63, 94, 0.08)";
      subtext = "Short-circuiting active. Inbound requests immediately rejected to prevent cascade.";
      iconSvg = `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`;
    } else if (state === "HALF-OPEN") {
      stateColor = "#f59e0b";
      stateBorder = "rgba(245, 158, 11, 0.3)";
      stateBg = "rgba(245, 158, 11, 0.08)";
      subtext = "Cooldown elapsed. Canary trial requests are probing downstream availability.";
      iconSvg = `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>`;
    }

    const cooldownHtml = (state === "OPEN" && telemetry.remaining_cooldown_ms > 0)
      ? `<div class="cooldown-pill"><span class="cooldown-pulse"></span>Cooldown Remaining: ${(telemetry.remaining_cooldown_ms / 1000).toFixed(1)}s</div>`
      : '';

    this.container.innerHTML = `
      <div class="state-card-content">
        <div class="state-indicator" style="background: ${stateBg}; border-color: ${stateBorder}; color: ${stateColor};">
          ${iconSvg}
          <div class="state-glow" style="background: ${stateColor};"></div>
        </div>
        <div class="state-details">
          <div class="state-label">Circuit Status</div>
          <div class="state-title-row">
            <h2 class="state-name" style="color: ${stateColor};">${state}</h2>
            ${cooldownHtml}
          </div>
          <p class="state-desc">${subtext}</p>
        </div>
      </div>
      <div class="state-actions">
        <button id="btnResetCircuit" class="btn btn-secondary btn-sm">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 12a9 9 0 0 1 15-6.7L21 8"/><path d="M21 3v5h-5"/><path d="M21 12a9 9 0 0 1-15 6.7L3 16"/><path d="M3 21v-5h5"/></svg>
          Reset Circuit
        </button>
      </div>
    `;
  }
}
