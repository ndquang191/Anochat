"use client";

import { useEffect } from "react";
import Chatbox from "@/components/chat-box";
import { useAuth } from "@/contexts/auth";
import { Loader2 } from "lucide-react";
import { getWebSocketClient } from "@/lib/websocket";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";

const Page = () => {
	const { user, room, inQueue, loading: authLoading } = useAuth();
	const invalidateUserState = useInvalidateUserState();

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
			invalidateUserState();
		};
		client.on("match_found", handleMatchFound);
		return () => client.off("match_found", handleMatchFound);
	}, [invalidateUserState]);

	if (authLoading) {
		return (
			<div className="h-full w-full flex items-center justify-center">
				<div className="text-center space-y-4">
					<div className="flex justify-center">
						<Loader2 className="h-12 w-12 animate-spin text-primary" />
					</div>
					<div className="space-y-2">
						<h2 className="text-xl font-semibold">Đang tải...</h2>
						<p className="text-sm text-muted-foreground">
							Vui lòng chờ trong giây lát
						</p>
					</div>
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

	if (inQueue) {
		return (
			<div className="h-full w-full flex items-center justify-center">
				<div className="text-center space-y-4">
					<div className="flex justify-center">
						<Loader2 className="h-12 w-12 animate-spin text-primary" />
					</div>
					<div className="space-y-2">
						<h2 className="text-xl font-semibold">Đang trong hàng chờ...</h2>
						<p className="text-sm text-muted-foreground">
							Vui lòng chờ trong khi chúng tôi tìm kiếm người chat cho bạn
						</p>
					</div>
				</div>
			</div>
		);
	}

	return (
		<div className="h-full w-full">
			<Chatbox />
		</div>
	);
};

export default Page;
