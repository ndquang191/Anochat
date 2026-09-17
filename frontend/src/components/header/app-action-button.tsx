"use client";

import { useCallback, useEffect, useState } from "react";
import { DoorOpen, UserRoundSearch } from "lucide-react";
import { useAuth } from "@/contexts/auth";
import { useLanguage } from "@/contexts/theme";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { toast } from "sonner";
import { getWebSocketClient } from "@/lib/websocket";
import { useAlertDialogContext } from "@/contexts/alert-dialog";
import { Button } from "@/components/ui/button";
import { useQueue } from "@/hooks/use-queue";

const ROOM_LEAVE_ACK_TIMEOUT_MS = 5000;

export function AppActionButton() {
	const { room } = useAuth();
	const { t } = useLanguage();
	const invalidateUserState = useInvalidateUserState();
	const alertDialog = useAlertDialogContext();
	const { joinQueue } = useQueue();
	const [isLoading, setIsLoading] = useState(false);
	const [partnerLeft, setPartnerLeft] = useState(false);

	useEffect(() => {
		const client = getWebSocketClient();
		const handlePartnerLeft = () => setPartnerLeft(true);
		const handleRoomActive = () => setPartnerLeft(false);

		client.on("partner_left", handlePartnerLeft);
		client.on("match_found", handleRoomActive);
		client.on("room_rejoined", handleRoomActive);
		return () => {
			client.off("partner_left", handlePartnerLeft);
			client.off("match_found", handleRoomActive);
			client.off("room_rejoined", handleRoomActive);
		};
	}, []);

	const leaveRoom = useCallback(async () => {
		if (!room) return;

		const client = getWebSocketClient();
		await new Promise<void>((resolve, reject) => {
			const cleanup = () => {
				clearTimeout(timeoutId);
				client.off("room_left", handleRoomLeft);
				client.off("room_leave_failed", handleRoomLeaveFailed);
				client.off("disconnected", handleDisconnected);
			};
			const handleRoomLeft = (message: { payload: Record<string, unknown> }) => {
				if (message.payload.room_id !== room.id) return;
				cleanup();
				resolve();
			};
			const handleRoomLeaveFailed = (message: { payload: Record<string, unknown> }) => {
				if (message.payload.room_id && message.payload.room_id !== room.id) return;
				cleanup();
				reject(new Error(t("leaveChatRoomFailed")));
			};
			const handleDisconnected = () => {
				cleanup();
				reject(new Error(t("leaveChatRoomFailed")));
			};

			client.on("room_left", handleRoomLeft);
			client.on("room_leave_failed", handleRoomLeaveFailed);
			client.on("disconnected", handleDisconnected);
			const timeoutId = setTimeout(() => {
				cleanup();
				reject(new Error(t("leaveChatRoomFailed")));
			}, ROOM_LEAVE_ACK_TIMEOUT_MS);

			if (!client.send("leave_room", { room_id: room.id })) {
				cleanup();
				reject(new Error(t("leaveChatRoomFailed")));
			}
		});
		invalidateUserState();
	}, [invalidateUserState, room, t]);

	const handleClick = useCallback(async () => {
		if ((!room && !partnerLeft) || isLoading) return;
		if (!partnerLeft) {
			const confirmed = await alertDialog.open({
				title: t("leaveChatRoom"),
				description: t("leaveChatRoomConfirmDescription"),
				confirmText: t("leaveChatRoom"),
				cancelText: t("cancel"),
			});
			if (!confirmed) return;
		}

		setIsLoading(true);
		try {
			if (partnerLeft) {
				await joinQueue();
				setPartnerLeft(false);
			} else {
				await leaveRoom();
				toast.success(t("leaveChatRoomSuccess"));
			}
		} catch (error) {
			console.error("Operation failed:", error);
			toast.error(
				error instanceof Error ? error.message : t("somethingWentWrong")
			);
		} finally {
			setIsLoading(false);
		}
	}, [alertDialog, isLoading, joinQueue, leaveRoom, partnerLeft, room, t]);

	useEffect(() => {
		if (!room && !partnerLeft) return;

		const handleKeyDown = (event: KeyboardEvent) => {
			if (!(event.ctrlKey && event.key === "Enter")) return;
			if (event.repeat) return;

			event.preventDefault();
			void handleClick();
		};

		window.addEventListener("keydown", handleKeyDown);
		return () => window.removeEventListener("keydown", handleKeyDown);
	}, [handleClick, partnerLeft, room]);

	if (!room && !partnerLeft) return null;

	const title = partnerLeft
		? t("findNewPartnerShortcut")
		: t("leaveChatRoomShortcut");
	const Icon = partnerLeft ? UserRoundSearch : DoorOpen;

	return (
		<Button
			onClick={handleClick}
			disabled={isLoading}
			size="icon"
			className="shrink-0"
			aria-label={title}
			title={title}
		>
			<Icon aria-hidden="true" />
		</Button>
	);
}
