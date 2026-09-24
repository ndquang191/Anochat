"use client";

import { useEffect, useRef, useState } from "react";
import { CheckCircle2, Loader2, RefreshCw, WifiOff } from "lucide-react";
import { useWebSocketChat } from "@/hooks/use-websocket-chat";
import { type ChatMessage } from "@/lib/websocket";
import { useAuth } from "@/contexts/auth";
import { useLanguage } from "@/contexts/theme";
import { toast } from "sonner";
import { ChatLoadingState } from "@/components/chat/chat-loading-state";
import { ChatEmptyState } from "@/components/chat/chat-empty-state";
import { ChatMessages } from "@/components/chat/chat-messages";
import { ChatInput } from "@/components/chat/chat-input";
import { Button } from "@/components/ui/button";

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
	const [isRetrying, setIsRetrying] = useState(false);
	const [showReconnected, setShowReconnected] = useState(false);
	const wasDisconnectedRef = useRef(false);

	const {
		messages,
		sendMessage,
		isConnected,
		hasConnectedOnce,
		reconnect,
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
					descriptionClassName: "!text-foreground",
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

	useEffect(() => {
		if (hasConnectedOnce && !isConnected) {
			wasDisconnectedRef.current = true;
			setShowReconnected(false);
			return;
		}
		if (!isConnected || !wasDisconnectedRef.current) return;

		wasDisconnectedRef.current = false;
		setShowReconnected(true);
		const timer = window.setTimeout(() => setShowReconnected(false), 3000);
		return () => window.clearTimeout(timer);
	}, [hasConnectedOnce, isConnected]);

	const handleReconnect = async () => {
		if (isRetrying) return;
		setIsRetrying(true);
		setConnectionTimedOut(false);
		const connected = await reconnect();
		setIsRetrying(false);
		if (!connected) {
			setConnectionTimedOut(true);
		}
	};

	if (!user) {
		return <ChatLoadingState message={t("loading")} />;
	}

	if (!isConnected && !hasConnectedOnce) {
		if (!connectionTimedOut) {
			return <ChatLoadingState message={t("connectingWebSocket")} />;
		}

		return (
			<div className="flex h-full w-full items-center justify-center p-6">
				<div className="flex max-w-sm flex-col items-center gap-3 text-center">
					<WifiOff className="size-7 text-muted-foreground" aria-hidden="true" />
					<h2 className="text-base font-semibold">{t("connectionUnavailable")}</h2>
					<p className="text-sm text-muted-foreground">{t("connectionUnavailableDescription")}</p>
					<Button type="button" variant="outline" onClick={() => void handleReconnect()} disabled={isRetrying}>
						{isRetrying ? <Loader2 className="animate-spin" aria-hidden="true" /> : <RefreshCw aria-hidden="true" />}
						{t("retryConnection")}
					</Button>
				</div>
			</div>
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
			{!isConnected && (
				<div className="flex items-center justify-between gap-3 border-b border-amber-500/25 bg-amber-500/10 px-4 py-2.5 text-sm text-amber-900 dark:text-amber-200" role="status" aria-live="polite">
					<div className="flex min-w-0 items-center gap-2">
						<WifiOff className="size-4 shrink-0" aria-hidden="true" />
						<span className="truncate">{t("reconnecting")}</span>
					</div>
					<Button type="button" variant="outline" size="sm" className="h-7 shrink-0 bg-background/70" onClick={() => void handleReconnect()} disabled={isRetrying}>
						<RefreshCw className={isRetrying ? "animate-spin" : ""} aria-hidden="true" />
						{t("retryNow")}
					</Button>
				</div>
			)}
			{showReconnected && (
				<div className="flex items-center justify-center gap-2 border-b border-emerald-500/25 bg-emerald-500/10 px-4 py-2 text-sm text-emerald-800 dark:text-emerald-300" role="status" aria-live="polite">
					<CheckCircle2 className="size-4" aria-hidden="true" />
					{t("reconnected")}
				</div>
			)}
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
				disabled={!roomId || !isConnected}
			/>
		</div>
	);
}
