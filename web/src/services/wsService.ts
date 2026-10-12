import { StreamMessage } from '../types/log';

export class ScanWebSocket {
  private ws: WebSocket | null = null;
  private scanId: string;
  private onMessage: (msg: StreamMessage) => void;
  private onClose?: () => void;

  constructor(scanId: string, onMessage: (msg: StreamMessage) => void, onClose?: () => void) {
    this.scanId = scanId;
    this.onMessage = onMessage;
    this.onClose = onClose;
  }

  public connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/ws/scans/${this.scanId}/logs`;

    this.ws = new WebSocket(wsUrl);

    this.ws.onmessage = (event) => {
      try {
        const msg: StreamMessage = JSON.parse(event.data);
        this.onMessage(msg);
      } catch (err) {
        console.error('Failed to parse WebSocket stream message', err);
      }
    };

    this.ws.onclose = () => {
      if (this.onClose) {
        this.onClose();
      }
    };

    this.ws.onerror = (err) => {
      console.error('WebSocket connection error:', err);
    };
  }

  public disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}
