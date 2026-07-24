"use client";

import { useEffect, useState } from "react";
import { useWebSocketChat } from "@/hooks/use-websocket-chat";
import { type ChatMessage } from "@/lib/websocket";
import { useAuth } from "@/contexts/auth";
import { useLanguage } from "@/contexts/theme";
import { toast } from "sonner";
import { ChatLoadingState } from "@/components/chat/chat-loading-state";
import { ChatEmptyState } from "@/components/chat/chat-empty-state";
import { ChatMessages } from "@/components/chat/chat-messages";
import { ChatInput } from "@/components/chat/chat-input";
import { GenericErrorState } from "@/components/generic-error-state";

const CONNECTION_TIMEOUT_MS = 10000;

export default function LocalizedChatBox() {
	const {
		user,
		messages: initialMessages,
		messagesNextCursor,
		messagesHasMore,
	} = useAuth();
	const { t } = useLanguage();
	const [connectionTimedOut, setConnectionTimedOut] = useState(false);

	const {
		messages,
		sendMessage,
		isConnected,
		roomId,
		partnerLeft,
		hasMoreMessages,
		isLoadingOlder,
		loadOlderError,
		loadOlderMessages,
	} =
		useWebSocketChat({
			userId: user?.id || "",
			initialMessages: initialMessages as ChatMessage[],
			initialNextCursor: messagesNextCursor,
			initialHasMore: messagesHasMore,
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

		const timer = window.setTimeout(
			() => setConnectionTimedOut(true),
			CONNECTION_TIMEOUT_MS
		);
		return () => window.clearTimeout(timer);
	}, [isConnected]);

	if (!user) {
		return <ChatLoadingState message={t("loading")} />;
	}

	if (!isConnected) {
		return connectionTimedOut ? (
			<GenericErrorState />
		) : (
			<ChatLoadingState message={t("loading")} />
		);
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
			<ChatMessages
				messages={messages}
				currentUserId={user.id}
				hasMore={hasMoreMessages}
				isLoadingOlder={isLoadingOlder}
				loadOlderError={loadOlderError}
				onLoadOlder={loadOlderMessages}
			/>
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
