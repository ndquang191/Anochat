"use client";

import { useState, useEffect } from "react";
import { useWebSocketChat } from "@/hooks/use-websocket-chat";
import { type ChatMessage } from "@/lib/websocket";
import { useAuth } from "@/contexts/auth";
import { toast } from "sonner";
import { ChatLoadingState } from "./chat/chat-loading-state";
import { ChatMessages } from "./chat/chat-messages";
import { ChatInput } from "./chat/chat-input";
import { WifiOff } from "lucide-react";
import { useLanguage } from "@/contexts/theme";

const CONNECTION_TIMEOUT_MS = 10000;

export default function ChatBox() {
	const { user, messages: initialMessages } = useAuth();
	const [connectionTimedOut, setConnectionTimedOut] = useState(false);
	const { t } = useLanguage();

	const { messages, sendMessage, isConnected, roomId } =
		useWebSocketChat({
			userId: user?.id || "",
			initialMessages: initialMessages as ChatMessage[],
			onMatchFound: () => {
				toast.success(t("matchFound"), {
					description: t("matchFoundDescription"),
				});
			},
			onPartnerLeft: () => {
				toast.info(t("partnerLeft"), {
					description: t("partnerLeftDescription"),
				});
			},
		});

	useEffect(() => {
		if (isConnected) {
			setConnectionTimedOut(false);
			return;
		}
		const timer = setTimeout(() => setConnectionTimedOut(true), CONNECTION_TIMEOUT_MS);
		return () => clearTimeout(timer);
	}, [isConnected]);

	if (!user) {
		return <ChatLoadingState message="Đang tải thông tin người dùng..." />;
	}

	if (!isConnected) {
		if (connectionTimedOut) {
			return (
				<div className="h-full w-full flex items-center justify-center">
					<div className="text-center space-y-3">
						<WifiOff className="h-4 w-4 md:h-5 md:w-5 text-muted-foreground mx-auto" />
						<p className="text-sm md:text-base font-medium">Không thể kết nối</p>
						<p className="text-sm md:text-base text-muted-foreground">Vui lòng tải lại trang</p>
					</div>
				</div>
			);
		}
		return <ChatLoadingState message="Đang kết nối WebSocket..." />;
	}

	return (
		<div className="flex flex-col bg-card text-card-foreground shadow-sm h-full relative">
			<ChatMessages messages={messages} currentUserId={user.id} />
			<ChatInput
				onSendMessage={sendMessage}
				disabled={!roomId}
			/>
		</div>
	);
}
