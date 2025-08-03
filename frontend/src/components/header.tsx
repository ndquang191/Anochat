"use client";
import React from "react";
import { Loader2, RotateCw, X } from "lucide-react";
import { useQueue } from "@/hooks/use-queue";

function ConnectDisconnectButton() {
	const { isInQueue, isLoading, joinQueue, leaveQueue, queueStatus } = useQueue();

	const handleClick = async () => {
		try {
			if (isInQueue) {
				await leaveQueue();
			} else {
				await joinQueue("polite"); // Default category
			}
		} catch (error) {
			console.error("Queue operation failed:", error);
		}
	};

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
			title={isInQueue ? `Đang trong queue (vị trí: ${queueStatus?.position || 0})` : "Tham gia queue"}
		>
			{/* Icon loading */}
			<div className={getIconClass(isLoading, true)}>
				<Loader2 size={18} />
			</div>

			{/* Icon đang trong queue */}
			<div className={getIconClass(isInQueue && !isLoading)}>
				<X size={18} />
			</div>

			{/* Icon ban đầu (chưa tham gia queue) */}
			<div className={getIconClass(!isInQueue && !isLoading)}>
				<RotateCw size={18} />
			</div>
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
				<div className="flex items-center">
					<ConnectDisconnectButton />
				</div>
			</header>
		</div>
	);
};

export default Header;
