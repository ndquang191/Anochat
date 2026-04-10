"use client";

import React, { useCallback, useEffect } from "react";
import { RotateCw, LogOut } from "lucide-react";
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
		client.send("leave_room", { room_id: room.id });
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
			toast.error(t("pleaseTryAgain"), {
				description: error instanceof Error ? error.message : t("pleaseTryAgain"),
			});
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
				icon: <LogOut size={18} />,
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
