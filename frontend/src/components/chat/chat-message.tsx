"use client";

import { useState } from "react";
import { useLanguage } from "@/contexts/theme";

interface ChatMessageProps {
	content: string;
	isCurrentUser: boolean;
	created_at?: number;
	status?: "pending" | "sent" | "failed";
}

export function ChatMessage({ content, isCurrentUser, created_at, status }: ChatMessageProps) {
	const [showTime, setShowTime] = useState(false);
	const { t } = useLanguage();

	const timeStr = created_at
		? new Date(created_at * 1000).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
		: null;

	return (
		<div className={`flex flex-col gap-0.5 ${isCurrentUser ? "items-end" : "items-start"}`}>
			<div
				onClick={() => timeStr && setShowTime((v) => !v)}
				className={`max-w-[70%] break-words rounded-md px-2.5 py-1.5 text-sm leading-5 ${
					timeStr ? "cursor-pointer" : ""
				} ${
					isCurrentUser
						? "bg-primary text-primary-foreground"
						: "bg-muted text-foreground"
				} ${status === "pending" ? "opacity-70" : ""} ${
					status === "failed" ? "ring-1 ring-destructive" : ""
				}`}
			>
				<p>{content}</p>
			</div>
			{isCurrentUser && status === "pending" && (
				<span className="px-1 text-[10px] text-muted-foreground">
					{t("messageSending")}
				</span>
			)}
			{isCurrentUser && status === "failed" && (
				<span className="px-1 text-[10px] text-destructive">
					{t("messageFailed")}
				</span>
			)}
			{showTime && timeStr && (
				<span className="text-[10px] text-muted-foreground px-1">{timeStr}</span>
			)}
		</div>
	);
}
