"use client";

import Chatbox from "@/components/chat-box";
import { useQueue } from "@/hooks/use-queue";
import { useAuth } from "@/contexts/auth";
import { Loader2 } from "lucide-react";

const Page = () => {
	const { room, loading: authLoading } = useAuth();
	const { isInQueue, queueStatus } = useQueue();

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
