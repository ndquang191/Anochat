"use client";

import React from "react";
import { RotateCw, LogOut } from "lucide-react";
import { useQueue } from "@/hooks/use-queue";
import { useAuth } from "@/contexts/auth";
import { toast } from "sonner";
import { getWebSocketClient } from "@/lib/websocket";
import { useLanguage } from "@/contexts/theme";

interface ButtonConfig {
	bgColor: string;
	icon: React.ReactNode;
	title: string;
	spinning: boolean;
}

export function ActionButton() {
	const { room, inQueue } = useAuth();
	const { isLoading, joinQueue, leaveQueue } = useQueue();
	const { t } = useLanguage();

	const inRoom = !!room;

	const handleClick = async () => {
		if (isLoading) return;

		try {
			if (inRoom) {
				const client = getWebSocketClient();
				client.send("leave_room", { room_id: room.id });
				toast.success(t("leaveChatRoomSuccess"));
				return;
			}

			if (inQueue) {
				await leaveQueue();
			} else {
				await joinQueue();
			}
		} catch (error) {
			console.error("Operation failed:", error);
			toast.error(t("somethingWentWrong"), {
				description: error instanceof Error ? error.message : t("pleaseTryAgain"),
			});
		}
	};

	const getButtonConfig = (): ButtonConfig => {
		if (inRoom) {
			return {
				bgColor: "bg-primary hover:bg-primary/90",
				icon: <LogOut size={18} />,
				title: t("leaveChatRoom"),
				spinning: false,
			};
		}
		if (inQueue) {
			return {
				bgColor: "bg-primary hover:bg-primary/90 opacity-70",
				icon: <RotateCw size={18} />,
				title: t("leaveQueue"),
				spinning: true,
			};
		}
		return {
			bgColor: "bg-primary hover:bg-primary/90",
			icon: <RotateCw size={18} />,
			title: t("joinQueue"),
			spinning: false,
		};
	};

	const config = getButtonConfig();

	return (
		<button
			onClick={handleClick}
			disabled={isLoading}
			className={`relative w-10 h-10 rounded-full transition-all duration-300 ease-in-out transform ${
				isLoading ? "cursor-not-allowed opacity-70" : "hover:scale-110 cursor-pointer"
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
