"use client";

import React from "react";
import { RotateCw, LogOut } from "lucide-react";
import { useQueue } from "@/hooks/use-queue";
import { useAuth } from "@/contexts/auth";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { toast } from "sonner";
import { getWebSocketClient } from "@/lib/websocket";

interface ButtonConfig {
	bgColor: string;
	icon: React.ReactNode;
	title: string;
	spinning: boolean;
}

export function ActionButton() {
	const { room, inQueue } = useAuth();
	const invalidateUserState = useInvalidateUserState();
	const { isLoading, joinQueue, leaveQueue } = useQueue();

	const inRoom = !!room;

	const handleClick = async () => {
		if (isLoading) return;

		try {
			if (inRoom) {
				const client = getWebSocketClient();
				client.send("leave_room", { room_id: room.id });
				invalidateUserState();
				toast.success("Đã rời phòng chat");
				return;
			}

			if (inQueue) {
				await leaveQueue();
			} else {
				await joinQueue();
			}
		} catch (error) {
			console.error("Operation failed:", error);
			toast.error("Có lỗi xảy ra", {
				description: error instanceof Error ? error.message : "Vui lòng thử lại",
			});
		}
	};

	const getButtonConfig = (): ButtonConfig => {
		if (inRoom) {
			return {
				bgColor: "bg-primary hover:bg-primary/90",
				icon: <LogOut size={18} />,
				title: "Rời phòng chat",
				spinning: false,
			};
		}
		if (inQueue) {
			return {
				bgColor: "bg-primary hover:bg-primary/90",
				icon: <RotateCw size={18} />,
				title: "Nhấn để rời hàng chờ",
				spinning: true,
			};
		}
		return {
			bgColor: "bg-primary hover:bg-primary/90",
			icon: <RotateCw size={18} />,
			title: "Tham gia hàng chờ",
			spinning: false,
		};
	};

	const config = getButtonConfig();

	return (
		<button
			onClick={handleClick}
			disabled={isLoading}
			className={`relative w-10 h-10 rounded-full transition-all duration-300 ease-in-out transform ${
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
