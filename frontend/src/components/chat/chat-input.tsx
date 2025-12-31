"use client";

import { useState, useCallback, useRef, FormEvent, ChangeEvent } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface ChatInputProps {
	onSendMessage: (message: string) => void;
	onTypingChange: (isTyping: boolean) => void;
	disabled: boolean;
}

const TYPING_TIMEOUT = 2000;

export function ChatInput({ onSendMessage, onTypingChange, disabled }: ChatInputProps) {
	const [message, setMessage] = useState("");
	const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);

	const handleSubmit = useCallback(
		(e: FormEvent) => {
			e.preventDefault();
			if (message.trim() === "" || disabled) return;

			onSendMessage(message);
			setMessage("");
			onTypingChange(false);
		},
		[message, disabled, onSendMessage, onTypingChange]
	);

	const handleChange = useCallback(
		(e: ChangeEvent<HTMLInputElement>) => {
			setMessage(e.target.value);

			if (!disabled && e.target.value.length > 0) {
				onTypingChange(true);

				if (typingTimeoutRef.current) {
					clearTimeout(typingTimeoutRef.current);
				}

				typingTimeoutRef.current = setTimeout(() => {
					onTypingChange(false);
				}, TYPING_TIMEOUT);
			}
		},
		[disabled, onTypingChange]
	);

	return (
		<form
			onSubmit={handleSubmit}
			className="absolute bottom-0 left-0 right-0 flex items-center gap-2 border-t p-4 bg-background"
		>
			<Input
				placeholder="Nhập tin nhắn của bạn..."
				value={message}
				onChange={handleChange}
				className="flex-1"
				disabled={disabled}
			/>
			<Button type="submit" className="shrink-0" disabled={disabled || !message.trim()}>
				Gửi
			</Button>
		</form>
	);
}
