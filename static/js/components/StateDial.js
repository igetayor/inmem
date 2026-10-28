/**
 * Component: StateDial renders the prominent circular state indicator and cool-down timer.
 */
export class StateDial {
  constructor(containerId) {
    this.container = document.getElementById(containerId);
  }

  update(telemetry) {
    if (!this.container || !telemetry) return;

    const state = telemetry.state;
    let icon = "🟢";
    let subtext = "All downstream RPC calls flowing normally.";

    if (state === "OPEN") {
      icon = "🔴";
      subtext = "Failing Fast: All traffic rejected to prevent cascading collapse.";
    } else if (state === "HALF-OPEN") {
      icon = "🟡";
      subtext = "Cooldown elapsed. Probing downstream health with test requests.";
    }

    const cooldownHtml = (state === "OPEN" && telemetry.remaining_cooldown_ms > 0)
      ? `<div class="state-cooldown-timer">⏳ Cooldown: ${(telemetry.remaining_cooldown_ms / 1000).toFixed(1)}s</div>`
      : '';

    this.container.innerHTML = `
      <div class="state-badge-container">
        <div class="state-ring-indicator ${state}">
          ${icon}
        </div>
        <div class="state-text-info">
          <div style="font-size: 0.8rem; text-transform: uppercase; color: var(--text-dim); font-weight: 700; letter-spacing: 0.05em;">Current Circuit State</div>
          <h2 style="color: ${state === 'CLOSED' ? '#10b981' : state === 'OPEN' ? '#ef4444' : '#f59e0b'};">
            ${state}
          </h2>
          <p>${subtext}</p>
          ${cooldownHtml}
        </div>
      </div>
      <div>
        <button id="btnResetCircuit" class="btn btn-reset">
          ↺ Force Reset Circuit
        </button>
      </div>
    `;
  }
}
