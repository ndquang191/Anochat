"use client";

import { useRef, useEffect } from "react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { ChatMessage } from "./chat-message";

interface Message {
	id: string;
	content: string;
	sender_id: string;
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

	if (messages.length === 0) {
		return (
			<div ref={scrollAreaRef} className="flex-1 h-[calc(100vh-200px)]">
				<ScrollArea className="h-full p-2">
					<div className="text-center text-muted-foreground py-8">
						<p>Chưa có tin nhắn nào</p>
						<p className="text-sm mt-2">Hãy bắt đầu cuộc trò chuyện!</p>
					</div>
				</ScrollArea>
			</div>
		);
	}

	return (
		<div ref={scrollAreaRef} className="flex-1 h-[calc(100vh-200px)]">
			<ScrollArea className="h-full p-2">
				<div className="flex flex-col gap-3">
					{messages.map((message) => (
						<ChatMessage
							key={message.id}
							content={message.content}
							isCurrentUser={message.sender_id === currentUserId}
						/>
					))}
				</div>
			</ScrollArea>
		</div>
	);
}
