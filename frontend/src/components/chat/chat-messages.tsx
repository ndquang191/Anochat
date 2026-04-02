"use client";

import { useRef, useEffect } from "react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { ChatMessage } from "./chat-message";

interface Message {
	id: string;
	content: string;
	sender_id: string;
	created_at?: number;
}

interface ChatMessagesProps {
	messages: Message[];
	currentUserId: string;
}

export function ChatMessages({ messages, currentUserId }: ChatMessagesProps) {
	const scrollAreaRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const viewport = scrollAreaRef.current?.querySelector(
			"[data-radix-scroll-area-viewport]"
		);
		if (viewport) {
			viewport.scrollTop = viewport.scrollHeight;
		}
	}, [messages]);

	return (
		<div ref={scrollAreaRef} className="flex-1 min-h-0 overflow-hidden">
			<ScrollArea className="h-full p-2">
				{messages.length === 0 ? (
					<div className="text-center text-muted-foreground py-8">
						<p className="text-sm md:text-base">Chưa có tin nhắn nào</p>
						<p className="text-sm md:text-base mt-2">Hãy bắt đầu cuộc trò chuyện!</p>
					</div>
				) : (
					<div className="flex flex-col gap-3 pb-20">
						{messages.map((message) => (
							<ChatMessage
								key={message.id}
								content={message.content}
								isCurrentUser={message.sender_id === currentUserId}
								created_at={message.created_at}
							/>
						))}
					</div>
				)}
			</ScrollArea>
		</div>
	);
}
