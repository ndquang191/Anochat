"use client";

import { useWebSocketChat } from "@/hooks/use-websocket-chat";
import { type ChatMessage } from "@/lib/websocket";
import { useAuth } from "@/contexts/auth";
import { useLanguage } from "@/contexts/theme";
import { toast } from "sonner";
import { ChatLoadingState } from "@/components/chat/chat-loading-state";
import { ChatEmptyState } from "@/components/chat/chat-empty-state";
import { ChatMessages } from "@/components/chat/chat-messages";
import { ChatInput } from "@/components/chat/chat-input";

export default function LocalizedChatBox() {
	const { user, messages: initialMessages } = useAuth();
	const { t } = useLanguage();

	const { messages, sendMessage, isConnected, roomId, partnerLeft } =
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

	if (!user) {
		return <ChatLoadingState message={t("loadingUser")} />;
	}

	if (!isConnected) {
		return <ChatLoadingState message={t("connectingWebSocket")} />;
	}

	if (!roomId && !partnerLeft) {
		return (
			<ChatEmptyState
				title={t("noChatRoom")}
				description={t("findPartnerDescription")}
			/>
		);
	}

	return (
		<div className="relative flex h-full flex-col bg-card text-card-foreground shadow-sm">
			<ChatMessages messages={messages} currentUserId={user.id} />
			{partnerLeft && (
				<div className="border-t bg-muted/35 px-4 py-3 text-center text-sm text-muted-foreground">
					{t("partnerLeftInlineNotice")}
				</div>
			)}
			<ChatInput
				onSendMessage={sendMessage}
				disabled={!roomId}
			/>
		</div>
	);
}
