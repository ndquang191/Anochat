"use client";

import { useCallback } from "react";
import { useWebSocketChat } from "@/hooks/use-websocket-chat";
import { useAuth } from "@/contexts/auth";
import { toast } from "sonner";
import { ChatLoadingState } from "./chat/chat-loading-state";
import { ChatEmptyState } from "./chat/chat-empty-state";
import { ChatHeader } from "./chat/chat-header";
import { ChatMessages } from "./chat/chat-messages";
import { ChatInput } from "./chat/chat-input";

export default function ChatBox() {
	const { user } = useAuth();

	const { messages, sendMessage, isConnected, roomId, leaveRoom, isPartnerTyping, sendTypingIndicator } =
		useWebSocketChat({
			userId: user?.id || "",
			onMatchFound: () => {
				toast.success("Đã tìm thấy đối tác chat!", {
					description: `Bạn đã được kết nối với người dùng khác`,
				});
			},
			onPartnerLeft: () => {
				toast.info("Đối tác đã rời phòng", {
					description: "Bạn có thể tìm kiếm đối tác mới",
				});
			},
		});

	const handleLeaveRoom = useCallback(() => {
		if (roomId) {
			leaveRoom();
			toast.info("Bạn đã rời khỏi phòng chat");
		}
	}, [roomId, leaveRoom]);

	if (!user) {
		return <ChatLoadingState message="Đang tải thông tin người dùng..." />;
	}

	if (!isConnected) {
		return <ChatLoadingState message="Đang kết nối WebSocket..." />;
	}

	if (!roomId) {
		return (
			<ChatEmptyState
				title="Chưa có phòng chat nào"
				description="Vui lòng tham gia hàng chờ để tìm đối tác chat"
			/>
		);
	}

	return (
		<div className="flex flex-col bg-card text-card-foreground shadow-sm h-full relative">
			<ChatHeader isPartnerTyping={isPartnerTyping} onLeave={handleLeaveRoom} />
			<ChatMessages messages={messages} currentUserId={user.id} />
			<ChatInput
				onSendMessage={sendMessage}
				onTypingChange={sendTypingIndicator}
				disabled={!roomId}
			/>
		</div>
	);
}
