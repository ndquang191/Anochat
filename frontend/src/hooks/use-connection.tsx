"use client";

import { createContext, useContext, useState, useCallback } from "react";
import { useAlertDialogContext } from "@/components/alert-dialog-provider";

type ConnectionStatus = "disconnected" | "connecting" | "connected";

interface ConnectionContextType {
	isConnected: boolean;
	isConnecting: boolean;
	connect: () => Promise<void>;
	disconnect: () => Promise<void>;
}

const ConnectionContext = createContext<ConnectionContextType | null>(null);

export function ConnectionProvider({ children }: { children: React.ReactNode }) {
	const [status, setStatus] = useState<ConnectionStatus>("disconnected");
	const { open } = useAlertDialogContext();

	const isConnected = status === "connected";
	const isConnecting = status === "connecting";

	const connect = useCallback(async () => {
		if (status !== "disconnected") return;
		setStatus("connecting");
		// Simulate connection delay
		setStatus("connected");
	}, [status]);

	const disconnect = useCallback(async () => {
		if (status !== "connected") return;

		const confirmed = await open({
			title: "Ngừng cuộc trò chuyện?",
			description: "Bạn có chắc chắn muốn ngắt cuộc trò chuyện hiện tại?",
			confirmText: "Đồng ý, ngắt kết nối",
			cancelText: "Huỷ",
		});

		if (confirmed) {
			setStatus("disconnected");
		}
	}, [status, open]);

	return (
		<ConnectionContext.Provider
			value={{
				isConnected,
				isConnecting,
				connect,
				disconnect,
			}}
		>
			{children}
		</ConnectionContext.Provider>
	);
}

export function useConnection() {
	const context = useContext(ConnectionContext);
	if (!context) {
		throw new Error("useConnection must be used within ConnectionProvider");
	}
	return context;
}
