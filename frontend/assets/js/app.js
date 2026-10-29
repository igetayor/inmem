import { Api } from './api.js';
import { StateDial } from './components/StateDial.js';
import { MetricsCards } from './components/MetricsCards.js';
import { RequestStream } from './components/RequestStream.js';

class BreakerApp {
  constructor() {
    this.stateDial = new StateDial('stateHeroContainer');
    this.metricsCards = new MetricsCards('metricsContainer');
    this.requestStream = new RequestStream('activityLogContainer');
    this.pollInterval = null;
  }

  async init() {
    this.bindEvents();
    await this.refresh();
    this.pollInterval = setInterval(() => this.refresh(), 500);
  }

  bindEvents() {
    const successBtn = document.getElementById('btnSendSuccess');
    if (successBtn) {
      successBtn.addEventListener('click', () => this.handleSend(false));
    }

    const failBtn = document.getElementById('btnSendFailure');
    if (failBtn) {
      failBtn.addEventListener('click', () => this.handleSend(true));
    }

    document.addEventListener('click', async (e) => {
      if (e.target && e.target.id === 'btnResetCircuit') {
        await Api.reset();
        await this.refresh();
      }
    });
  }

  async handleSend(simulateFailure) {
    const serviceName = document.getElementById('serviceName')?.value || "Payment Gateway";
    await Api.execute({ simulate_failure: simulateFailure, service_name: serviceName });
    await this.refresh();
  }

  async refresh() {
    const data = await Api.getStatus();
    if (!data) return;

    this.stateDial.update(data.telemetry);
    this.metricsCards.update(data.telemetry);
    if (data.history) {
      this.requestStream.update(data.history);
    }
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const app = new BreakerApp();
  app.init();
});
