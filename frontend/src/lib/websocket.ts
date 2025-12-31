// lib/websocket.ts - Native WebSocket client for backend integration

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
	private maxReconnectAttempts = 5;
	private reconnectDelay = 1000;
	private isIntentionallyClosed = false;
	private isConnecting = false;

	constructor(url: string) {
		this.url = url;
	}

	connect(): Promise<void> {
		// Prevent multiple simultaneous connection attempts
		if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.CONNECTING)) {
			// Silently skip duplicate connection attempts
			return Promise.resolve();
		}

		// If already connected, resolve immediately
		if (this.ws && this.ws.readyState === WebSocket.OPEN) {
			// Already connected, return silently
			return Promise.resolve();
		}

		return new Promise((resolve, reject) => {
			try {
				this.isConnecting = true;

				// Create WebSocket connection
				// Note: HTTP-only cookies are automatically sent with the WebSocket upgrade request
				// No need to manually check token here, backend will validate
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

					// Attempt to reconnect if not intentionally closed
					if (!this.isIntentionallyClosed && this.reconnectAttempts < this.maxReconnectAttempts) {
						this.reconnectAttempts++;
						console.log(`Reconnecting... Attempt ${this.reconnectAttempts}`);
						setTimeout(() => {
							this.connect().catch(console.error);
						}, this.reconnectDelay * this.reconnectAttempts);
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

		// Also call wildcard handlers
		const wildcardHandlers = this.handlers.get("*");
		if (wildcardHandlers) {
			wildcardHandlers.forEach((handler) => handler(message));
		}
	}

	isConnected(): boolean {
		return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
	}
}

// Create singleton instance
let wsClient: WebSocketClient | null = null;

export function getWebSocketClient(): WebSocketClient {
	if (!wsClient) {
		const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
		const wsUrl = apiUrl.replace("http://", "ws://").replace("https://", "wss://") + "/ws";
		wsClient = new WebSocketClient(wsUrl);
	}
	return wsClient;
}

// Reset WebSocket singleton (call on logout)
export function resetWebSocketClient(): void {
	if (wsClient) {
		wsClient.disconnect();
		wsClient = null;
	}
}

export default getWebSocketClient;
