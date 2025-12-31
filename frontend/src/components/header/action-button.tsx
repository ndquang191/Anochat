"use client";

import React from "react";
import { RotateCw, LogOut } from "lucide-react";
import { useQueue } from "@/hooks/use-queue";
import { useAuth } from "@/contexts/auth";
import { toast } from "sonner";
import { roomAPI } from "@/lib/api";

interface ButtonConfig {
	bgColor: string;
	icon: React.ReactNode;
	title: string;
	spinning: boolean;
}

export function ActionButton() {
	const { user, checkAuth } = useAuth();
	const { isInQueue, isLoading: isQueueLoading, joinQueue, leaveQueue } = useQueue();
	const [isRoomLoading, setIsRoomLoading] = React.useState(false);

	const inRoom = !!user?.room;
	const isLoading = isQueueLoading || isRoomLoading;

	const handleClick = async () => {
		if (isLoading) return;

		try {
			// Priority 1: Leave room if in room
			if (inRoom) {
				setIsRoomLoading(true);
				await roomAPI.leaveRoom();
				await checkAuth();
				toast.success("Đã rời phòng chat");
				setIsRoomLoading(false);
				return;
			}

			// Priority 2: Leave queue if in queue
			if (isInQueue) {
				await leaveQueue();
				toast.success("Đã rời khỏi hàng chờ");
			} else {
				// Priority 3: Join queue
				const result = await joinQueue("polite");
				if (result) {
					toast.success("Đã tham gia hàng chờ", {
						description: "Đang tìm kiếm đối tác chat...",
					});
				}
			}
		} catch (error) {
			console.error("Operation failed:", error);
			toast.error("Có lỗi xảy ra", {
				description: error instanceof Error ? error.message : "Vui lòng thử lại",
			});
			setIsRoomLoading(false);
		}
	};

	const getButtonConfig = (): ButtonConfig => {
		if (inRoom) {
			return {
				bgColor: "bg-red-500 hover:bg-red-600",
				icon: <LogOut size={18} />,
				title: "Rời phòng chat",
				spinning: false,
			};
		}
		if (isInQueue) {
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
