"use client";

import Chatbox from "@/components/chat-box";
import { useQueue } from "@/hooks/use-queue";
import { Loader2 } from "lucide-react";
import { useEffect } from "react";

const Page = () => {
	const { isInQueue, queueStatus, leaveQueue } = useQueue();

	// Handle page unload and visibility changes to leave queue
	useEffect(() => {
		const handleBeforeUnload = () => {
			if (isInQueue) {
				// Try to leave queue before page unload
				leaveQueue();
			}
		};

		const handleVisibilityChange = () => {
			if (document.hidden && isInQueue) {
				// Leave queue when page becomes hidden (user switches tabs or minimizes)
				leaveQueue();
			}
		};

		// Add event listeners
		window.addEventListener("beforeunload", handleBeforeUnload);
		document.addEventListener("visibilitychange", handleVisibilityChange);

		// Cleanup event listeners
		return () => {
			window.removeEventListener("beforeunload", handleBeforeUnload);
			document.removeEventListener("visibilitychange", handleVisibilityChange);
		};
	}, [isInQueue, leaveQueue]);

	// Show loading indicator when in queue
	if (isInQueue) {
		return (
			<div className="h-full w-full flex items-center justify-center">
				<div className="text-center space-y-4">
					<div className="flex justify-center">
						<Loader2 className="h-12 w-12 animate-spin text-primary" />
					</div>
					<div className="space-y-2">
						<h2 className="text-xl font-semibold">Đang trong hàng chờ...</h2>
						<p className="text-muted-foreground">Vị trí của bạn: {queueStatus?.position || 0}</p>
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
