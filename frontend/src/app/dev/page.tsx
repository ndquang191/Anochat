"use client";

import { useAuth } from "@/contexts/auth";
import { Loader2 } from "lucide-react";
import { ChatMessages } from "@/components/chat/chat-messages";
import { ChatEmptyState } from "@/components/chat/chat-empty-state";

export default function DevPage() {
	const { user, room, inQueue, messages } = useAuth();

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

	if (room && user) {
		return (
			<div className="flex flex-col bg-card text-card-foreground h-full relative">
				<ChatMessages messages={messages} currentUserId={user.id} />
				<div className="border-t p-3 text-xs text-muted-foreground text-center">
					Chat input disabled in dev mode
				</div>
			</div>
		);
	}

	return (
		<ChatEmptyState
			title="Chưa có phòng chat nào"
			description='Dùng nút DEV ở góc dưới phải để chuyển trạng thái'
		/>
	);
}
