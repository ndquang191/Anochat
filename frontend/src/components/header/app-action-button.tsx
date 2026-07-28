"use client";

import React, { useCallback, useEffect } from "react";
import { DoorOpen, RotateCw } from "lucide-react";
import { useQueue } from "@/hooks/use-queue";
import { useAuth } from "@/contexts/auth";
import { useLanguage } from "@/contexts/theme";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { toast } from "sonner";
import { getWebSocketClient } from "@/lib/websocket";
import { useAlertDialogContext } from "@/contexts/alert-dialog";

interface ButtonConfig {
	bgColor: string;
	icon: React.ReactNode;
	title: string;
	spinning: boolean;
}

const ROOM_LEAVE_ACK_TIMEOUT_MS = 5000;

export function AppActionButton() {
	const { room, inQueue } = useAuth();
	const { t } = useLanguage();
	const invalidateUserState = useInvalidateUserState();
	const { isLoading, joinQueue, leaveQueue } = useQueue();
	const alertDialog = useAlertDialogContext();

	const inRoom = !!room;

	const leaveRoom = useCallback(async () => {
		if (!room) return;

		const confirmed = await alertDialog.open({
			title: t("leaveChatRoom"),
			description: t("leaveChatRoomConfirmDescription"),
			confirmText: t("leaveChatRoom"),
			cancelText: t("cancel"),
		});

		if (!confirmed) return;

		const client = getWebSocketClient();
		await new Promise<void>((resolve, reject) => {
			const cleanup = () => {
				clearTimeout(timeoutId);
				client.off("room_left", handleRoomLeft);
				client.off("room_leave_failed", handleRoomLeaveFailed);
				client.off("disconnected", handleDisconnected);
			};
			const handleRoomLeft = (message: { payload: Record<string, unknown> }) => {
				if (message.payload.room_id !== room.id) return;
				cleanup();
				resolve();
			};
			const handleRoomLeaveFailed = (message: { payload: Record<string, unknown> }) => {
				if (message.payload.room_id && message.payload.room_id !== room.id) return;
				cleanup();
				reject(new Error(t("leaveChatRoomFailed")));
			};
			const handleDisconnected = () => {
				cleanup();
				reject(new Error(t("leaveChatRoomFailed")));
			};

			client.on("room_left", handleRoomLeft);
			client.on("room_leave_failed", handleRoomLeaveFailed);
			client.on("disconnected", handleDisconnected);
			const timeoutId = setTimeout(() => {
				cleanup();
				reject(new Error(t("leaveChatRoomFailed")));
			}, ROOM_LEAVE_ACK_TIMEOUT_MS);

			if (!client.send("leave_room", { room_id: room.id })) {
				cleanup();
				reject(new Error(t("leaveChatRoomFailed")));
			}
		});
		invalidateUserState();
		toast.success(t("leaveChatRoomSuccess"));
	}, [alertDialog, invalidateUserState, room, t]);

	const handleClick = useCallback(async () => {
		if (isLoading) return;

		try {
			if (inRoom) {
				await leaveRoom();
				return;
			}

			if (inQueue) {
				await leaveQueue();
			} else {
				await joinQueue();
			}
		} catch (error) {
			console.error("Operation failed:", error);
			toast.error(
				error instanceof Error ? error.message : t("somethingWentWrong")
			);
		}
	}, [inQueue, inRoom, isLoading, joinQueue, leaveQueue, leaveRoom, t]);

	useEffect(() => {
		const handleKeyDown = (event: KeyboardEvent) => {
			if (!(event.ctrlKey && event.key === "Enter")) return;
			if (event.repeat) return;

			event.preventDefault();
			void handleClick();
		};

		window.addEventListener("keydown", handleKeyDown);
		return () => window.removeEventListener("keydown", handleKeyDown);
	}, [handleClick]);

	const getButtonConfig = (): ButtonConfig => {
		if (inRoom) {
			return {
				bgColor: "bg-primary hover:bg-primary/90",
				icon: <DoorOpen size={18} />,
				title: t("leaveChatRoomShortcut"),
				spinning: false,
			};
		}
		if (inQueue) {
			return {
				bgColor: "bg-primary hover:bg-primary/90",
				icon: <RotateCw size={18} />,
				title: t("leaveQueueShortcut"),
				spinning: true,
			};
		}
		return {
			bgColor: "bg-primary hover:bg-primary/90",
			icon: <RotateCw size={18} />,
			title: t("joinQueueShortcut"),
			spinning: false,
		};
	};

	const config = getButtonConfig();

	return (
		<button
			onClick={handleClick}
			disabled={isLoading}
			className={`relative h-10 w-10 rounded-full transform transition-all duration-300 ease-in-out ${
				isLoading ? "cursor-not-allowed opacity-70" : "hover:scale-110"
			} ${config.bgColor}`}
			aria-label={config.title}
			title={config.title}
		>
			<div
				className={`absolute inset-0 flex items-center justify-center text-white ${
					config.spinning ? "animate-spin" : ""
				}`}
			>
				{config.icon}
			</div>
		</button>
	);
}
