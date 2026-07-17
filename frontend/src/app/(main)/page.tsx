"use client";

import { useEffect, useState } from "react";
import LocalizedChatBox from "@/components/localized-chat-box";
import { AdminPanel } from "@/components/admin/admin-panel";
import { useAuth } from "@/contexts/auth";
import { useAdmin } from "@/contexts/admin";
import { useLanguage } from "@/contexts/theme";
import { Loader2, MessageCircle } from "lucide-react";
import { getWebSocketClient } from "@/lib/websocket";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { useQueue } from "@/hooks/use-queue";
import { playMatchSound } from "@/hooks/use-sound-notification";

function RippleEffect({ active, onClick }: { active: boolean; onClick: () => void }) {
	return (
		<div className="relative flex h-72 w-72 items-center justify-center">
			{active ? (
				<>
					<div className="absolute h-16 w-16 rounded-full bg-primary/20 animate-ping" />
					{[0, 1, 2, 3].map((i) => (
						<div
							key={i}
							className="absolute h-16 w-16 rounded-full border border-primary/60 animate-ripple"
							style={{ animationDelay: `${i * 0.9}s` }}
						/>
					))}
				</>
			) : (
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
				className="relative z-10 flex h-16 w-16 cursor-pointer items-center justify-center rounded-full bg-primary shadow-lg transition-all duration-150 hover:bg-primary/85 active:scale-95"
			>
				<MessageCircle className="h-7 w-7 text-primary-foreground" />
			</button>
		</div>
	);
}

const Page = () => {
	const { user, room, inQueue, loading: authLoading } = useAuth();
	const { isAdminOpen } = useAdmin();
	const { t } = useLanguage();
	const invalidateUserState = useInvalidateUserState();
	const isAdmin = user?.is_admin === true;
	const { joinQueue, leaveQueue, isLoading: isQueueLoading } = useQueue();
	const [showEndedChat, setShowEndedChat] = useState(false);

	useEffect(() => {
		if (!user || !inQueue) return;
		const client = getWebSocketClient();
		if (!client.isConnected()) {
			client.connect().catch(console.error);
		}
	}, [user, inQueue]);

	useEffect(() => {
		const client = getWebSocketClient();
		const handleMatchFound = () => {
			playMatchSound();
			setShowEndedChat(false);
			invalidateUserState();
		};
		const handlePartnerLeft = () => {
			setShowEndedChat(true);
		};
		const handleRoomLeft = () => {
			setShowEndedChat(false);
		};
		client.on("match_found", handleMatchFound);
		client.on("partner_left", handlePartnerLeft);
		client.on("room_left", handleRoomLeft);
		return () => {
			client.off("match_found", handleMatchFound);
			client.off("partner_left", handlePartnerLeft);
			client.off("room_left", handleRoomLeft);
		};
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
			<div className="flex h-full w-full items-center justify-center">
				<div className="space-y-4 text-center">
					<div className="flex justify-center">
						<Loader2 className="h-8 w-8 animate-spin text-primary" />
					</div>
					<h2 className="text-sm font-semibold md:text-base">{t("loading")}</h2>
				</div>
			</div>
		);
	}

	if (room || showEndedChat) {
		return (
			<div className="h-full w-full">
				<LocalizedChatBox />
			</div>
		);
	}

	const handleCTA = () => {
		if (isQueueLoading) return;
		setShowEndedChat(false);
		if (inQueue) {
			leaveQueue();
			return;
		}
		joinQueue();
	};

	return (
		<div className="flex h-full w-full items-center justify-center">
			<div className="flex flex-col items-center gap-1">
				<RippleEffect active={inQueue} onClick={handleCTA} />
				<div className="space-y-1 text-center">
					<h2 className="text-sm font-semibold md:text-base">
						{inQueue ? t("queueing") : t("noChatRoom")}
					</h2>
					<p className="text-xs text-muted-foreground md:text-sm">
						{inQueue ? t("leaveQueue") : t("findPartnerDescription")}
					</p>
				</div>
			</div>
		</div>
	);
};

export default Page;
