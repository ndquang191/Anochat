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
	const messagesEndRef = useRef<HTMLDivElement>(null);

	const scrollToBottom = () => {
		messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
	};

	useEffect(() => {
		scrollToBottom();
	}, [messages]);

	if (messages.length === 0) {
		return (
			<ScrollArea className="flex-1 p-2 h-[calc(100vh-200px)]">
				<div className="text-center text-muted-foreground py-8">
					<p>Chưa có tin nhắn nào</p>
					<p className="text-sm mt-2">Hãy bắt đầu cuộc trò chuyện!</p>
				</div>
			</ScrollArea>
		);
	}

	return (
		<ScrollArea className="flex-1 p-2 h-[calc(100vh-200px)]">
			<div className="flex flex-col gap-3">
				{messages.map((message) => (
					<ChatMessage
						key={message.id}
						content={message.content}
						isCurrentUser={message.sender_id === currentUserId}
					/>
				))}
				<div ref={messagesEndRef} />
			</div>
		</ScrollArea>
	);
}
