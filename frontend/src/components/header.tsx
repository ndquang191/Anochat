"use client";
import React from "react";
import { RotateCw, X, LogOut } from "lucide-react";
import { useQueue } from "@/hooks/use-queue";
import { roomAPI } from "@/lib/api";
import { useAuth } from "@/contexts/auth";

function ConnectDisconnectButton() {
	const { isInQueue, isLoading, joinQueue, leaveQueue, queueStatus } = useQueue();

	const handleClick = async () => {
		try {
			if (isInQueue) {
				console.log("🔄 Leaving queue...");
				await leaveQueue();
			} else {
				console.log("🚀 Joining queue...");
				await joinQueue("polite"); // Default category
			}
		} catch (error) {
			console.error("❌ Queue operation failed:", error);
		}
	};

	// Calculate time remaining until TTL expires
	const getTimeRemaining = () => {
		if (!queueStatus?.expiresAt) return null;
		const now = new Date();
		const expiresAt = new Date(queueStatus.expiresAt);
		const remaining = expiresAt.getTime() - now.getTime();

		if (remaining <= 0) return "Hết hạn";
		const seconds = Math.ceil(remaining / 1000);
		return `${seconds}s`;
	};

	const timeRemaining = getTimeRemaining();
	const iconBaseClass = `absolute inset-0 flex items-center justify-center text-white transition-all duration-300`;
	const getIconClass = (visible: boolean, spinning = false) =>
		`${iconBaseClass} ${visible ? "opacity-100 scale-100 rotate-0" : "opacity-0 scale-0 rotate-90"} ${
			spinning ? "animate-spin" : ""
		}`;

	return (
		<button
			onClick={handleClick}
			disabled={isLoading}
			className={`relative w-10 h-10 rounded-full transition-all duration-300 ease-in-out transform hover:scale-110 bg-primary hover:bg-primary/90`}
			title={
				isInQueue
					? `Đang trong queue (vị trí: ${queueStatus?.position || 0}, còn lại: ${
							timeRemaining || "N/A"
					  })`
					: "Tham gia queue"
			}
		>
			{/* Icon đang trong queue */}
			<div className={getIconClass(isInQueue && !isLoading)}>
				<X size={18} />
			</div>

			{/* Icon connect (idle + loading states) */}
			<div className={getIconClass(!isInQueue || isLoading, isLoading)}>
				<RotateCw size={18} />
			</div>

			{/* TTL indicator - small dot showing time remaining */}
			{isInQueue && timeRemaining && (
				<div className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 rounded-full text-xs text-white flex items-center justify-center font-bold">
					{timeRemaining === "Hết hạn" ? "!" : timeRemaining}
				</div>
			)}
		</button>
	);
}

function LeaveRoomButton() {
	const { user } = useAuth();
	const [isLoading, setIsLoading] = React.useState(false);

	// Only show if user has an active room
	const hasActiveRoom = user?.room !== null;

	const handleLeaveRoom = async () => {
		if (!hasActiveRoom) return;

		setIsLoading(true);
		try {
			await roomAPI.leaveRoom();
			// The user state will be updated automatically via the auth context
		} catch (error) {
			console.error("❌ Leave room failed:", error);
		} finally {
			setIsLoading(false);
		}
	};

	if (!hasActiveRoom) return null;

	return (
		<button
			onClick={handleLeaveRoom}
			disabled={isLoading}
			className={`w-10 h-10 rounded-full transition-all duration-300 ease-in-out transform hover:scale-110 bg-red-500 hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center`}
			title="Rời phòng chat"
		>
			<LogOut size={18} className="text-white" />
		</button>
	);
}

const Header = ({ trigger }: { trigger: React.ReactNode }) => {
	return (
		<div>
			<header className="absolute top-0 left-0 right-0 flex h-16 shrink-0 items-center justify-between border-b px-4">
				<div className="flex items-center gap-2">
					{trigger}
					<h1 className="text-xl font-semibold">Chat ẩn danh</h1>
				</div>
				<div className="flex items-center gap-2">
					<LeaveRoomButton />
					<ConnectDisconnectButton />
				</div>
			</header>
		</div>
	);
};

export default Header;
