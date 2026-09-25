"use client";

import { useState, useCallback } from "react";
import { queueAPI } from "@/lib/api";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { warmUpAudio } from "@/hooks/use-sound-notification";
import { getStoredPushSubscriptionID, usePWA } from "@/contexts/pwa";
import { useAlertDialogContext } from "@/contexts/alert-dialog";
import { useLanguage } from "@/contexts/theme";

export function useQueue() {
	const [isLoading, setIsLoading] = useState(false);
	const invalidateUserState = useInvalidateUserState();
	const { ensurePushSubscription, isIOS, isStandalone, pushEnabled } = usePWA();
	const alertDialog = useAlertDialogContext();
	const { language, t } = useLanguage();

	const joinQueue = useCallback(async () => {
		warmUpAudio();
		setIsLoading(true);
		try {
			let subscriptionID = getStoredPushSubscriptionID();
			if (pushEnabled && !subscriptionID && "Notification" in window && Notification.permission !== "denied") {
				if (isIOS && !isStandalone) {
					await alertDialog.open({
						title: language === "vi" ? "Cài AnoChat để nhận thông báo" : "Install AnoChat for notifications",
						description: language === "vi" ? "Trên iPhone/iPad, hãy dùng menu Chia sẻ → Thêm vào Màn hình chính. Bạn vẫn có thể tiếp tục tìm mà không bật thông báo." : "On iPhone/iPad, use Share → Add to Home Screen. You can continue searching without notifications.",
						confirmText: t("confirm"),
					});
					const result = await queueAPI.join(null);
					invalidateUserState();
					return result;
				}
				const accepted = await alertDialog.open({
					title: language === "vi" ? "Bật thông báo ghép đôi?" : "Enable match notifications?",
					description: language === "vi" ? "AnoChat sẽ báo khi tìm thấy người trò chuyện nếu bạn chuyển app xuống nền." : "AnoChat can notify you when a partner is found while the app is in the background.",
					confirmText: language === "vi" ? "Bật thông báo" : "Enable",
					cancelText: language === "vi" ? "Để sau" : "Not now",
				});
				if (accepted) subscriptionID = await ensurePushSubscription();
			}
			const result = await queueAPI.join(subscriptionID);
			invalidateUserState();
			return result;
		} finally {
			setIsLoading(false);
		}
	}, [invalidateUserState, ensurePushSubscription, alertDialog, language, isIOS, isStandalone, pushEnabled, t]);

	const leaveQueue = useCallback(async () => {
		setIsLoading(true);
		try {
			const result = await queueAPI.leave();
			invalidateUserState();
			return result;
		} finally {
			setIsLoading(false);
		}
	}, [invalidateUserState]);

	return { joinQueue, leaveQueue, isLoading };
}
