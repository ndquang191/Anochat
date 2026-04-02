"use client";

import { useState } from "react";

interface ChatMessageProps {
	content: string;
	isCurrentUser: boolean;
	created_at?: number;
}

export function ChatMessage({ content, isCurrentUser, created_at }: ChatMessageProps) {
	const [showTime, setShowTime] = useState(false);

	const timeStr = created_at
		? new Date(created_at * 1000).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
		: null;

	return (
		<div className={`flex flex-col gap-0.5 ${isCurrentUser ? "items-end" : "items-start"}`}>
			<div
				onClick={() => timeStr && setShowTime((v) => !v)}
				className={`max-w-[70%] rounded-md px-3 py-2 break-words ${
					timeStr ? "cursor-pointer" : ""
				} ${
					isCurrentUser
						? "bg-primary text-primary-foreground"
						: "bg-muted text-foreground"
				}`}
			>
				<p>{content}</p>
			</div>
			{showTime && timeStr && (
				<span className="text-[10px] text-muted-foreground px-1">{timeStr}</span>
			)}
		</div>
	);
}
