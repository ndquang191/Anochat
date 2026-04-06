"use client";

import { useEffect } from "react";
import Chatbox from "@/components/chat-box";
import { AdminPanel } from "@/components/admin/admin-panel";
import { useAuth } from "@/contexts/auth";
import { useAdmin } from "@/contexts/admin";
import { Loader2, MessageCircle } from "lucide-react";
import { getWebSocketClient } from "@/lib/websocket";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { AdminUserID } from "@/types";
import { useQueue } from "@/hooks/use-queue";
import { playMatchSound } from "@/hooks/use-sound-notification";

function RippleEffect({ active, onClick }: { active: boolean; onClick: () => void }) {
	return (
		<div className="relative flex items-center justify-center w-72 h-72">
			{active ? (
				<>
					{/* Ping layer — fast fill pulse behind button */}
					<div className="absolute w-16 h-16 rounded-full bg-primary/20 animate-ping" />
					{/* Expanding border rings — start at button size */}
					{[0, 1, 2, 3].map((i) => (
						<div
							key={i}
							className="absolute w-16 h-16 rounded-full border border-primary/60 animate-ripple"
							style={{ animationDelay: `${i * 0.9}s` }}
						/>
					))}
				</>
			) : (
				/* Static radar rings */
				<>
					{[28, 44, 60].map((size) => (
						<div
							key={size}
							className="absolute rounded-full border border-primary/15"
							style={{ width: size * 2, height: size * 2 }}
						/>
					))}
				</>
			)}
			<button
				onClick={onClick}
				className="relative z-10 w-16 h-16 rounded-full bg-primary hover:bg-primary/85 active:scale-95 transition-all duration-150 cursor-pointer flex items-center justify-center shadow-lg"
				title={active ? "Huỷ tìm kiếm" : "Kết nối"}
			>
				<MessageCircle className="w-7 h-7 text-primary-foreground" />
			</button>
		</div>
	);
}

const Page = () => {
	const { user, room, inQueue, loading: authLoading } = useAuth();
	const { isAdminOpen } = useAdmin();
	const invalidateUserState = useInvalidateUserState();
	const isAdmin = user?.id === AdminUserID;
	const { joinQueue, leaveQueue, isLoading: isQueueLoading } = useQueue();

	// Ensure WebSocket is connected when in queue (e.g. after page refresh while inQueue=true)
	useEffect(() => {
		if (!user || !inQueue) return;
		const client = getWebSocketClient();
		if (!client.isConnected()) {
			client.connect().catch(console.error);
		}
	}, [user, inQueue]);

	// Handle match_found at page level so it works even when ChatBox is unmounted (inQueue=true)
	useEffect(() => {
		const client = getWebSocketClient();
		const handleMatchFound = () => {
			playMatchSound();
			invalidateUserState();
		};
		client.on("match_found", handleMatchFound);
		return () => client.off("match_found", handleMatchFound);
	}, [invalidateUserState]);

	if (isAdmin && isAdminOpen) {
		return (
			<div className="h-full w-full">
				<AdminPanel />
			</div>
		);
	}

	if (authLoading) {
		return (
			<div className="h-full w-full flex items-center justify-center">
				<div className="text-center space-y-4">
					<div className="flex justify-center">
						<Loader2 className="h-8 w-8 animate-spin text-primary" />
					</div>
					<h2 className="text-sm md:text-base font-semibold">Đang tải...</h2>
				</div>
			</div>
		);
	}

	if (room) {
		return (
			<div className="h-full w-full">
				<Chatbox />
			</div>
		);
	}

	const handleCTA = () => {
		if (isQueueLoading) return;
		inQueue ? leaveQueue() : joinQueue();
	};

	return (
		<div className="h-full w-full flex items-center justify-center">
			<div className="flex flex-col items-center gap-1">
				<RippleEffect active={inQueue} onClick={handleCTA} />
				<div className="text-center space-y-1">
					<h2 className="text-sm md:text-base font-semibold">
						{inQueue ? "App còn mới nên bạn chịu khó đợi chút nhé" : "Chưa có cuộc trò chuyện"}
					</h2>
				</div>
			</div>
		</div>
	);
};

export default Page;
