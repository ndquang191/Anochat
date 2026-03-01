export interface WebSocketMessage {
	type: string;
	payload: Record<string, unknown>;
}

export interface ChatMessage {
	id: string;
	room_id: string;
	sender_id: string;
	content: string;
	created_at: number;
}

type MessageHandler = (message: WebSocketMessage) => void;

export class WebSocketClient {
	private ws: WebSocket | null = null;
	private url: string;
	private handlers: Map<string, Set<MessageHandler>> = new Map();
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 20;
	private reconnectDelay = 1000;
	private maxReconnectDelay = 30000;
	private isIntentionallyClosed = false;
	private isConnecting = false;

	constructor(url: string) {
		this.url = url;
	}

	connect(): Promise<void> {
		if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.CONNECTING)) {
			return Promise.resolve();
		}

		if (this.ws && this.ws.readyState === WebSocket.OPEN) {
			return Promise.resolve();
		}

		return new Promise((resolve, reject) => {
			try {
				this.isConnecting = true;

				this.ws = new WebSocket(this.url);

				this.ws.onopen = () => {
					console.log("WebSocket connected");
					this.reconnectAttempts = 0;
					this.isIntentionallyClosed = false;
					this.isConnecting = false;
					resolve();
				};

				this.ws.onmessage = (event) => {
					try {
						const message: WebSocketMessage = JSON.parse(event.data);
						this.handleMessage(message);
					} catch (error) {
						console.error("Failed to parse WebSocket message:", error);
					}
				};

				this.ws.onerror = (error) => {
					console.error("WebSocket error:", error);
					this.isConnecting = false;
					reject(error);
				};

				this.ws.onclose = () => {
					console.log("WebSocket closed");
					this.ws = null;
					this.isConnecting = false;

					// Notify listeners that connection was lost
					this.handleMessage({ type: "disconnected", payload: {} });

					if (!this.isIntentionallyClosed && this.reconnectAttempts < this.maxReconnectAttempts) {
						this.reconnectAttempts++;
						const delay = Math.min(this.reconnectDelay * this.reconnectAttempts, this.maxReconnectDelay);
						console.log(`Reconnecting... Attempt ${this.reconnectAttempts} (delay: ${delay}ms)`);
						setTimeout(() => {
							this.connect().catch(console.error);
						}, delay);
					}
				};
			} catch (error) {
				reject(error);
			}
		});
	}

	disconnect() {
		this.isIntentionallyClosed = true;
		if (this.ws) {
			this.ws.close();
			this.ws = null;
		}
	}

	send(type: string, payload: Record<string, unknown>) {
		if (this.ws && this.ws.readyState === WebSocket.OPEN) {
			const message: WebSocketMessage = { type, payload };
			this.ws.send(JSON.stringify(message));
		} else {
			console.warn("WebSocket is not connected");
		}
	}

	on(type: string, handler: MessageHandler) {
		if (!this.handlers.has(type)) {
			this.handlers.set(type, new Set());
		}
		this.handlers.get(type)!.add(handler);
	}

	off(type: string, handler: MessageHandler) {
		const handlers = this.handlers.get(type);
		if (handlers) {
			handlers.delete(handler);
		}
	}

	private handleMessage(message: WebSocketMessage) {
		const handlers = this.handlers.get(message.type);
		if (handlers) {
			handlers.forEach((handler) => handler(message));
		}

		const wildcardHandlers = this.handlers.get("*");
		if (wildcardHandlers) {
			wildcardHandlers.forEach((handler) => handler(message));
		}
	}

	isConnected(): boolean {
		return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
	}
}

let wsClient: WebSocketClient | null = null;

export function getWebSocketClient(): WebSocketClient {
	if (!wsClient) {
		const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
		const wsUrl = apiUrl.replace("http://", "ws://").replace("https://", "wss://") + "/ws";
		wsClient = new WebSocketClient(wsUrl);
	}
	return wsClient;
}

export function resetWebSocketClient(): void {
	if (wsClient) {
		wsClient.disconnect();
		wsClient = null;
	}
}

export default getWebSocketClient;
