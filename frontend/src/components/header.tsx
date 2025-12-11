"use client";
import React from "react";
import { RotateCw, X, LogOut } from "lucide-react";
import { useQueue } from "@/hooks/use-queue";
import { useAuth } from "@/contexts/auth";
import { toast } from "sonner";

function ConnectDisconnectButton() {
	const { user } = useAuth();
	const { isInQueue, isLoading, joinQueue, leaveQueue: leaveQueueFn, queueStatus } = useQueue();
	const [inRoom, setInRoom] = React.useState(false);

	// Check if user is in room via user state
	React.useEffect(() => {
		const checkRoomStatus = async () => {
			if (!user) return;
			// User state from auth context should have room info
			setInRoom(!!user.room);
		};
		checkRoomStatus();
	}, [user]);

	const handleClick = async () => {
		// Prevent multiple clicks
		if (isLoading) {
			return;
		}

		try {
			// Priority 1: If in room, show message (leave via chat interface)
			if (inRoom) {
				toast.info("Vui lòng sử dụng nút 'Rời phòng' trong chat");
				return;
			}

			// Priority 2: If in queue, leave queue
			if (isInQueue) {
				await leaveQueueFn();
				toast.success("Đã rời khỏi hàng chờ");
			} else {
				const result = await joinQueue("polite"); // Default category
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
		}
	};

	// Calculate time remaining until TTL expires
	const getTimeRemaining = () => {
		if (!queueStatus?.expires_at) return null;
		const now = new Date();
		const expiresAt = new Date(queueStatus.expires_at);
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
			className={`relative w-10 h-10 rounded-full transition-all duration-300 ease-in-out transform ${
				isLoading ? "cursor-not-allowed opacity-70" : "hover:scale-110"
			} ${inRoom ? "bg-red-500 hover:bg-red-600" : "bg-primary hover:bg-primary/90"}`}
			title={
				isLoading
					? "Đang xử lý..."
					: inRoom
					? "Rời phòng chat"
					: isInQueue
					? `Rời khỏi hàng chờ (vị trí: ${queueStatus?.position || 0})`
					: "Tham gia hàng chờ"
			}
		>
			{/* Loading state - always show spinning icon when loading */}
			{isLoading && (
				<div className={getIconClass(true, true)}>
					<RotateCw size={18} />
				</div>
			)}

			{/* Not loading - show appropriate icon based on state */}
			{!isLoading && (
				<>
					{/* Icon in room (disconnect) */}
					{inRoom && (
						<div className={getIconClass(true, false)}>
							<LogOut size={18} />
						</div>
					)}

					{/* Icon in queue (leave queue) */}
					{isInQueue && !inRoom && (
						<div className={getIconClass(true, false)}>
							<X size={18} />
						</div>
					)}

					{/* Icon idle (join queue) */}
					{!inRoom && !isInQueue && (
						<div className={getIconClass(true, false)}>
							<RotateCw size={18} />
						</div>
					)}
				</>
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
	const { user } = useAuth();

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
			{/* User info bar below header */}
			{user && user.profile && (
				<div className="absolute top-16 left-0 right-0 bg-muted/30 px-4 py-2 text-sm text-muted-foreground border-b">
					<div className="flex items-center gap-2">
						<span className="font-medium">{user.profile.name || "Người dùng"}</span>
						{user.profile.age && (
							<>
								<span>•</span>
								<span>{user.profile.age} tuổi</span>
							</>
						)}
						{user.profile.is_male !== null && (
							<>
								<span>•</span>
								<span>{user.profile.is_male ? "Nam" : "Nữ"}</span>
							</>
						)}
					</div>
				</div>
			)}
		</div>
	);
};

export default Header;
